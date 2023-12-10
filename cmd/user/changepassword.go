package user

import (
	"context"
	"database/sql"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/models"
)

var changepasswordCmd = &cobra.Command{
	Use:   "changepassword",
	Short: "Change the password of a user",
	Long:  `Change the password of a user. This command is intended only for recovery/initial setup. For the normal use, use the web interface.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		var database db.TransactionQuerier
		database = db.InitDatabase(viper.GetString("database"))
		database, err := database.StartTransaction(context.Background())
		if err != nil {
			return err
		}
		defer database.EndTransaction(context.Background(), err)

		user, err := database.GetUserByUsernameOrEmail(context.Background(), args[0])
		if err != nil {
			return err
		}
		if user == nil {
			return nil
		}

		hash, err := models.GenerateHash(context.Background(), args[1])
		if err != nil {
			return err
		}

		err = database.UpdateUser(context.Background(), queries.UpdateUserParams{
			ID:       user.ID,
			Password: sql.NullString{Valid: true, String: hash},
		})
		if err != nil {
			return err
		}
		log.Println("Password changed.")
		return nil
	},
}

func init() {
	userCmd.AddCommand(changepasswordCmd)
}
