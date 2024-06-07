package scan

import (
	"context"

	"log/slog"

	r "github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tedyst/licenta/bruteforce"
	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/scanner"
	"github.com/tedyst/licenta/scanner/redis"
)

var scanRedisCmd = &cobra.Command{
	Use:   "redis [connection string]",
	Short: "Scan a Redis database",
	Long:  `This command allows you to scan a Redis database manually. It does not update the database, so the results will not be visible in the web interface. The database connection string is not required, but is recommended for the bruteforce module. If a bruteforce is successful, the database will store that result in order to avoid repeating the bruteforce.`,
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		rdb := r.NewClient(&r.Options{
			Addr:     "localhost:6379",
			Username: "default",
			Password: "password",
			DB:       0,
		})

		sc, err := redis.NewScanner(context.Background(), rdb)
		if err != nil {
			return err
		}

		err = sc.Ping(ctx)
		if err != nil {
			return err
		}
		slog.InfoContext(cmd.Context(), "Connection established")

		v, err := sc.GetVersion(ctx)
		if err != nil {
			return err
		}
		slog.InfoContext(cmd.Context(), "Got version", "version", v)

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
			passProvider, err := bruteforce.NewDatabasePasswordProvider(ctx, database, 1)
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
	scanRedisCmd.Flags().String("database", "", "Database connection string")
	scanCmd.AddCommand(scanRedisCmd)
}
