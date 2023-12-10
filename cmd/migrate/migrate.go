package migrate

import (
	"github.com/spf13/cobra"
)

func NewMigrateCmd() *cobra.Command {
	return migrateCmd
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrations management",
	Long:  `This command allows you to manage migrations for the database.`,
}

func init() {
	migrateCmd.PersistentFlags().String("database", "", "Database connection string")
	migrateCmd.MarkPersistentFlagRequired("database")
}
