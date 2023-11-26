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
	Short: "Run the extractor tool for the provided file",
	Long:  `Run the extractor tool for the provided file`,
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
