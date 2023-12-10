package bruteforce

import (
	"net/http"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tedyst/licenta/bruteforce"
	"github.com/tedyst/licenta/db"
)

const baseCainAndAbelURL = "https://raw.githubusercontent.com/danielmiessler/SecLists/master/Passwords/Software/cain-and-abel.txt"

var cainAndAbelCmd = &cobra.Command{
	Use:   "cain-and-abel",
	Short: "Download the cain-and-abel.txt file and import it into the database",
	Long:  `This command downloads the cain-and-abel password list and imports it into the database.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

		database, err := db.InitDatabase(viper.GetString("database")).StartTransaction(cmd.Context())
		if err != nil {
			return err
		}
		defer database.EndTransaction(cmd.Context(), err)

		reader, err := http.Get(baseCainAndAbelURL)
		if err != nil {
			return err
		}
		defer reader.Body.Close()

		return bruteforce.ImportFromReader(cmd.Context(), reader.Body, database)
	},
}

func init() {
	bruteforceCmd.AddCommand(cainAndAbelCmd)
}
