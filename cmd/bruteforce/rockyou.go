package bruteforce

import (
	"net/http"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tedyst/licenta/bruteforce"
	"github.com/tedyst/licenta/db"
	"golang.org/x/text/encoding/charmap"
)

const baseRockyouURL = "https://github.com/brannondorsey/naive-hashcat/releases/download/data/rockyou.txt"

var rockyouCmd = &cobra.Command{
	Use:   "rockyou",
	Short: "Download the rockyou.txt file and import it into the database",
	Long:  `This command downloads the rockyou password list and imports it into the database.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

		database, err := db.InitDatabase(viper.GetString("database")).StartTransaction(cmd.Context())
		if err != nil {
			return err
		}
		defer database.EndTransaction(cmd.Context(), err)

		reader, err := http.Get(baseRockyouURL)
		if err != nil {
			return err
		}
		defer reader.Body.Close()

		return bruteforce.ImportFromReader(cmd.Context(), charmap.ISO8859_1.NewDecoder().Reader(reader.Body), database)
	},
}

func init() {
	bruteforceCmd.AddCommand(rockyouCmd)
}
