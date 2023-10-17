package scan

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tedyst/licenta/scanner/postgres"
	"golang.org/x/exp/slog"
)

var scanPostgresCmd = &cobra.Command{
	Use:   "postgres",
	Short: "Run the extractor tool for the provided file",
	Long:  `Run the extractor tool for the provided file`,
	Run: func(cmd *cobra.Command, args []string) {
		connString := viper.GetString("connection-string")

		fmt.Println("Scanning postgres...")

		ctx := context.Background()

		conn, err := pgx.Connect(ctx, connString)
		if err != nil {
			panic(err)
		}

		scanner, err := postgres.NewScanner(context.Background(), conn)
		if err != nil {
			panic(err)
		}

		err = scanner.Ping(ctx)
		if err != nil {
			panic(err)
		}
		slog.Info("Connection established")

		err = scanner.CheckPermissions(ctx)
		if err != nil {
			panic(err)
		}
		slog.Info("Permissions checked")

		results, err := scanner.ScanConfig(ctx)
		if err != nil {
			panic(err)
		}
		slog.Info("Config scanned")

		for _, result := range results {
			slog.Info("Config result: %s", "config", result)
		}

		users, err := scanner.GetUsers(ctx)
		if err != nil {
			panic(err)
		}
		slog.Info("Users scanned")

		for _, user := range users {
			slog.Info("User: %s", "user", user)
		}
	},
}

func init() {
	scanPostgresCmd.Flags().String("connection-string", "", "Connection string to the postgres database")
	scanPostgresCmd.MarkFlagRequired("connection-string")

	scanCmd.AddCommand(scanPostgresCmd)
}
