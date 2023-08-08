package user

import (
	"github.com/spf13/cobra"
)

func NewUserCmd() *cobra.Command {
	return userCmd
}

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Users management",
	Long:  `Users management.`,
}
