package tasks

import (
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tedyst/licenta/bruteforce"
	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/db/queries"
	localExchange "github.com/tedyst/licenta/messages/local"
	"github.com/tedyst/licenta/saver"
	"github.com/tedyst/licenta/tasks/local"
)

var taskCreateScan = &cobra.Command{
	Use:   "createscan [scanType] [databaseID]",
	Short: "Scan any DB task",
	Long:  `This task will scan a specific database type.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		database := db.InitDatabase(viper.GetString("database"))

		localExchange := localExchange.NewLocalExchange()
		bruteforceProvider := bruteforce.NewDatabaseBruteforceProvider(database)

		taskRunner := local.NewLocalRunner(true, nil, database, localExchange, bruteforceProvider, viper.GetString("db-encryption-salt"))

		dbid, err := strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			return err
		}

		db, err := database.GetMysqlDatabase(cmd.Context(), queries.GetMysqlDatabaseParams{
			ID:      dbid,
			SaltKey: viper.GetString("db-encryption-salt"),
		})
		if err != nil {
			return err
		}

		scanGroup, err := database.CreateScanGroup(cmd.Context(), queries.CreateScanGroupParams{
			ProjectID: db.ProjectID,
		})
		if err != nil {
			return err
		}

		scans, err := saver.CreateScans(cmd.Context(), database, scanGroup.ID, -1, args[0])
		if err != nil {
			return err
		}

		if len(scans) == 0 {
			return nil
		}

		return taskRunner.RunSaverRemote(cmd.Context(), scans[0], args[0])
	},
}

func init() {
	tasksCmd.AddCommand(taskCreateScan)
}
