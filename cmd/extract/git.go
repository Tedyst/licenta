package extract

import (
	"log/slog"
	"strings"

	gitgo "github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
	"github.com/tedyst/licenta/extractors/file"
	"github.com/tedyst/licenta/extractors/git"
)

var extractGitCmd = &cobra.Command{
	Use:   "git [filename]",
	Short: "Runs the docker extractor",
	Long: `This command allows you to run the git extractor for the provided git repository. The git extractor will find all the passwords and usernames from a git repository and show them to you. It does not require a database running. The outputs are printed to stdout. The git repository can be local or remote.

	If it is remote, it must be accessible without authentication.
	If it is local, it must be a valid git repository.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var repo *gitgo.Repository
		var err error
		var scanner *git.GitScan

		fileScanner, err := file.NewScanner()

		if strings.HasPrefix(args[0], "https://") || strings.HasPrefix(args[0], "http://") || strings.HasPrefix(args[0], "git://") || strings.HasPrefix(args[0], "ssh://") {
			slog.InfoContext(cmd.Context(), "Opening remote git repo", "url", args[0])
			scanner, err = git.New(args[0], fileScanner)
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
			scanner, err = git.NewFromRepo(repo, fileScanner)
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
