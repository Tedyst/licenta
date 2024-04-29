package cmd

import (
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tedyst/licenta/api"
	"github.com/tedyst/licenta/api/auth"
	"github.com/tedyst/licenta/api/auth/workerauth"
	"github.com/tedyst/licenta/api/authorization"
	"github.com/tedyst/licenta/bruteforce"
	"github.com/tedyst/licenta/cache"
	database "github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/email"
	localExchange "github.com/tedyst/licenta/messages/local"
	"github.com/tedyst/licenta/tasks/local"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the server using the prod configuration",
	Long:  `This command starts the API server and waits for requests. It uses all of the local implementations of the dependencies.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		db := database.InitDatabase(viper.GetString("database"))

		hashKey, err := hex.DecodeString(viper.GetString("hash-key"))
		if err != nil {
			return fmt.Errorf("hash key must be a hex string: %w", err)
		}
		if len(hashKey) != 32 {
			return fmt.Errorf("hash key must be 32 bytes long (64 hex characters)")
		}

		encryptKey, err := hex.DecodeString(viper.GetString("encrypt-key"))
		if err != nil {
			return fmt.Errorf("encrypt key must be a hex string: %w", err)
		}
		if len(encryptKey) != 32 {
			return fmt.Errorf("encrypt key must be 32 bytes long (64 hex characters)")
		}

		localExchange := localExchange.NewLocalExchange()
		bruteforceProvider := bruteforce.NewDatabaseBruteforceProvider(db)

		emailSender := email.NewSendGridEmailSender(
			viper.GetString("email.sendgrid"),
			viper.GetString("email.senderName"),
			viper.GetString("email.sender"),
		)

		taskRunner := local.NewLocalRunner(
			viper.GetBool("debug"),
			emailSender,
			db,
			localExchange,
			bruteforceProvider,
		)

		userCacheProvider, err := cache.NewLocalCacheProvider[queries.User]()
		if err != nil {
			return err
		}

		userAuth, err := auth.NewAuthenticationProvider(viper.GetString("baseurl"), db, hashKey, encryptKey, taskRunner, userCacheProvider)
		if err != nil {
			return fmt.Errorf("failed to initialize user authentication: %w", err)
		}

		waCacheProvider, err := cache.NewLocalCacheProvider[queries.Worker]()
		if err != nil {
			return err
		}

		workerAuth := workerauth.NewWorkerAuth(waCacheProvider, db)

		authorizationCache, err := cache.NewLocalCacheProvider[int16]()
		if err != nil {
			return err
		}
		authorizationManager := authorization.NewAuthorizationManager(db, authorizationCache)

		app, err := api.Initialize(api.ApiConfig{
			Debug:                viper.GetBool("debug"),
			Origin:               viper.GetString("baseurl"),
			TaskRunner:           taskRunner,
			MessageExchange:      localExchange,
			WorkerAuth:           workerAuth,
			UserAuth:             userAuth,
			Database:             db,
			AuthorizationManager: authorizationManager,
		})
		if err != nil {
			return err
		}

		slog.Info("Started web server", "port", viper.GetString("port"), "baseurl", viper.GetString("baseurl"))
		err = http.ListenAndServe(":"+viper.GetString("port"), app)
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	serveCmd.Flags().String("email.sendgrid", "", "Sendgrid API Key")

	serveCmd.Flags().String("email.sender", "no-reply@tedyst.ro", "Email sender")
	serveCmd.Flags().String("email.senderName", "Licenta", "Email sender name")

	serveCmd.Flags().String("baseurl", "http://localhost:8080", "Base URL")

	serveCmd.Flags().Int16P("port", "p", 5000, "Port to listen on")

	serveCmd.Flags().String("database", "", "Database connection string")

	serveCmd.Flags().String("hash-key", "", "Hash key used for signing Cookies")
	if err := serveCmd.MarkFlagRequired("hash-key"); err != nil {
		panic(err)
	}
	serveCmd.Flags().String("encrypt-key", "", "Encrypt key used for signing Cookies")
	if err := serveCmd.MarkFlagRequired("encrypt-key"); err != nil {
		panic(err)
	}

	rootCmd.AddCommand(serveCmd)
}
