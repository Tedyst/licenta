package user

import (
	"context"
	"errors"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/models"
)

// createadminCmd represents the createadmin command
var createadminCmd = &cobra.Command{
	Use:   "createadmin [email] [password] [username]",
	Short: "Create an admin user",
	Long:  `Create an admin user with the provided email and password. This command is intended only for recovery/initial setup. For the normal use, use the web interface.`,
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		var database db.TransactionQuerier
		database = db.InitDatabase(viper.GetString("database"))
		database, err = database.StartTransaction(context.Background())
		if err != nil {
			return err
		}
		defer func() {
			err = errors.Join(err, database.EndTransaction(context.Background(), err))
		}()

		_, err = database.GetUserByUsernameOrEmail(context.Background(), args[0])
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		if err == nil {
			return errors.New("user already exists")
		}

		password, err := models.GenerateHash(context.Background(), args[1])
		if err != nil {
			return err
		}

		user, err := database.CreateUser(context.Background(), queries.CreateUserParams{
			Email:    args[0],
			Password: password,
			Username: args[2],
		})
		if err != nil {
			return err
		}
		log.Printf("Admin user %s created.", user.Username)
		return nil
	},
}

func init() {
	userCmd.AddCommand(createadminCmd)
}
