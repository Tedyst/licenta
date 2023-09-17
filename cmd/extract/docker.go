package extract

import (
	"context"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/spf13/cobra"
	"github.com/tedyst/licenta/extractors/docker"
)

var extractDockerCmd = &cobra.Command{
	Use:   "docker [filename]",
	Short: "Run the extractor tool for the provided file",
	Long:  `Run the extractor tool for the provided file`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		callbackFunc := func(result docker.LayerResult) {
			for _, res := range result.Results {
				println(res.FileName)
			}
		}
		docker.ProcessImage(context.Background(), args[0], callbackFunc, docker.WithCredentials(&authn.Basic{
			Username: "Tedyst",
			Password: "",
		}))
	},
}

func init() {
	extractCmd.AddCommand(extractDockerCmd)
}
