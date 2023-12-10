package bruteforce

import (
	"net/http"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tedyst/licenta/bruteforce"
	"github.com/tedyst/licenta/db"
	"golang.org/x/text/encoding/charmap"
)

var downloadCmd = &cobra.Command{
	Use:   "download [file]",
	Short: "Download a file and import it into the database",
	Long:  `Reads a file from the internet and imports it into the database. The file must be in ISO8859-1 encoding, otherwise the import will fail. The file must be a text file, with one password per line. Duplicate passwords will be ignored.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

		database, err := db.InitDatabase(viper.GetString("database")).StartTransaction(cmd.Context())
		if err != nil {
			return err
		}
		defer database.EndTransaction(cmd.Context(), err)

		reader, err := http.Get(args[0])
		if err != nil {
			return err
		}
		defer reader.Body.Close()

		return bruteforce.ImportFromReader(cmd.Context(), charmap.ISO8859_1.NewDecoder().Reader(reader.Body), database)
	},
}

func init() {
	bruteforceCmd.AddCommand(downloadCmd)
}
