package user

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tedyst/licenta/db"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [username or email]",
	Short: "Delete a user by username or email",
	Long: `Delete a user by username or email. Usage:
	licenta deleteuser [username or email]`,
	Args: cobra.ExactArgs(1),
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
			if errors.Is(err, pgx.ErrNoRows) {
				log.Fatal("User does not exist.")
			}
			log.Panic(err)
		}
		err = database.DeleteUser(context.Background(), user.ID)
		if err != nil {
			log.Panic(err)
		}
		fmt.Printf("User %s deleted.", user.Username)
	},
}

func init() {
	userCmd.AddCommand(deleteCmd)
}
