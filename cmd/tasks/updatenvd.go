package tasks

import (
	"errors"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/nvd"
	"github.com/tedyst/licenta/tasks/local"
)

var updateNVDTask = &cobra.Command{
	Use:   "updatenvd [product]",
	Short: "Scan postgres DB task",
	Long:  `Scan postgres DB task`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var product nvd.Product
		switch args[0] {
		case "postgres":
			product = nvd.POSTGRESQL
		default:
			return errors.New("product does not exist")
		}

		database := db.InitDatabase(viper.GetString("database"))

		taskRunner := local.NewLocalRunner(true, nil, database)

		return taskRunner.UpdateNVDVulnerabilitiesForProduct(cmd.Context(), product)
	},
}

func init() {
	updateNVDTask.Flags().String("database", "", "Database connection string")
	updateNVDTask.MarkFlagRequired("database")

	tasksCmd.AddCommand(updateNVDTask)
}
