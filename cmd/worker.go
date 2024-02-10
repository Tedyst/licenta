package cmd

import (
	"fmt"

	"github.com/deepmap/oapi-codegen/v2/pkg/securityprovider"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tedyst/licenta/api/v1/generated"
	"github.com/tedyst/licenta/worker"
)

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Run the worker using the production configuration",
	Long:  `This command connects to the API server and listens to tasks.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		apiKeyProvider, err := securityprovider.NewSecurityProviderApiKey("header", "X-Worker-Token", viper.GetString("worker-token"))
		if err != nil {
			return fmt.Errorf("error creating security provider: %w", err)
		}

		httpClient, err := initHttpCsrfClient()
		if err != nil {
			return fmt.Errorf("error creating http client: %w", err)
		}

		client, err := generated.NewClientWithResponses(viper.GetString("api")+"/api/v1", generated.WithRequestEditorFn(apiKeyProvider.Intercept), generated.WithHTTPClient(httpClient))
		if err != nil {
			return fmt.Errorf("error creating client: %w", err)
		}

		return worker.ReceiveTasks(cmd.Context(), client)
	},
}

func init() {
	workerCmd.Flags().String("api", "http://localhost:5000", "API Server URL")
	workerCmd.Flags().String("worker-token", "", "Worker token")
	if err := workerCmd.MarkFlagRequired("worker-token"); err != nil {
		panic(err)
	}

	rootCmd.AddCommand(workerCmd)
}
