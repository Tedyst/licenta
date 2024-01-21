package tasks

import (
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tedyst/licenta/bruteforce"
	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/db/queries"
	localExchange "github.com/tedyst/licenta/messages/local"
	"github.com/tedyst/licenta/models"
	"github.com/tedyst/licenta/tasks/local"
)

var taskScanPostgresCmd = &cobra.Command{
	Use:   "scanpostgres [databaseID]",
	Short: "Scan postgres DB task",
	Long:  `This task will scan a postgres database. The parameter is the ID of the postgres_database entry in the database. A new postgres scan object will be created and started. The scan will update the postgres_scan object with the status and the results.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		database := db.InitDatabase(viper.GetString("database"))

		localExchange := localExchange.NewLocalExchange()
		bruteforceProvider := bruteforce.NewDatabaseBruteforceProvider(database)

		taskRunner := local.NewLocalRunner(true, nil, database, localExchange, bruteforceProvider)

		dbid, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return err
		}

		db, err := database.GetPostgresDatabase(cmd.Context(), dbid)
		if err != nil {
			return err
		}

		scan, err := database.CreateScan(cmd.Context(), queries.CreateScanParams{
			Status:    models.SCAN_NOT_STARTED,
			ProjectID: db.PostgresDatabase.ProjectID,
		})
		if err != nil {
			return err
		}

		postgresScan, err := database.CreatePostgresScan(cmd.Context(), queries.CreatePostgresScanParams{
			ScanID:     scan.ID,
			DatabaseID: dbid,
		})
		if err != nil {
			return err
		}

		return taskRunner.ScanPostgresDB(cmd.Context(), postgresScan)
	},
}

func init() {
	tasksCmd.AddCommand(taskScanPostgresCmd)
}
