package scan

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/spf13/cobra"
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

		scanner, err := postgres.NewScanner(context.Background(), conn)
		if err != nil {
			return err
		}

		err = scanner.Ping(ctx)
		if err != nil {
			return err
		}
		slog.Info("Connection established")

		err = scanner.CheckPermissions(ctx)
		if err != nil {
			return err
		}
		slog.Info("Permissions checked")

		results, err := scanner.ScanConfig(ctx)
		if err != nil {
			return err
		}
		slog.Info("Config scanned")

		for _, result := range results {
			slog.Info("Config result: %s", "config", result)
		}

		users, err := scanner.GetUsers(ctx)
		if err != nil {
			return err
		}
		slog.Info("Users scanned")

		for _, user := range users {
			slog.Info("User: %s", "user", user)
		}

		return nil
	},
}

func init() {
	scanCmd.AddCommand(scanPostgresCmd)
}
