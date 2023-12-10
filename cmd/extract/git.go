package extract

import (
	"log/slog"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		var repo *gitgo.Repository
		var err error
		var scanner *git.GitScan

		if strings.HasPrefix(args[0], "https://") || strings.HasPrefix(args[0], "http://") || strings.HasPrefix(args[0], "git://") || strings.HasPrefix(args[0], "ssh://") {
			slog.InfoContext(cmd.Context(), "Opening remote git repo", "url", args[0])
			scanner, err = git.New(args[0])
			if err != nil {
				return err
			}
		}
		if strings.HasPrefix(args[0], "/") || strings.HasPrefix(args[0], ".") {
			slog.InfoContext(cmd.Context(), "Opening local git repo", "path", args[0])
			repo, err = gitgo.PlainOpen(args[0])
			if err != nil {
				return err
			}
			scanner, err = git.NewFromRepo(repo)
			if err != nil {
				return err
			}
		}
		if repo == nil {
			return err
		}
		slog.InfoContext(cmd.Context(), "Opened git repo")

		err = scanner.Scan(cmd.Context())
		if err != nil {
			return err
		}
		slog.InfoContext(cmd.Context(), "Finished git scan")

		return nil
	},
}

func init() {
	extractCmd.AddCommand(extractGitCmd)
}
