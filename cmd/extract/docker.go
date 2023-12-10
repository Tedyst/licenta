package extract

import (
	"context"
	"fmt"
	"os"

	"log/slog"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/daemon"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tedyst/licenta/extractors/docker"
)

var extractDockerCmd = &cobra.Command{
	Use:   "docker [filename]",
	Short: "Run the docker extractor",
	Long:  `This command scans all the layers from a docker image and extracts the usernames and passwords from each layer. It does not require a database running. It can use the local Docker daemon to load images. If Docker daemon is not available, it will use the remote registry. The results will be printed to stdout.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var programLevel = new(slog.LevelVar)
		h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: programLevel})
		slog.SetDefault(slog.New(h))
		programLevel.Set(slog.LevelDebug)

		callbackFunc := func(scanner *docker.DockerScan, result *docker.LayerResult) error {
			return nil
		}
		ctx := context.Background()
		scanner, err := docker.NewScanner(ctx, args[0], docker.WithCallbackResult(callbackFunc))
		if err != nil {
			fmt.Printf("%+v\n", err)
		}
		var layers []v1.Layer
		if viper.GetBool("local") {
			ref, err := name.ParseReference(args[0])
			if err != nil {
				fmt.Printf("%+v\n", err)
			}
			asd, err := daemon.Image(ref)
			if err != nil {
				fmt.Printf("%+v\n", err)
			}
			layers, err = asd.Layers()
			if err != nil {
				fmt.Printf("%+v\n", err)
			}
		} else {
			layers, err = scanner.FindLayers(ctx)
			if err != nil {
				fmt.Printf("%+v\n", err)
			}
		}

		// ctx, cancelCtx := context.WithTimeout(ctx, time.Second*10)
		// defer cancelCtx()
		digests := ""
		for _, layer := range layers {
			asd, err := layer.Digest()
			if err != nil {
				fmt.Printf("%+v\n", err)
			}
			digests += asd.String() + "\n"
		}

		err = scanner.ProcessLayers(ctx, layers)
		if err != nil {
			fmt.Printf("%+v\n", err)
		}
		fmt.Printf("done")
	},
}

func init() {
	extractDockerCmd.Flags().Bool("local", false, "Use local Docker daemon for loading images")

	extractCmd.AddCommand(extractDockerCmd)
}
