package extract

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tedyst/licenta/extractors/git"
)

var extractGitCmd = &cobra.Command{
	Use:   "git [filename]",
	Short: "Run the extractor tool for the provided git repo",
	Long:  `Run the extractor tool for the provided git repo`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := git.ExtractGit(context.Background(), args[0])
		if err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	extractCmd.AddCommand(extractGitCmd)
}
