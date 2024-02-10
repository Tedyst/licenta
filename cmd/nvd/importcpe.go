package nvd

import (
	"database/sql"
	errorss "errors"
	"fmt"
	"io"
	"os"
	"time"

	"errors"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/nvd"
)

var importCpeCmd = &cobra.Command{
	Use:   "importcpe",
	Short: "Import CPE from NVD API or file",
	Long:  `This command allows you to import CPE from NVD API or file into the database.`,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		var reader io.ReadCloser
		var product nvd.Product

		switch viper.GetString("product") {
		case "postgresql":
			product = nvd.POSTGRESQL
		default:
			return errors.New("invalid product")
		}

		if viper.GetString("file") == "" {
			reader, err = nvd.DownloadCpe(cmd.Context(), product)
			if err != nil {
				return err
			}
			defer func() {
				err = errorss.Join(err, reader.Close())
			}()

			fmt.Println("Downloaded CPE from NVD API")
		} else {
			reader, err = os.Open(viper.GetString("file"))
			if err != nil {
				return err
			}
			defer func() {
				err = errorss.Join(err, reader.Close())
			}()
		}

		result, err := nvd.ParseCpeAPI(cmd.Context(), reader)
		if err != nil {
			return err
		}

		fmt.Println("Parsed CPE from NVD API")

		database, err := db.InitDatabase(viper.GetString("database")).StartTransaction(cmd.Context())
		if err != nil {
			return err
		}
		defer func() {
			err = errorss.Join(err, database.EndTransaction(cmd.Context(), err))
		}()

		dbCpes, err := database.GetNvdCPEsByDBType(cmd.Context(), int32(nvd.POSTGRESQL))
		if err != nil {
			return err
		}

		for _, result := range result.Products {
			fmt.Println("Trying to import", result.Cpe.CpeName, "...")
			version, err := nvd.ExtractCpeVersionProduct(nvd.POSTGRESQL, result.Cpe.Titles)
			if err != nil {
				continue
			}

			t, err := time.Parse("2006-01-02T15:04:05.000", result.Cpe.LastModified)
			if err != nil {
				return err
			}

			var cpe *queries.NvdCpe
			found := false
			for _, dbCpe := range dbCpes {
				if dbCpe.Cpe == result.Cpe.CpeName {
					cpe = dbCpe
					found = true
					break
				}
			}

			if !found {
				cpe, err = database.CreateNvdCPE(cmd.Context(), queries.CreateNvdCPEParams{
					Cpe:          result.Cpe.CpeName,
					DatabaseType: int32(nvd.POSTGRESQL),
					LastModified: pgtype.Timestamptz{Time: t, Valid: true},
					Version:      version,
				})
				if err != nil {
					return err
				}
			}

			if !cpe.LastModified.Time.Equal(t) {
				err = database.UpdateNvdCPE(cmd.Context(), queries.UpdateNvdCPEParams{
					ID:           cpe.ID,
					LastModified: pgtype.Timestamptz{Time: t, Valid: true},
					Version:      sql.NullString{String: version, Valid: true},
				})
				if err != nil {
					return err
				}
			}
		}

		return nil
	},
}

func init() {
	importCpeCmd.Flags().String("file", "", "Load from file instead from API")
	importCpeCmd.Flags().String("product", "", "Product to import for: postgresql/mysql/redis")
	err := importCpeCmd.MarkFlagRequired("product")
	if err != nil {
		panic(err)
	}

	nvdCmd.AddCommand(importCpeCmd)
}
