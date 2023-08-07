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

var changepasswordCmd = &cobra.Command{
	Use:   "changepassword",
	Short: "Change the password of a user",
	Long: `Change the password of a user. Usage:
	licenta changepassword [username or email] [new password]`,
	Run: func(cmd *cobra.Command, args []string) {
		user, err := database.DatabaseQueries.GetUserByUsernameOrEmail(context.Background(), args[0])
		if err != nil {
			panic(err)
		}
		if user == nil {
			log.Panic("User does not exist.")
		}
		err = database.DatabaseQueries.UpdateUser(context.Background(), db.UpdateUserParams{
			ID:       user.ID,
			Password: args[1],
		})
		if err != nil {
			log.Panic(err)
		}
		log.Println("Password changed.")
	},
}

func init() {
	rootCmd.AddCommand(changepasswordCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// changepasswordCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// changepasswordCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
