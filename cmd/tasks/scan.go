package tasks

import (
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tedyst/licenta/bruteforce"
	"github.com/tedyst/licenta/db"
	localExchange "github.com/tedyst/licenta/messages/local"
	"github.com/tedyst/licenta/tasks/local"
)

var taskScanCmd = &cobra.Command{
	Use:   "scan [scanID]",
	Short: "Scan DB task",
	Long:  `This task will scan using a Scanner.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		database := db.InitDatabase(viper.GetString("database"))

		localExchange := localExchange.NewLocalExchange()
		bruteforceProvider := bruteforce.NewDatabaseBruteforceProvider(database)

		taskRunner := local.NewLocalRunner(true, nil, database, localExchange, bruteforceProvider)

		scanID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return err
		}

		scan, err := database.GetScan(cmd.Context(), scanID)
		if err != nil {
			return err
		}

		return taskRunner.RunSaverRemote(cmd.Context(), &scan.Scan, "all")
	},
}

func init() {
	tasksCmd.AddCommand(taskScanCmd)
}
