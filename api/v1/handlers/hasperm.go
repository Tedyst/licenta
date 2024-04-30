package handlers

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/tedyst/licenta/api/authorization"
	"github.com/tedyst/licenta/db/queries"
)

func checkUserHasProjectPermission[TError ~struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}](server *serverHandler, ctx context.Context, projectID int64, role authorization.RBACGroup) (*queries.User, *queries.Project, TError, error) {
	user, err := server.userAuth.GetUser(ctx)
	if err != nil {
		return nil, nil, TError{}, fmt.Errorf("error getting user: %w", err)
	}
	if user == nil {
		return nil, nil, TError{Message: "You are not authorized to view this project", Success: false}, nil
	}

	project, err := server.DatabaseProvider.GetProject(ctx, projectID)
	if err != nil && err != pgx.ErrNoRows {
		return nil, nil, TError{}, fmt.Errorf("error getting project: %w", err)
	}
	if err == pgx.ErrNoRows {
		return nil, nil, TError{Message: "You are not authorized to view this project", Success: false}, nil
	}

	hasPerm, err := server.authorization.UserHasPermissionForProject(ctx, project, user, role)
	if err != nil {
		return user, nil, TError{}, fmt.Errorf("error checking permissions: %w", err)
	}
	if !hasPerm {
		return user, nil, TError{Message: "You are not authorized to view this project", Success: false}, nil
	}

	return user, project, TError{Success: true}, nil
}
