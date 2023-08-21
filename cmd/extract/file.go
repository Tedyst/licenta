package extract

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tedyst/licenta/extractors/file"
)

var extractFileCmd = &cobra.Command{
	Use:   "file [filename]",
	Short: "Run the extractor tool for the provided file",
	Long:  `Run the extractor tool for the provided file`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		f, err := os.OpenFile(args[0], os.O_RDONLY, 0)
		if err != nil {
			panic(err)
		}
		results, err := file.ExtractFromReader(f)
		if err != nil {
			panic(err)
		}
		for _, result := range results {
			fmt.Printf("ExtractResult{Name: %s, Line: %s, LineNumber: %d, Match: %s, Severity: %d, Probability: %f, Username: %s, Password: %s}\n", result.Name, result.Line, result.LineNumber, result.Match, result.Severity, result.Probability, result.Username, result.Password)
		}
	},
}

func init() {
	extractCmd.AddCommand(extractFileCmd)
}
