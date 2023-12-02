package scan

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tedyst/licenta/bruteforce"
	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/scanner"
	"github.com/tedyst/licenta/scanner/postgres"
	"golang.org/x/exp/slog"
)

var scanPostgresCmd = &cobra.Command{
	Use:   "postgres [connection string]",
	Short: "Run the extractor tool for the provided file",
	Long:  `Run the extractor tool for the provided file`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		conn, err := pgx.Connect(ctx, args[0])
		if err != nil {
			return err
		}

		sc, err := postgres.NewScanner(context.Background(), conn)
		if err != nil {
			return err
		}

		err = sc.Ping(ctx)
		if err != nil {
			return err
		}
		slog.Info("Connection established")

		err = sc.CheckPermissions(ctx)
		if err != nil {
			return err
		}
		slog.Info("Permissions checked")

		results, err := sc.ScanConfig(ctx)
		if err != nil {
			return err
		}
		slog.Info("Config scanned")

		for _, result := range results {
			slog.Info("Config result: %s", "config", result)
		}

		users, err := sc.GetUsers(ctx)
		if err != nil {
			return err
		}
		slog.Info("Users scanned")

		for _, user := range users {
			slog.Info("User: %s", "user", user)
		}

		if viper.GetString("database") != "" {
			database := db.InitDatabase(viper.GetString("database"))
			result, err := bruteforce.BruteforcePasswordAllUsers(cmd.Context(), sc, database, func(m map[scanner.User]bruteforce.BruteforceUserStatus) error {
				slog.Info("Received update from bruteforce", "update", m)
				return nil
			})
			slog.Info("asd", "result", result, "err", err)
		}

		return nil
	},
}

func init() {
	scanPostgresCmd.Flags().String("database", "", "Database connection string")
	scanCmd.AddCommand(scanPostgresCmd)
}
