package user

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tedyst/licenta/db"
	database "github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/models"
)

// createadminCmd represents the createadmin command
var createadminCmd = &cobra.Command{
	Use:   "createadmin [email] [password] [username]",
	Short: "Create an admin user",
	Long:  `Create an admin user with the provided email and password.`,
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		var db db.TransactionQuerier
		db = database.InitDatabase(viper.GetString("database"))
		db, err := db.StartTransaction(context.Background())
		if err != nil {
			log.Panic(err)
		}
		defer db.EndTransaction(context.Background(), err)

		_, err = db.GetUserByUsernameOrEmail(context.Background(), args[0])
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			log.Panic(err)
		}
		if err == nil {
			log.Fatal("User already exists.")
		}
		password, err := models.GenerateHash(context.Background(), args[1])
		if err != nil {
			log.Panic(err)
		}
		user, err := db.CreateUser(context.Background(), queries.CreateUserParams{
			Email:    args[0],
			Password: password,
			Username: args[2],
		})
		if err != nil {
			log.Panic(err)
		}
		err = db.UpdateUser(context.Background(), queries.UpdateUserParams{
			ID:    user.ID,
			Admin: sql.NullBool{Valid: true, Bool: true},
		})
		if err != nil {
			log.Panic(err)
		}
		log.Printf("Admin user %s created.", user.Username)
	},
}

func init() {
	userCmd.AddCommand(createadminCmd)
}
