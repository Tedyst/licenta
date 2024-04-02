package user

import (
	"context"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tedyst/licenta/api/auth"
	"github.com/tedyst/licenta/cache"
	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/db/queries"
)

var changepasswordCmd = &cobra.Command{
	Use:   "changepassword",
	Short: "Change the password of a user",
	Long:  `Change the password of a user. This command is intended only for recovery/initial setup. For the normal use, use the web interface.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		database := db.InitDatabase(viper.GetString("database"))
		c, err := cache.NewLocalCacheProvider[queries.User]()
		if err != nil {
			return err
		}

		userAuth, err := auth.NewAuthenticationProvider("http://localhost:5000", database, nil, nil, nil, c)
		if err != nil {
			return err
		}

		user, err := database.GetUserByUsernameOrEmail(context.Background(), queries.GetUserByUsernameOrEmailParams{
			Username: args[0],
			Email:    args[0],
		})
		if err != nil {
			return err
		}

		userAuth.UpdatePassword(cmd.Context(), user, args[1])

		slog.Info("Password changed for user", "user", args[0])

		return nil
	},
}

func init() {
	userCmd.AddCommand(changepasswordCmd)
}
