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
	"github.com/tedyst/licenta/db/queries"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [username or email]",
	Short: "Delete a user by username or email",
	Long:  `Delete a user by username or email. This command is intended only for recovery/initial setup. For the normal use, use the web interface.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		var database db.TransactionQuerier
		database = db.InitDatabase(viper.GetString("database"))
		database, err = database.StartTransaction(context.Background())
		if err != nil {
			return err
		}
		defer func() {
			err = errors.Join(err, database.EndTransaction(context.Background(), err != nil))
		}()

		user, err := database.GetUserByUsernameOrEmail(context.Background(), queries.GetUserByUsernameOrEmailParams{
			Username: args[0],
			Email:    args[0],
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("user not found")
			}
			return err
		}
		err = database.DeleteUser(context.Background(), user.ID)
		if err != nil {
			return err
		}
		log.Printf("User %s deleted.", user.Username)
		return nil
	},
}

func init() {
	userCmd.AddCommand(deleteCmd)
}
