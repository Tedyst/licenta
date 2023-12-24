package extract

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tedyst/licenta/extractors/file"
)

var extractFileCmd = &cobra.Command{
	Use:   "file [filename]",
	Short: "Runs the file extractor",
	Long:  `This command allows you to run the file extractor for the provided file. The file extractor will find all the passwords and usernames from a file and show them to you. It does not require a database running.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f, err := os.OpenFile(args[0], os.O_RDONLY, 0)
		if err != nil {
			return err
		}
		fileScanner, err := file.NewScanner()
		if err != nil {
			return err
		}

		results, err := fileScanner.ExtractFromReader(context.Background(), args[0], f)
		if err != nil {
			return err
		}
		for _, result := range results {
			fmt.Printf("%s\n", result.String())
		}

		return nil
	},
}

func init() {
	extractCmd.AddCommand(extractFileCmd)
}
