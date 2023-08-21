package cmd

import (
	"net/http"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tedyst/licenta/api"
	"github.com/tedyst/licenta/api/v1/middleware/session"
	database "github.com/tedyst/licenta/db"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the server",
	Long:  `Run the server.`,
	Run: func(cmd *cobra.Command, args []string) {
		db := database.InitDatabase()
		sessionStore := session.New(db, viper.GetBool("debug"))
		app := api.Initialize(db, sessionStore, api.ApiConfig{
			Debug:  false,
			Origin: viper.GetString("baseurl"),
		})

		print("Listening on port " + viper.GetString("port") + "\n")
		err := http.ListenAndServe(":"+viper.GetString("port"), app)
		if err != nil {
			panic(err)
		}
	},
}

func init() {
	serveCmd.Flags().String("sendgrid", "", "Sendgrid API Key")
	serveCmd.Flags().String("postmark", "", "Postmark API Key")

	serveCmd.Flags().String("email.sender", "no-reply@tedyst.ro", "Email sender")
	serveCmd.Flags().String("email.senderName", "Licenta", "Email sender name")

	serveCmd.Flags().String("baseurl", "http://localhost:8080", "Base URL")

	serveCmd.Flags().Bool("telemetry.metrics.enabled", false, "Enable metrics")
	serveCmd.Flags().Bool("telemetry.tracing.enabled", false, "Enable tracing")
	serveCmd.Flags().String("telemetry.tracing.jaeger", "", "Jaeger URL")

	serveCmd.Flags().Int16P("port", "p", 5000, "Port to listen on")

	serveCmd.Flags().String("database", "", "Database connection string")

	rootCmd.AddCommand(serveCmd)
}
