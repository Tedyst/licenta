package user

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/spf13/cobra"
	database "github.com/tedyst/licenta/db"
	db "github.com/tedyst/licenta/db/generated"
	"github.com/tedyst/licenta/models"
)

// createadminCmd represents the createadmin command
var createadminCmd = &cobra.Command{
	Use:   "createadmin [email] [password] [username]",
	Short: "Create an admin user",
	Long:  `Create an admin user with the provided email and password.`,
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		database.InitDatabase()
		_, err := database.DatabaseQueries.GetUserByUsernameOrEmail(context.Background(), args[0])
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
		user, err := database.DatabaseQueries.CreateUser(context.Background(), db.CreateUserParams{
			Email:    args[0],
			Password: password,
			Username: args[2],
		})
		if err != nil {
			log.Panic(err)
		}
		err = database.DatabaseQueries.UpdateUser(context.Background(), db.UpdateUserParams{
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
