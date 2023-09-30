package user

import (
	"context"
	"database/sql"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tedyst/licenta/db"
	database "github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/db/queries"
)

var changepasswordCmd = &cobra.Command{
	Use:   "changepassword",
	Short: "Change the password of a user",
	Long: `Change the password of a user. Usage:
	licenta changepassword [username or email] [new password]`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		var db db.TransactionQuerier
		db = database.InitDatabase(viper.GetString("database"))
		db, err := db.StartTransaction(context.Background())
		if err != nil {
			log.Panic(err)
		}
		defer db.EndTransaction(context.Background(), err)

		user, err := db.GetUserByUsernameOrEmail(context.Background(), args[0])
		if err != nil {
			panic(err)
		}
		if user == nil {
			log.Panic("User does not exist.")
		}
		err = db.UpdateUser(context.Background(), queries.UpdateUserParams{
			ID:       user.ID,
			Password: sql.NullString{Valid: true, String: args[1]},
		})
		if err != nil {
			log.Panic(err)
		}
		log.Println("Password changed.")
	},
}

func init() {
	userCmd.AddCommand(changepasswordCmd)
}
