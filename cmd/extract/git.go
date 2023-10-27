package extract

import (
	"context"
	"fmt"
	"log"
	"strings"

	gitgo "github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
	"github.com/tedyst/licenta/extractors/git"
)

var extractGitCmd = &cobra.Command{
	Use:   "git [filename]",
	Short: "Run the extractor tool for the provided git repo",
	Long:  `Run the extractor tool for the provided git repo`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var repo *gitgo.Repository
		var err error

		if strings.HasPrefix(args[0], "https://") || strings.HasPrefix(args[0], "http://") || strings.HasPrefix(args[0], "git://") || strings.HasPrefix(args[0], "ssh://") {
			repo, err = git.PullGitRepository(context.Background(), args[0], 0, nil)
			if err != nil {
				log.Fatal(err)
			}
		}
		if strings.HasPrefix(args[0], "/") || strings.HasPrefix(args[0], ".") {
			repo, err = gitgo.PlainOpen(args[0])
			if err != nil {
				log.Fatal(err)
			}
		}
		if repo == nil {
			log.Fatal("Invalid git repo")
		}
		fmt.Println("Extracting from git repo...")

		err = git.ExtractGit(context.Background(), repo)
		if err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	extractCmd.AddCommand(extractGitCmd)
}
