package user

import (
	"github.com/spf13/cobra"
)

func NewUserCmd() *cobra.Command {
	return userCmd
}

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "User management",
	Long:  `Allows you to manage users and their permissions. These commands are intended only for recovery/initial setup. For the normal use, use the web interface.`,
}

func init() {
	userCmd.Flags().String("database", "", "Database connection string")
}
