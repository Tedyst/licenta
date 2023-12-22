package cmd

import (
	"net/http"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tedyst/licenta/api"
	"github.com/tedyst/licenta/api/v1/middleware/session"
	"github.com/tedyst/licenta/bruteforce"
	database "github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/email"
	localExchange "github.com/tedyst/licenta/messages/local"
	"github.com/tedyst/licenta/tasks/local"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the server using the production configuration",
	Long:  `This command starts the API server and waits for requests. By default, it uses the local runner for async tasks.`,
	Run: func(cmd *cobra.Command, args []string) {
		viper.SetConfigFile("config.yaml")
		viper.ReadInConfig()

		db := database.InitDatabase(viper.GetString("database"))
		sessionStore := session.New(db, viper.GetBool("debug"))

		localExchange := localExchange.NewLocalExchange()
		bruteforceProvider := bruteforce.NewDatabaseBruteforceProvider(db)

		taskRunner := local.NewLocalRunner(true, email.NewConsoleEmailSender(
			viper.GetString("email.senderName"),
			viper.GetString("email.sender"),
		), db, localExchange, bruteforceProvider)

		app := api.Initialize(db, sessionStore, api.ApiConfig{
			Debug:      viper.GetBool("debug"),
			Origin:     viper.GetString("baseurl"),
			TaskRunner: taskRunner,
		}, localExchange)

		print("Listening on port " + viper.GetString("port") + "\n")
		err := http.ListenAndServe(":"+viper.GetString("port"), app)
		if err != nil {
			panic(err)
		}
	},
}

func init() {
	serveCmd.Flags().String("baseurl", "http://localhost:8080", "Base URL")

	serveCmd.Flags().Int16P("port", "p", 5000, "Port to listen on")

	serveCmd.Flags().String("database", "", "Database connection string")

	rootCmd.AddCommand(serveCmd)
}
