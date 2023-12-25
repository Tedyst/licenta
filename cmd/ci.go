package cmd

import (
	"log/slog"
	"os"

	"github.com/deepmap/oapi-codegen/v2/pkg/securityprovider"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tedyst/licenta/api/v1/generated"
	"github.com/tedyst/licenta/ci"
)

var ciCmd = &cobra.Command{
	Use:   "ci",
	Short: "Signal the Server that a build should be started and wait for it to finish",
	Long:  `This command connects to the API server and listens to tasks.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		apiKeyProvider, err := securityprovider.NewSecurityProviderApiKey("header", "X-Worker-Token", viper.GetString("worker-token"))
		if err != nil {
			return errors.Wrap(err, "error creating security provider")
		}
		client, err := generated.NewClientWithResponses(viper.GetString("api")+"/api/v1", generated.WithRequestEditorFn(apiKeyProvider.Intercept))
		if err != nil {
			return errors.Wrap(err, "error creating client")
		}

		severity, err := ci.ProjectRunAndWaitResults(cmd.Context(), client, viper.GetInt("project"))
		if err != nil {
			return errors.Wrap(err, "error running project")
		}

		if severity > viper.GetInt("severity") {
			slog.Error("Severity is higher than allowed, failing build", "severity", severity, "allowed", viper.GetInt("severity"))
			os.Exit(1)
		}

		return nil
	},
}

func init() {
	ciCmd.Flags().String("api", "http://localhost:5000", "API Server URL")
	ciCmd.Flags().Int("project", 0, "The project ID to scan")
	if err := ciCmd.MarkFlagRequired("project"); err != nil {
		panic(err)
	}
	ciCmd.Flags().String("worker-token", "", "Worker token")
	if err := ciCmd.MarkFlagRequired("worker-token"); err != nil {
		panic(err)
	}
	ciCmd.Flags().Int("severity", 2, "Minimum severity to fail the build. 0 - low, 1 - medium, 2 - high")

	rootCmd.AddCommand(ciCmd)
}
