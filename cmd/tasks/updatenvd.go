package tasks

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tedyst/licenta/bruteforce"
	"github.com/tedyst/licenta/db"
	localExchange "github.com/tedyst/licenta/messages/local"
	"github.com/tedyst/licenta/nvd"
	"github.com/tedyst/licenta/tasks/local"
)

var updateNVDTask = &cobra.Command{
	Use:   "updatenvd [product]",
	Short: "Update NVD vulnerabilities for a product",
	Long:  `This task will update the NVD vulnerabilities for a product. The parameter is the product name. Currently only postgres is supported. The task will update the nvd_cpes and nvd_cves tables on success. If the product does not exist, or the update fails, the tables will not be updated. Will not schedule sending notifications for the new vulnerabilities.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		product := nvd.GetNvdProductType(args[0])

		database := db.InitDatabase(viper.GetString("database"))

		transaction, err := database.StartTransaction(cmd.Context())
		if err != nil {
			return err
		}

		localExchange := localExchange.NewLocalExchange()
		bruteforceProvider := bruteforce.NewDatabaseBruteforceProvider(transaction)

		taskRunner := local.NewLocalRunner(true, nil, transaction, localExchange, bruteforceProvider, viper.GetString("db-encryption-salt"))

		err = taskRunner.UpdateNVDVulnerabilitiesForProduct(cmd.Context(), product)
		transaction.EndTransaction(cmd.Context(), err != nil)
		return err
	},
}

func init() {
	updateNVDTask.Flags().String("database", "", "Database connection string")
	if err := updateNVDTask.MarkFlagRequired("database"); err != nil {
		panic(err)
	}

	tasksCmd.AddCommand(updateNVDTask)
}
