package tasks

import (
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/models"
	"github.com/tedyst/licenta/tasks/local"
)

var taskScanPostgresCmd = &cobra.Command{
	Use:   "scanpostgres [databaseID]",
	Short: "Scan postgres DB task",
	Long:  `Scan postgres DB task`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		database := db.InitDatabase(viper.GetString("database"))

		taskRunner := local.NewLocalRunner(true, nil, database)

		dbid, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return err
		}

		scan, err := database.CreatePostgresScan(cmd.Context(), queries.CreatePostgresScanParams{
			PostgresDatabaseID: dbid,
			Status:             models.SCAN_NOT_STARTED,
		})
		if err != nil {
			return err
		}

		return taskRunner.ScanPostgresDB(cmd.Context(), scan)
	},
}

func init() {
	taskScanPostgresCmd.Flags().String("database", "", "Database connection string")
	taskScanPostgresCmd.MarkFlagRequired("database")

	tasksCmd.AddCommand(taskScanPostgresCmd)
}
