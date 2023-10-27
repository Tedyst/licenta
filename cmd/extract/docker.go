package extract

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tedyst/licenta/extractors/docker"
	"golang.org/x/exp/slog"
)

var extractDockerCmd = &cobra.Command{
	Use:   "docker [filename]",
	Short: "Run the extractor tool for the provided file",
	Long:  `Run the extractor tool for the provided file`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var programLevel = new(slog.LevelVar)
		h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: programLevel})
		slog.SetDefault(slog.New(h))
		programLevel.Set(slog.LevelDebug)

		callbackFunc := func(result docker.LayerResult) {
			for _, res := range result.Results {
				fmt.Printf("%s:%s:%s\n", result.Layer, result.FileName, res.String())
			}
		}
		err := docker.ProcessImage(context.Background(), args[0], callbackFunc)
		if err != nil {
			fmt.Printf("%+v\n", err)
		}
		fmt.Printf("done")
	},
}

func init() {
	extractCmd.AddCommand(extractDockerCmd)
}
