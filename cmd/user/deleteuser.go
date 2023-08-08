package user

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	database "github.com/tedyst/licenta/db"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [username or email]",
	Short: "Delete a user by username or email",
	Long: `Delete a user by username or email. Usage:
	licenta deleteuser [username or email]`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		user, err := database.DatabaseQueries.GetUserByUsernameOrEmail(context.Background(), args[0])
		if err != nil {
			panic(err)
		}
		if user == nil {
			log.Panic("User does not exist.")
		}
		err = database.DatabaseQueries.DeleteUser(context.Background(), user.ID)
		if err != nil {
			log.Panic(err)
		}
		fmt.Println("User deleted.")
	},
}

func init() {
	userCmd.AddCommand(deleteCmd)
}
