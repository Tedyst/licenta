/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"log"

	"github.com/spf13/cobra"
	database "github.com/tedyst/licenta/db"
	db "github.com/tedyst/licenta/db/generated"
)

// createadminCmd represents the createadmin command
var createadminCmd = &cobra.Command{
	Use:   "createadmin [email] [password] [username]",
	Short: "Create an admin user",
	Long:  `Create an admin user with the provided email and password.`,
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		user, err := database.DatabaseQueries.GetUserByUsernameOrEmail(context.Background(), args[0])
		if err != nil {
			panic(err)
		}
		if user != nil {
			log.Panic("User already exists. Please use a different username or email.")
		}
		user, err = database.DatabaseQueries.CreateUser(context.Background(), db.CreateUserParams{
			Email:    args[0],
			Password: args[1],
			Username: args[2],
		})
		if err != nil {
			log.Panic(err)
		}
		err = database.DatabaseQueries.UpdateUser(context.Background(), db.UpdateUserParams{
			ID:    user.ID,
			Admin: true,
		})
		if err != nil {
			log.Panic(err)
		}

	},
}

func init() {
	rootCmd.AddCommand(createadminCmd)
}
