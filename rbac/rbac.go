package rbac

import (
	"context"
	"database/sql"

	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/models"
)

type RBACGroup int16

const (
	Owner RBACGroup = iota
	Admin
	Viewer
	None
)

type RBACAction int8

const (
	CreateProject RBACAction = iota
	ReadProject
	UpdateProject
	DeleteProject
	PromoteToOwner
	PromoteToAdmin
	InviteMember
	RemoveMember
	RunProjectScan
)

type rbacImpl struct {
	database db.TransactionQuerier
}

func NewRBAC(database db.TransactionQuerier) *rbacImpl {
	return &rbacImpl{database: database}
}

func (r *rbacImpl) GetPermissionsForOrganization(ctx context.Context, org *models.Organization, user *models.User) (RBACGroup, error) {
	permission, err := r.database.GetOrganizationPermissionsForUser(ctx, queries.GetOrganizationPermissionsForUserParams{
		OrganizationID: org.ID,
		UserID:         user.ID,
	})
	if err != nil && err != sql.ErrNoRows {
		return None, err
	}

	return RBACGroup(permission), nil
}

func (r *rbacImpl) GetPermissionsForProject(ctx context.Context, project *models.Project, user *models.User) (RBACGroup, error) {
	permission, err := r.database.GetProjectPermissionsForUser(ctx, queries.GetProjectPermissionsForUserParams{
		ProjectID:      project.ID,
		UserID:         user.ID,
		OrganizationID: project.OrganizationID,
	})
	if err != nil && err != sql.ErrNoRows {
		return None, err
	}

	return RBACGroup(permission), nil
}

func (r *rbacImpl) HasPermission(group RBACGroup, action RBACAction) bool {
	switch action {
	case CreateProject:
		return group == Owner || group == Admin
	case ReadProject:
		return group == Owner || group == Admin || group == Viewer
	case UpdateProject:
		return group == Owner || group == Admin
	case DeleteProject:
		return group == Owner || group == Admin
	case PromoteToOwner:
		return group == Owner
	case PromoteToAdmin:
		return group == Owner || group == Admin
	case InviteMember:
		return group == Owner || group == Admin
	case RemoveMember:
		return group == Owner || group == Admin
	case RunProjectScan:
		return group == Owner || group == Admin || group == Viewer
	default:
		return false
	}
}
