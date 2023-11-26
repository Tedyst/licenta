package bruteforce

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tedyst/licenta/db"
)

var listPasswordsCmd = &cobra.Command{
	Use:   "list-passwords",
	Short: "Run the extractor tool for the provided file",
	Long:  `Run the extractor tool for the provided file`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		database := db.InitDatabase(viper.GetString("database"))

		rows, err := database.GetRawPool().Query(cmd.Context(), "SELECT password FROM default_bruteforce_passwords")
		if err != nil {
			return err
		}

		for rows.Next() {
			var password string
			err = rows.Scan(&password)
			if err != nil {
				return err
			}
			cmd.Println(password)
		}

		return nil
	},
}

func init() {
	bruteforceCmd.AddCommand(listPasswordsCmd)
}
