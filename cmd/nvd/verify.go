package nvd

import (
	"log/slog"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/nvd"
)

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify product vulnerabilities",
	Long:  `This command allows you to check a specific version of a product against the known vulnerabilities from the database.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		var product nvd.Product
		switch args[0] {
		case "postgres":
			product = nvd.POSTGRESQL
		default:
			return errors.New("product does not exist")
		}

		database := db.InitDatabase(viper.GetString("database"))

		slog.InfoContext(cmd.Context(), "Verifying product for vulnerabilities", "product", product, "version", args[1])

		cves, err := database.GetCvesByProductAndVersion(cmd.Context(), queries.GetCvesByProductAndVersionParams{
			DatabaseType: int32(product),
			Version:      args[1],
		})
		if err != nil {
			return err
		}

		if len(cves) == 0 {
			slog.InfoContext(cmd.Context(), "No vulnerabilities found", "product", product, "version", args[1])
		}

		for _, cve := range cves {
			slog.WarnContext(cmd.Context(), "Vulnerability found", "cve_id", cve.NvdCfe.CveID, "detail", cve.NvdCfe.Description, "score", cve.NvdCfe.Score)
		}

		return nil
	},
}

func init() {
	nvdCmd.AddCommand(verifyCmd)
}
