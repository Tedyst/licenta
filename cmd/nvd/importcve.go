package nvd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tedyst/licenta/nvd"
)

var importCveCmd = &cobra.Command{
	Use:   "importcve",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if viper.GetString("file") == "" {
			fmt.Println("not implemented")
			return nil
		}

		reader, err := os.Open(viper.GetString("file"))
		if err != nil {
			return err
		}
		defer reader.Close()

		result, err := nvd.ParseCveAPI(cmd.Context(), reader)
		if err != nil {
			return err
		}

		for _, result := range result.Vulnerabilities {
			asd, err := nvd.GetCveScore(cmd.Context(), result.Cve)
			fmt.Println(result.Cve.ID, asd, err)
		}

		println("done")

		return nil
	},
}

func init() {
	importCveCmd.Flags().String("file", "", "Load from file instead from API")
	importCveCmd.Flags().String("product", "", "Product to import for: postgresql/mysql/redis")
	importCveCmd.Flags().String("version", "", "Version to import for: 9.6.0/5.7.0/3.2.0")

	nvdCmd.AddCommand(importCveCmd)
}
