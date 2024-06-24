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
	v1 "github.com/tedyst/licenta/api/v1"
	"github.com/tedyst/licenta/api/v1/handlers"
	"github.com/tedyst/licenta/bruteforce"
	"github.com/tedyst/licenta/cache"
	database "github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/email"
	localExchange "github.com/tedyst/licenta/messages/local"
	"github.com/tedyst/licenta/tasks/local"
)

var serveLocalCmd = &cobra.Command{
	Use:   "servelocal",
	Short: "Run the server using the development configuration",
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
		brutefroceProvider := bruteforce.NewDatabaseBruteforceProvider(db)

		taskRunner := local.NewLocalRunner(viper.GetBool("debug"), email.NewConsoleEmailSender(
			"no-reply@localhost",
			"no-reply@localhost",
		), db, localExchange, brutefroceProvider, viper.GetString("db-encryption-salt"))

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

		serverCache, err := cache.NewLocalCacheProvider[string]()
		if err != nil {
			return err
		}

		app, err := api.Initialize(api.ApiConfig{
			Origin: viper.GetString("baseurl"),
			ApiV1Config: v1.ApiV1Config{
				Debug: viper.GetBool("debug"),
				HandlerConfig: handlers.HandlerConfig{
					TaskRunner:           taskRunner,
					MessageExchange:      localExchange,
					AuthorizationManager: authorizationManager,
					SaltKey:              viper.GetString("db-encryption-salt"),
					DatabaseProvider:     db,
					WorkerAuth:           workerAuth,
					UserAuth:             userAuth,
					Cache:                serverCache,
				},
			},
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
	serveLocalCmd.Flags().String("baseurl", "localhost", "Base URL")

	serveLocalCmd.Flags().Int16P("port", "p", 5000, "Port to listen on")

	serveLocalCmd.Flags().String("database", "", "Database connection string")

	serveLocalCmd.Flags().String("hash-key", "", "Hash key used for signing Cookies")
	if err := serveLocalCmd.MarkFlagRequired("hash-key"); err != nil {
		panic(err)
	}
	serveLocalCmd.Flags().String("encrypt-key", "", "Encrypt key used for signing Cookies")
	if err := serveLocalCmd.MarkFlagRequired("encrypt-key"); err != nil {
		panic(err)
	}
	serveLocalCmd.Flags().String("db-encryption-salt", "", "Database salt encryption key")
	if err := serveLocalCmd.MarkFlagRequired("db-encryption-salt"); err != nil {
		panic(err)
	}

	rootCmd.AddCommand(serveLocalCmd)
}
