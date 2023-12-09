package migrate

import (
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tedyst/licenta/db"
)

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Migrations management",
	Long:  `Migrations management.`,
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

		return m.Up()
	},
}

func init() {
	migrateCmd.AddCommand(upCmd)
}
