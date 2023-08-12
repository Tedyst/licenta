package user

import (
	"context"
	"database/sql"
	"log"

	"github.com/spf13/cobra"
	database "github.com/tedyst/licenta/db"
	db "github.com/tedyst/licenta/db/generated"
)

var changepasswordCmd = &cobra.Command{
	Use:   "changepassword",
	Short: "Change the password of a user",
	Long: `Change the password of a user. Usage:
	licenta changepassword [username or email] [new password]`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		database.InitDatabase()
		user, err := database.DatabaseQueries.GetUserByUsernameOrEmail(context.Background(), args[0])
		if err != nil {
			panic(err)
		}
		if user == nil {
			log.Panic("User does not exist.")
		}
		err = database.DatabaseQueries.UpdateUser(context.Background(), db.UpdateUserParams{
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
