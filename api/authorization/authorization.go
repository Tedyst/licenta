package authorization

import (
	"context"

	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/db/queries"
)

type RBACGroup int16

const (
	Owner RBACGroup = iota
	Admin
	Viewer
	None
)

type AuthorizationManager interface {
	UserHasPermissionForOrganization(ctx context.Context, organization *queries.Organization, user *queries.User, permission RBACGroup) (bool, error)
	UserHasPermissionForProject(ctx context.Context, project *queries.Project, user *queries.User, permission RBACGroup) (bool, error)

	WorkerHasPermissionForProject(ctx context.Context, project *queries.Project, worker *queries.Worker, permission RBACGroup) (bool, error)

	UserPermissionsChanged(ctx context.Context, user *queries.User) error
	WorkerPermissionsChanged(ctx context.Context, worker *queries.Worker) error
}

type authorizationManagerImpl struct {
	querier db.TransactionQuerier
}

func NewAuthorizationManager(querier db.TransactionQuerier) AuthorizationManager {
	return &authorizationManagerImpl{querier: querier}
}

func (a *authorizationManagerImpl) UserHasPermissionForOrganization(ctx context.Context, organization *queries.Organization, user *queries.User, permission RBACGroup) (bool, error) {
	p, err := a.querier.GetOrganizationPermissionsForUser(ctx, queries.GetOrganizationPermissionsForUserParams{
		OrganizationID: organization.ID,
		UserID:         user.ID,
	})
	if err != nil {
		return false, err
	}

	return RBACGroup(p) >= permission, nil
}

func (a *authorizationManagerImpl) UserHasPermissionForProject(ctx context.Context, project *queries.Project, user *queries.User, permission RBACGroup) (bool, error) {
	p, err := a.querier.GetProjectPermissionsForUser(ctx, queries.GetProjectPermissionsForUserParams{
		ProjectID:      project.ID,
		UserID:         user.ID,
		OrganizationID: project.OrganizationID,
	})
	if err != nil {
		return false, err
	}

	return RBACGroup(p) >= permission, nil
}

func (a *authorizationManagerImpl) WorkerHasPermissionForProject(ctx context.Context, project *queries.Project, worker *queries.Worker, permission RBACGroup) (bool, error) {
	return false, nil
}

func (a *authorizationManagerImpl) UserPermissionsChanged(ctx context.Context, user *queries.User) error {
	return nil
}

func (a *authorizationManagerImpl) WorkerPermissionsChanged(ctx context.Context, worker *queries.Worker) error {
	return nil
}
