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

var taskScanMysqlCmd = &cobra.Command{
	Use:   "scanmysql [databaseID]",
	Short: "Scan mysql DB task",
	Long:  `This task will scan a mysql database.`,
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

		db, err := database.GetMysqlDatabase(cmd.Context(), dbid)
		if err != nil {
			return err
		}

		scanGroup, err := database.CreateScanGroup(cmd.Context(), queries.CreateScanGroupParams{
			ProjectID: db.MysqlDatabase.ProjectID,
		})
		if err != nil {
			return err
		}

		scan, err := database.CreateScan(cmd.Context(), queries.CreateScanParams{
			Status:      models.SCAN_NOT_STARTED,
			ScanGroupID: scanGroup.ID,
		})
		if err != nil {
			return err
		}

		_, err = database.CreateMysqlScan(cmd.Context(), queries.CreateMysqlScanParams{
			ScanID:     scan.ID,
			DatabaseID: dbid,
		})
		if err != nil {
			return err
		}

		return taskRunner.RunSaverRemote(cmd.Context(), scan, "mysql")
	},
}

func init() {
	tasksCmd.AddCommand(taskScanMysqlCmd)
}
