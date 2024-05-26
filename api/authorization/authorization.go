package authorization

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/tedyst/licenta/cache"
	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/db/queries"
)

type RBACGroup int16

const (
	Owner RBACGroup = iota
	Admin
	Viewer
	None
	Worker
)

func (r RBACGroup) String() string {
	switch r {
	case Owner:
		return "Owner"
	case Admin:
		return "Admin"
	case Viewer:
		return "Viewer"
	case None:
		return "None"
	case Worker:
		return "Worker"
	default:
		return "Unknown"
	}
}

func hasPermission(required RBACGroup, actual RBACGroup) bool {
	switch required {
	case Owner:
		return actual == Owner
	case Admin:
		return actual == Owner || actual == Admin
	case Viewer:
		return actual == Owner || actual == Admin || actual == Viewer
	case None:
		return actual == None
	case Worker:
		return actual == Worker
	default:
		return false
	}
}

type AuthorizationManager interface {
	UserHasPermissionForOrganization(ctx context.Context, organization *queries.Organization, user *queries.User, permission RBACGroup) (bool, error)
	UserHasPermissionForProject(ctx context.Context, project *queries.Project, user *queries.User, permission RBACGroup) (bool, error)

	WorkerHasPermissionForProject(ctx context.Context, project *queries.Project, worker *queries.Worker, permission RBACGroup) (bool, error)

	UserPermissionsChanged(ctx context.Context, user *queries.User) error
	WorkerPermissionsChanged(ctx context.Context, worker *queries.Worker) error
}

func cacheKeyForWorkerOrganization(worker *queries.Worker, organization *queries.Organization) string {
	return "worker:" + worker.Token + ":organization:" + fmt.Sprint(organization.ID)
}

func cacheKeyForWorkerProject(worker *queries.Worker, project *queries.Project) string {
	return "worker:" + worker.Token + ":project:" + fmt.Sprint(project.ID)
}

func cacheKeyForUserOrganization(user *queries.User, organization *queries.Organization) string {
	return "user:" + fmt.Sprint(user.ID) + ":organization:" + fmt.Sprint(organization.ID)
}

func cacheKeyForUserProject(user *queries.User, project *queries.Project) string {
	return "user:" + fmt.Sprint(user.ID) + ":project:" + fmt.Sprint(project.ID)
}

func cachePatternForUser(user *queries.User) string {
	return "user:" + fmt.Sprint(user.ID) + ":.*"
}

func cachePatternForWorker(worker *queries.Worker) string {
	return "worker:" + worker.Token + ":.*"
}

type authorizationManagerImpl struct {
	cache   cache.CacheProvider[int16]
	querier db.TransactionQuerier
}

func NewAuthorizationManager(querier db.TransactionQuerier, cache cache.CacheProvider[int16]) AuthorizationManager {
	return &authorizationManagerImpl{querier: querier, cache: cache}
}

func (a *authorizationManagerImpl) UserHasPermissionForOrganization(ctx context.Context, organization *queries.Organization, user *queries.User, permission RBACGroup) (bool, error) {
	cached, ok, err := a.cache.Get(cacheKeyForUserOrganization(user, organization))
	if err != nil {
		return false, err
	}
	if ok {
		return hasPermission(permission, RBACGroup(cached)), nil
	}

	p, err := a.querier.GetOrganizationPermissionForUser(ctx, queries.GetOrganizationPermissionForUserParams{
		OrganizationID: organization.ID,
		UserID:         user.ID,
	})
	if err != nil && err != pgx.ErrNoRows {
		return false, err
	}
	if err == pgx.ErrNoRows {
		return false, a.cache.Set(cacheKeyForUserOrganization(user, organization), int16(None))
	}

	return hasPermission(permission, RBACGroup(p)), a.cache.Set(cacheKeyForUserOrganization(user, organization), int16(p))
}

func (a *authorizationManagerImpl) UserHasPermissionForProject(ctx context.Context, project *queries.Project, user *queries.User, permission RBACGroup) (bool, error) {
	cached, ok, err := a.cache.Get(cacheKeyForUserProject(user, project))
	if err != nil {
		return false, err
	}
	if ok {
		return hasPermission(permission, RBACGroup(cached)), nil
	}

	p, err := a.querier.GetOrganizationPermissionForUser(ctx, queries.GetOrganizationPermissionForUserParams{
		OrganizationID: project.OrganizationID,
		UserID:         user.ID,
	})
	if err != nil && err != pgx.ErrNoRows {
		return false, err
	}
	if err == pgx.ErrNoRows {
		return false, a.cache.Set(cacheKeyForUserProject(user, project), int16(None))
	}

	return hasPermission(permission, RBACGroup(p)), a.cache.Set(cacheKeyForUserProject(user, project), int16(p))
}

func (a *authorizationManagerImpl) WorkerHasPermissionForProject(ctx context.Context, project *queries.Project, worker *queries.Worker, permission RBACGroup) (bool, error) {
	cached, ok, err := a.cache.Get(cacheKeyForWorkerProject(worker, project))
	if err != nil {
		return false, err
	}
	if ok {
		return hasPermission(permission, RBACGroup(cached)), nil
	}

	_, err = a.querier.GetWorkerForProject(ctx, queries.GetWorkerForProjectParams{
		ProjectID: project.ID,
		Token:     worker.Token,
	})
	if err != nil {
		return false, err
	}

	return true, nil
}

func (a *authorizationManagerImpl) UserPermissionsChanged(ctx context.Context, user *queries.User) error {
	return a.cache.Invalidate(cachePatternForUser(user))
}

func (a *authorizationManagerImpl) WorkerPermissionsChanged(ctx context.Context, worker *queries.Worker) error {
	return a.cache.Invalidate(cachePatternForWorker(worker))
}
