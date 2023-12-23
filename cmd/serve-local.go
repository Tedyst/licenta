package cmd

import (
	"log/slog"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tedyst/licenta/api"
	"github.com/tedyst/licenta/api/v1/middleware/session"
	"github.com/tedyst/licenta/bruteforce"
	database "github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/email"
	localExchange "github.com/tedyst/licenta/messages/local"
	"github.com/tedyst/licenta/tasks"
	"github.com/tedyst/licenta/tasks/local"
)

var serveLocalCmd = &cobra.Command{
	Use:   "servelocal",
	Short: "Run the server using the development configuration",
	Long:  `This command starts the API server and waits for requests. It uses the local runner for async tasks, to allow easier debugging. It also uses the console email sender, to allow easier debugging.`,
	Run: func(cmd *cobra.Command, args []string) {
		db := database.InitDatabase(viper.GetString("database"))
		sessionStore := session.New(db, viper.GetBool("debug"))

		localExchange := localExchange.NewLocalExchange()
		brutefroceProvider := bruteforce.NewDatabaseBruteforceProvider(db)

		var taskRunner tasks.TaskRunner
		if viper.GetString("email.sendgrid") != "" {
			taskRunner = local.NewLocalRunner(viper.GetBool("debug"), email.NewConsoleEmailSender(
				viper.GetString("email.senderName"),
				viper.GetString("email.sender"),
			), db, localExchange, brutefroceProvider)
		} else {
			taskRunner = local.NewLocalRunner(viper.GetBool("debug"), email.NewSendGridEmailSender(
				viper.GetString("email.sendgrid"),
				viper.GetString("email.senderName"),
				viper.GetString("email.sender"),
			), db, localExchange, brutefroceProvider)
		}

		app := api.Initialize(db, sessionStore, api.ApiConfig{
			Debug:      viper.GetBool("debug"),
			Origin:     viper.GetString("baseurl"),
			TaskRunner: taskRunner,
		}, localExchange)

		slog.Info("Started web server", "port", viper.GetString("port"), "baseurl", viper.GetString("baseurl"))
		err := http.ListenAndServe(":"+viper.GetString("port"), app)
		if err != nil {
			panic(err)
		}
	},
}

func init() {
	serveLocalCmd.Flags().String("email.sendgrid", "", "Sendgrid API Key")

	serveLocalCmd.Flags().String("email.sender", "no-reply@tedyst.ro", "Email sender")
	serveLocalCmd.Flags().String("email.senderName", "Licenta", "Email sender name")

	serveLocalCmd.Flags().String("baseurl", "http://localhost:8080", "Base URL")

	// serveCmd.Flags().Bool("telemetry.metrics.enabled", false, "Enable metrics")
	// serveCmd.Flags().Bool("telemetry.tracing.enabled", false, "Enable tracing")
	// serveCmd.Flags().String("telemetry.tracing.jaeger", "", "Jaeger URL")

	serveLocalCmd.Flags().Int16P("port", "p", 5000, "Port to listen on")

	serveLocalCmd.Flags().String("database", "", "Database connection string")

	rootCmd.AddCommand(serveLocalCmd)
}
