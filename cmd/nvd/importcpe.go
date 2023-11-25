package nvd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
		if viper.GetString("file") == "" {
			fmt.Println("not implemented")
			return nil
		}

		reader, err := os.Open(viper.GetString("file"))
		if err != nil {
			return err
		}
		defer reader.Close()

		result, err := nvd.ParseCpeAPI(cmd.Context(), reader)
		if err != nil {
			return err
		}

		for _, result := range result.Products {
			fmt.Println(result.Cpe)
			fmt.Println(nvd.ExtractCpeVersionProduct(nvd.POSTGRESQL, result.Cpe.Titles))
		}

		println("done")

		return nil
	},
}

func init() {
	importCpeCmd.Flags().String("file", "", "Load from file instead from API")

	nvdCmd.AddCommand(importCpeCmd)
}
