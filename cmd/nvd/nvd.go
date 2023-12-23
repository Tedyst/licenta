package nvd

import (
	"github.com/spf13/cobra"
)

var nvdCmd = &cobra.Command{
	Use:   "nvd",
	Short: "NVD Vulnerabilities management",
	Long:  `This command allows you to manage NVD vulnerabilities stored in the database.`,
}

func NewNvdCmd() *cobra.Command {
	return nvdCmd
}

func init() {
	nvdCmd.PersistentFlags().String("database", "", "Database connection string")

	if err := nvdCmd.MarkFlagRequired("database"); err != nil {
		panic(err)
	}
}
