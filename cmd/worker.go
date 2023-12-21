package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tedyst/licenta/worker"
)

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Run the worker using the production configuration",
	Long:  `This command connects to the API server and listens to tasks.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return worker.ReceiveTasks(cmd.Context(), viper.GetString("api"), viper.GetString("worker-token"))
	},
}

func init() {
	workerCmd.Flags().String("api", "http://localhost:5000", "API Server URL")
	workerCmd.Flags().String("worker-token", "", "Worker token")
	workerCmd.MarkFlagRequired("worker-token")

	rootCmd.AddCommand(workerCmd)
}
