package handlers

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
	"github.com/tedyst/licenta/api/authorization"
	"github.com/tedyst/licenta/api/v1/generated"
	"github.com/tedyst/licenta/db/queries"
)

func (server *serverHandler) GetOrganizations(ctx context.Context, request generated.GetOrganizationsRequestObject) (generated.GetOrganizationsResponseObject, error) {
	user, err := server.userAuth.GetUser(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "error getting user")
	}

	organization, err := server.DatabaseProvider.GetOrganizationsForUser(ctx, user.ID)
	if err != nil {
		return nil, errors.Wrap(err, "error getting organizations")
	}

	projects, err := server.DatabaseProvider.GetAllOrganizationProjectsForUser(ctx, user.ID)
	if err != nil {
		return nil, errors.Wrap(err, "error getting organizations")
	}

	organizationProjects := make(map[int64][]generated.Project)
	for _, project := range projects {
		organizationProjects[project.OrganizationID] = append(organizationProjects[project.OrganizationID], generated.Project{
			Id:             int64(project.ID),
			CreatedAt:      project.CreatedAt.Time.Format("2006-01-02T15:04:05.000Z"),
			Name:           project.Name,
			OrganizationId: int64(project.OrganizationID),
			Remote:         project.Remote,
		})
	}

	response := generated.GetOrganizations200JSONResponse{
		Organizations: make([]generated.Organization, len(organization)),
		Success:       true,
	}
	for _, org := range organization {
		response.Organizations = append(response.Organizations, generated.Organization{
			Id:       int64(org.ID),
			Name:     org.Name,
			Projects: organizationProjects[int64(org.ID)],
		})
	}

	return &response, nil
}

func (server *serverHandler) GetOrganizationsId(ctx context.Context, request generated.GetOrganizationsIdRequestObject) (generated.GetOrganizationsIdResponseObject, error) {
	user, err := server.userAuth.GetUser(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "error getting user")
	}

	organization, err := server.DatabaseProvider.GetOrganization(ctx, request.Id)
	if err != nil && err != pgx.ErrNoRows {
		return nil, errors.Wrap(err, "error getting organization")
	}

	if organization == nil {
		return &generated.GetOrganizationsId404JSONResponse{
			Success: false,
			Message: "Organization not found",
		}, nil
	}

	hasPerm, err := server.authorization.UserHasPermissionForOrganization(ctx, organization, user, authorization.Viewer)
	if err != nil {
		return nil, errors.Wrap(err, "error checking permissions")
	}
	if !hasPerm {
		return &generated.GetOrganizationsId404JSONResponse{
			Success: false,
			Message: "Organization not found",
		}, nil
	}

	projects, err := server.DatabaseProvider.GetOrganizationProjects(ctx, organization.ID)
	if err != nil {
		return nil, errors.Wrap(err, "error getting projects")
	}

	response := generated.GetOrganizationsId200JSONResponse{
		Organization: generated.Organization{
			Id:       int64(organization.ID),
			Name:     organization.Name,
			Projects: make([]generated.Project, len(projects)),
		},
		Success: true,
	}

	for i, project := range projects {
		response.Organization.Projects[i] = generated.Project{
			Id:             int64(project.ID),
			CreatedAt:      project.CreatedAt.Time.Format("2006-01-02T15:04:05.000Z"),
			Name:           project.Name,
			OrganizationId: int64(project.OrganizationID),
			Remote:         project.Remote,
		}
	}

	return &response, nil
}

func (server *serverHandler) DeleteOrganizationsId(ctx context.Context, request generated.DeleteOrganizationsIdRequestObject) (generated.DeleteOrganizationsIdResponseObject, error) {
	user, err := server.userAuth.GetUser(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "error getting user")
	}

	organization, err := server.DatabaseProvider.GetOrganization(ctx, request.Id)
	if err != nil && err != pgx.ErrNoRows {
		return nil, errors.Wrap(err, "error getting organization")
	}

	if organization == nil {
		return &generated.DeleteOrganizationsId404JSONResponse{
			Success: false,
			Message: "Organization not found",
		}, nil
	}

	hasPerm, err := server.authorization.UserHasPermissionForOrganization(ctx, organization, user, authorization.Admin)
	if err != nil {
		return nil, errors.Wrap(err, "error checking permissions")
	}
	if !hasPerm {
		hasPermViewer, err := server.authorization.UserHasPermissionForOrganization(ctx, organization, user, authorization.Viewer)
		if err != nil {
			return nil, errors.Wrap(err, "error checking permissions")
		}
		if hasPermViewer {
			return &generated.DeleteOrganizationsId401JSONResponse{
				Success: false,
				Message: "Forbidden",
			}, nil
		}
		return &generated.DeleteOrganizationsId404JSONResponse{
			Success: false,
			Message: "Organization not found",
		}, nil
	}

	err = server.DatabaseProvider.DeleteOrganization(ctx, organization.ID)
	if err != nil {
		return nil, errors.Wrap(err, "error deleting organization")
	}

	return &generated.DeleteOrganizationsId204JSONResponse{
		Success: true,
	}, nil
}

func (server *serverHandler) PostOrganizations(ctx context.Context, request generated.PostOrganizationsRequestObject) (generated.PostOrganizationsResponseObject, error) {
	user, err := server.userAuth.GetUser(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "error getting user")
	}

	organization, err := server.DatabaseProvider.CreateOrganization(ctx, request.Body.Name)
	if err != nil {
		return nil, errors.Wrap(err, "error creating organization")
	}

	err = server.DatabaseProvider.AddUserToOrganization(ctx, queries.AddUserToOrganizationParams{
		OrganizationID: organization.ID,
		UserID:         user.ID,
		Role:           int32(authorization.Owner),
	})
	if err != nil {
		return nil, errors.Wrap(err, "error adding user to organization")
	}

	return &generated.PostOrganizations201JSONResponse{
		Organization: generated.Organization{
			Id:   int64(organization.ID),
			Name: organization.Name,
		},
		Success: true,
	}, nil
}
