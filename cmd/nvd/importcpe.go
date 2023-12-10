package nvd

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/models"
	"github.com/tedyst/licenta/nvd"
)

var importCpeCmd = &cobra.Command{
	Use:   "importcpe",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var reader io.ReadCloser
		var err error
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
			defer reader.Close()

			fmt.Println("Downloaded CPE from NVD API")
		} else {
			reader, err = os.Open(viper.GetString("file"))
			if err != nil {
				return err
			}
			defer reader.Close()
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
		defer database.EndTransaction(cmd.Context(), err)

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

			var cpe *models.NvdCPE
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
	importCpeCmd.Flags().String("database", "", "Database connection string")

	importCpeCmd.MarkFlagRequired("database")
	importCpeCmd.MarkFlagRequired("product")

	nvdCmd.AddCommand(importCpeCmd)
}
