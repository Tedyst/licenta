package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tedyst/licenta/api"
	database "github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/telemetry"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the server",
	Long:  `Run the server.`,
	Run: func(cmd *cobra.Command, args []string) {
		telemetry.InitTelemetry()
		database.InitDatabase()
		app := api.InitializeFiber()
		app.Listen(":" + viper.GetString("port"))
	},
}

func init() {
	serveCmd.Flags().String("sendgrid", "", "Sendgrid API Key")
	serveCmd.Flags().String("postmark", "", "Postmark API Key")

	serveCmd.Flags().String("email.sender", "no-reply@tedyst.ro", "Email sender")
	serveCmd.Flags().String("email.senderName", "Licenta", "Email sender name")

	serveCmd.Flags().String("baseurl", "http://localhost:8080", "Base URL")

	serveCmd.Flags().Bool("telemetry.metrics", false, "Enable metrics")
	serveCmd.Flags().Bool("telemetry.tracing", false, "Enable tracing")

	serveCmd.Flags().Int16P("port", "p", 5000, "Port to listen on")

	rootCmd.AddCommand(serveCmd)
}
