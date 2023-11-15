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
		var scanner *git.GitScan

		if strings.HasPrefix(args[0], "https://") || strings.HasPrefix(args[0], "http://") || strings.HasPrefix(args[0], "git://") || strings.HasPrefix(args[0], "ssh://") {
			scanner, err = git.New(args[0])
			if err != nil {
				log.Fatal(err)
			}
		}
		if strings.HasPrefix(args[0], "/") || strings.HasPrefix(args[0], ".") {
			repo, err = gitgo.PlainOpen(args[0])
			if err != nil {
				log.Fatal(err)
			}
			scanner, err = git.NewFromRepo(repo)
			if err != nil {
				log.Fatal(err)
			}
		}
		if repo == nil {
			log.Fatal("Invalid git repo")
		}
		fmt.Println("Extracting from git repo...")

		err = scanner.Scan(context.Background())
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Done")
	},
}

func init() {
	extractCmd.AddCommand(extractGitCmd)
}
