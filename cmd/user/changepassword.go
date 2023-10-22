package user

import (
	"context"
	"database/sql"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/db/queries"
)

var changepasswordCmd = &cobra.Command{
	Use:   "changepassword",
	Short: "Change the password of a user",
	Long: `Change the password of a user. Usage:
	licenta changepassword [username or email] [new password]`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		var database db.TransactionQuerier
		database = db.InitDatabase(viper.GetString("database"))
		database, err := database.StartTransaction(context.Background())
		if err != nil {
			log.Panic(err)
		}
		defer database.EndTransaction(context.Background(), err)

		user, err := database.GetUserByUsernameOrEmail(context.Background(), args[0])
		if err != nil {
			panic(err)
		}
		if user == nil {
			log.Panic("User does not exist.")
		}
		err = database.UpdateUser(context.Background(), queries.UpdateUserParams{
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
