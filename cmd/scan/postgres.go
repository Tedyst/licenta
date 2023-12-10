package scan

import (
	"context"

	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tedyst/licenta/bruteforce"
	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/scanner"
	"github.com/tedyst/licenta/scanner/postgres"
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
		slog.InfoContext(cmd.Context(), "Connection established")

		err = sc.CheckPermissions(ctx)
		if err != nil {
			return err
		}
		slog.InfoContext(cmd.Context(), "Permissions checked")

		results, err := sc.ScanConfig(ctx)
		if err != nil {
			return err
		}
		slog.InfoContext(cmd.Context(), "Config scanned")

		for _, result := range results {
			switch result.Severity() {
			case scanner.SEVERITY_HIGH:
				slog.ErrorContext(cmd.Context(), "Config scan result", "detail", result.Detail(), "severity", result.Severity())
			case scanner.SEVERITY_MEDIUM:
				slog.WarnContext(cmd.Context(), "Config scan result", "detail", result.Detail(), "severity", result.Severity())
			case scanner.SEVERITY_WARNING:
				slog.InfoContext(cmd.Context(), "Config scan result", "detail", result.Detail(), "severity", result.Severity())
			default:
			}
		}

		users, err := sc.GetUsers(ctx)
		if err != nil {
			return err
		}

		for _, user := range users {
			slog.DebugContext(cmd.Context(), "User: %s", "user", user)
		}

		slog.Info("Users scanned")

		if viper.GetString("database") != "" {
			database := db.InitDatabase(viper.GetString("database"))
			passProvider, err := bruteforce.NewDatabasePasswordProvider(ctx, database, -1)
			if err != nil {
				return err
			}
			defer passProvider.Close()

			bruteforcer := bruteforce.NewBruteforcer(passProvider, sc, func(m map[scanner.User]bruteforce.BruteforceUserStatus) error {
				for user, entry := range m {
					username, err := user.GetUsername()
					if err != nil {
						return err
					}
					slog.InfoContext(cmd.Context(), "Received update from function", "user", username, "password", entry.FoundPassword, "tried", entry.Tried, "total", entry.Total)
				}
				return nil
			})

			result, err := bruteforcer.BruteforcePasswordAllUsers(ctx)
			if err != nil {
				return err
			}
			for _, result := range result {
				slog.InfoContext(cmd.Context(), "Bruteforced passwords", "detail", result.Detail(), "severity", result.Severity())
			}
		}

		return nil
	},
}

func init() {
	scanPostgresCmd.Flags().String("database", "", "Database connection string")
	scanCmd.AddCommand(scanPostgresCmd)
}
