package user

import (
	"context"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tedyst/licenta/api/authorization"
	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/db/queries"
)

var changeOrgPermCmd = &cobra.Command{
	Use:   "changeorgperm",
	Short: "Change the permission of a user in an organization",
	Long:  `Change the permission of a user in an organization. This command is intended only for recovery/initial setup. For the normal use, use the web interface.`,
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		database := db.InitDatabase(viper.GetString("database"))

		user, err := database.GetUserByUsernameOrEmail(context.Background(), queries.GetUserByUsernameOrEmailParams{
			Username: args[0],
		})
		if err != nil {
			return err
		}

		orgId, err := strconv.Atoi(args[1])
		if err != nil {
			return err
		}

		organization, err := database.GetOrganization(context.Background(), int64(orgId))
		if err != nil {
			return err
		}

		_, err = database.GetOrganizationUser(context.Background(), queries.GetOrganizationUserParams{
			OrganizationID: organization.ID,
			UserID:         user.ID,
		})
		if err != nil {
			if err == pgx.ErrNoRows {
				_, err = database.AddOrganizationUser(context.Background(), queries.AddOrganizationUserParams{
					OrganizationID: organization.ID,
					UserID:         user.ID,
					Role:           int32(authorization.None),
				})
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}

		switch perm := args[2]; perm {
		case "owner":
			_, err = database.SetOrganizationPermissionsForUser(context.Background(), queries.SetOrganizationPermissionsForUserParams{
				OrganizationID: organization.ID,
				UserID:         user.ID,
				Role:           int32(authorization.Owner),
			})
		case "admin":
			_, err = database.SetOrganizationPermissionsForUser(context.Background(), queries.SetOrganizationPermissionsForUserParams{
				OrganizationID: organization.ID,
				UserID:         user.ID,
				Role:           int32(authorization.Admin),
			})
		case "viewer":
			_, err = database.SetOrganizationPermissionsForUser(context.Background(), queries.SetOrganizationPermissionsForUserParams{
				OrganizationID: organization.ID,
				UserID:         user.ID,
				Role:           int32(authorization.Viewer),
			})
		case "none":
			_, err = database.RemoveOrganizationUser(context.Background(), queries.RemoveOrganizationUserParams{
				OrganizationID: organization.ID,
				UserID:         user.ID,
			})
		}

		return err
	},
}

func init() {
	userCmd.AddCommand(changeOrgPermCmd)
}
