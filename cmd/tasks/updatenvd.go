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
	Short: "Update NVD vulnerabilities for a product",
	Long:  `This task will update the NVD vulnerabilities for a product. The parameter is the product name. Currently only postgres is supported. The task will update the nvd_cpes and nvd_cves tables on success. If the product does not exist, or the update fails, the tables will not be updated. Will not schedule sending notifications for the new vulnerabilities.`,
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
