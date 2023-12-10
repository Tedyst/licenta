package migrate

import (
	"log/slog"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tedyst/licenta/db"
)

var forceCmd = &cobra.Command{
	Use:   "force",
	Short: "Force a migration to a specific version",
	Long:  `This command allows you to force a migration to a specific version. If no migrations are found, the command will do nothing.This dosen't do a migration, just sets the version in the database.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		database := db.InitDatabase(viper.GetString("database"))
		std := stdlib.OpenDBFromPool(database.GetRawPool())

		driver, err := pgx.WithInstance(std, &pgx.Config{})
		if err != nil {
			return err
		}

		fs, err := iofs.New(db.Migrations, "migrations")
		if err != nil {
			return err
		}

		m, err := migrate.NewWithInstance("iofs", fs, "postgres", driver)
		if err != nil {
			return err
		}

		version, err := strconv.ParseInt(args[0], 10, 32)
		if err != nil {
			return err
		}

		err = m.Force(int(version))
		if err != nil {
			return err
		}

		slog.InfoContext(cmd.Context(), "Migration forced to specific version", "version", version)
		return nil
	},
}

func init() {
	migrateCmd.AddCommand(forceCmd)
}
