package nvd

import (
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tedyst/licenta/db"
)

var deleteCveCmd = &cobra.Command{
	Use:   "deletecve",
	Short: "Delete CVE from database",
	Long: `This command allows you to delete a CVE from the database.

This is needed since some vulnerabilities do not affect the actual product, but are still present in the NVD database. 
For example: CVE-2009-2943, CVE-2010-3781.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		database := db.InitDatabase(viper.GetString("database"))

		slog.InfoContext(cmd.Context(), "Deleting CVE", "cve_id", args[0])

		err := database.DeleteNvdCveByName(cmd.Context(), args[0])
		if err != nil {
			return err
		}

		slog.InfoContext(cmd.Context(), "Deleted CVE", "cve_id", args[0])

		return nil
	},
}

func init() {
	nvdCmd.AddCommand(deleteCveCmd)
}
