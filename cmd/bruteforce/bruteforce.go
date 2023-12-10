package bruteforce

import (
	"github.com/spf13/cobra"
)

func NewBruteforceCmd() *cobra.Command {
	return bruteforceCmd
}

var bruteforceCmd = &cobra.Command{
	Use:   "bruteforce",
	Short: "Bruteforce passwords management",
	Long:  `This command allows you to manage the default bruteforce passwords from the database`,
}

func init() {
	bruteforceCmd.PersistentFlags().String("database", "", "Database connection string")
	bruteforceCmd.MarkPersistentFlagRequired("database")
}
