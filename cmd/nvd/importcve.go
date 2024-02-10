package nvd

import (
	"fmt"
	"os"

	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/nvd"
)

var importCveCmd = &cobra.Command{
	Use:   "importcve",
	Short: "Import CVE from NVD API or file",
	Long:  `This command allows you to import CVE from NVD API or file into the database`,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		var product nvd.Product

		switch viper.GetString("product") {
		case "postgresql":
			product = nvd.POSTGRESQL
		default:
			return errors.New("invalid product")
		}

		database, err := db.InitDatabase(viper.GetString("database")).StartTransaction(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to start transaction: %w", err)
		}
		defer func() {
			err = errors.Join(database.EndTransaction(cmd.Context(), err != nil))
		}()

		cpe, err := database.GetCPEByProductAndVersion(cmd.Context(), queries.GetCPEByProductAndVersionParams{
			DatabaseType: int32(product),
			Version:      viper.GetString("version"),
		})
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("failed to get cpe: %w", err)
		}
		if cpe == nil {
			return errors.New("cpe not found")
		}

		if viper.GetString("file") == "" {
			fmt.Println("not implemented")
			return nil
		}

		reader, err := os.Open(viper.GetString("file"))
		if err != nil {
			return fmt.Errorf("failed to open file: %w", err)
		}
		defer func() {
			err = errors.Join(err, reader.Close())
		}()

		result, err := nvd.ParseCveAPI(cmd.Context(), reader)
		if err != nil {
			return fmt.Errorf("failed to parse cve api: %w", err)
		}

		for _, result := range result.Vulnerabilities {
			var cve *queries.NvdCfe
			cve, err = database.GetCveByCveID(cmd.Context(), result.Cve.ID)
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("failed to get cve: %w", err)
			}
			if errors.Is(err, pgx.ErrNoRows) {
				publishedDate, err := result.Cve.PubslihedDate()
				if err != nil {
					return fmt.Errorf("failed to parse published date: %w", err)
				}
				lastModified, err := result.Cve.LastModifiedDate()
				if err != nil {
					return fmt.Errorf("failed to parse last modified date: %w", err)
				}
				score, err := result.Cve.Score()
				if err != nil {
					return fmt.Errorf("failed to get score: %w", err)
				}

				cve, err = database.CreateNvdCve(cmd.Context(), queries.CreateNvdCveParams{
					CveID:        result.Cve.ID,
					Description:  result.Cve.Descriptions[0].Value,
					Published:    pgtype.Timestamptz{Time: publishedDate, Valid: true},
					LastModified: pgtype.Timestamptz{Time: lastModified, Valid: true},
					Score:        score,
				})
				if err != nil {
					return fmt.Errorf("failed to create cve: %w", err)
				}
			}

			_, err := database.GetCveCpeByCveAndCpe(cmd.Context(), queries.GetCveCpeByCveAndCpeParams{
				CveID: cve.ID,
				CpeID: cpe.ID,
			})
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("failed to get cvecpe: %w", err)
			}
			if errors.Is(err, pgx.ErrNoRows) {
				_, err = database.CreateNvdCveCPE(cmd.Context(), queries.CreateNvdCveCPEParams{
					CveID: cve.ID,
					CpeID: cpe.ID,
				})
				if err != nil {
					return fmt.Errorf("failed to create cvecpe: %w", err)
				}
			}
		}

		return nil
	},
}

func init() {
	importCveCmd.Flags().String("file", "", "Load from file instead from API")
	importCveCmd.Flags().String("product", "", "Product to import for: postgresql/mysql/redis")
	importCveCmd.Flags().String("version", "", "Version to import for: 9.6.0/5.7.0/3.2.0")

	if err := importCveCmd.MarkFlagRequired("product"); err != nil {
		panic(err)
	}
	if err := importCveCmd.MarkFlagRequired("version"); err != nil {
		panic(err)
	}

	nvdCmd.AddCommand(importCveCmd)
}
