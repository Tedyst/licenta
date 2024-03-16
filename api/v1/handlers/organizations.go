package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/tedyst/licenta/api/authorization"
	"github.com/tedyst/licenta/api/v1/generated"
	"github.com/tedyst/licenta/db/queries"
)

func (server *serverHandler) GetOrganizations(ctx context.Context, request generated.GetOrganizationsRequestObject) (generated.GetOrganizationsResponseObject, error) {
	user, err := server.userAuth.GetUser(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}
	if user == nil {
		return generated.GetOrganizations401JSONResponse{
			Message: "Unauthorized",
			Success: false,
		}, nil
	}

	organization, err := server.DatabaseProvider.GetOrganizationsByUser(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("error getting organizations: %w", err)
	}

	projects, err := server.DatabaseProvider.GetAllOrganizationProjectsForUser(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("error getting organizations: %w", err)
	}

	organizationProjects := map[int64][]generated.Project{}
	for _, project := range projects {
		organizationProjects[project.OrganizationID] = append(organizationProjects[project.OrganizationID], generated.Project{
			Id:             int64(project.ID),
			CreatedAt:      project.CreatedAt.Time.Format(time.RFC3339Nano),
			Name:           project.Name,
			OrganizationId: int64(project.OrganizationID),
			Remote:         project.Remote,
		})
	}

	response := generated.GetOrganizations200JSONResponse{
		Organizations: []generated.Organization{},
		Success:       true,
	}
	for _, org := range organization {
		val, ok := organizationProjects[int64(org.Organization.ID)]
		if !ok {
			val = []generated.Project{}
		}
		response.Organizations = append(response.Organizations, generated.Organization{
			Id:       int64(org.Organization.ID),
			Name:     org.Organization.Name,
			Projects: val,
			Stats: generated.OrganizationStats{
				FailedScans: int(org.MaximumSeverity),
				Projects:    int(org.Projects),
				Scans:       int(org.Scans),
				Users:       int(org.Users),
			},
		})
	}

	return &response, nil
}

func (server *serverHandler) checkForOrganizationPermission(ctx context.Context, organizationID int64, role authorization.RBACGroup) (*queries.User, *queries.Organization, bool, bool, error) {
	user, err := server.userAuth.GetUser(ctx)
	if err != nil {
		return user, nil, false, false, fmt.Errorf("error getting user: %w", err)
	}
	if user == nil {
		return nil, nil, false, false, nil
	}

	organization, err := server.DatabaseProvider.GetOrganization(ctx, organizationID)
	if err != nil && err != pgx.ErrNoRows {
		return user, nil, false, false, fmt.Errorf("error getting organization: %w", err)
	}

	if organization == nil {
		return user, nil, false, false, nil
	}

	hasPerm, err := server.authorization.UserHasPermissionForOrganization(ctx, organization, user, role)
	if err != nil {
		return user, nil, false, false, fmt.Errorf("error checking permissions: %w", err)
	}
	if !hasPerm {
		hasViewPerm, err := server.authorization.UserHasPermissionForOrganization(ctx, organization, user, authorization.Viewer)
		if err != nil {
			return user, nil, false, false, fmt.Errorf("error checking permissions: %w", err)
		}
		return user, organization, false, hasViewPerm, nil
	}

	return user, organization, true, true, nil
}

func (server *serverHandler) GetOrganizationsId(ctx context.Context, request generated.GetOrganizationsIdRequestObject) (generated.GetOrganizationsIdResponseObject, error) {
	user, organization, hasPerm, hasViewPerm, err := server.checkForOrganizationPermission(ctx, request.Id, authorization.Admin)
	if err != nil {
		return nil, fmt.Errorf("error checking permissions: %w", err)
	}
	if user == nil {
		return generated.GetOrganizationsId401JSONResponse{
			Message: "Unauthorized",
			Success: false,
		}, nil
	}
	if organization == nil {
		return &generated.GetOrganizationsId401JSONResponse{
			Success: false,
			Message: "Unauthorized",
		}, nil
	}
	if !hasPerm {
		if hasViewPerm {
			return &generated.GetOrganizationsId401JSONResponse{
				Success: false,
				Message: "Forbidden",
			}, nil
		}
		return &generated.GetOrganizationsId404JSONResponse{
			Success: false,
			Message: "Organization not found",
		}, nil
	}

	projects, err := server.DatabaseProvider.GetOrganizationProjects(ctx, organization.ID)
	if err != nil {
		return nil, fmt.Errorf("error getting projects: %w", err)
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
			CreatedAt:      project.CreatedAt.Time.Format(time.RFC3339Nano),
			Name:           project.Name,
			OrganizationId: int64(project.OrganizationID),
			Remote:         project.Remote,
		}
	}

	return &response, nil
}

func (server *serverHandler) DeleteOrganizationsId(ctx context.Context, request generated.DeleteOrganizationsIdRequestObject) (generated.DeleteOrganizationsIdResponseObject, error) {
	user, organization, hasPerm, hasViewPerm, err := server.checkForOrganizationPermission(ctx, request.Id, authorization.Admin)
	if err != nil {
		return nil, fmt.Errorf("error checking permissions: %w", err)
	}
	if user == nil {
		return generated.DeleteOrganizationsId401JSONResponse{
			Message: "Unauthorized",
			Success: false,
		}, nil
	}
	if organization == nil {
		return &generated.DeleteOrganizationsId401JSONResponse{
			Success: false,
			Message: "Unauthorized",
		}, nil
	}
	if !hasPerm {
		if hasViewPerm {
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
		return nil, fmt.Errorf("error deleting organization: %w", err)
	}

	return &generated.DeleteOrganizationsId204JSONResponse{
		Success: true,
	}, nil
}

func (server *serverHandler) PostOrganizations(ctx context.Context, request generated.PostOrganizationsRequestObject) (generated.PostOrganizationsResponseObject, error) {
	err := valid.Struct(request)
	if err != nil {
		return generated.PostOrganizations400JSONResponse{
			Success: false,
			Message: "Validation error: " + err.Error(),
		}, nil
	}

	user, err := server.userAuth.GetUser(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}
	if user == nil {
		return generated.PostOrganizations401JSONResponse{
			Message: "Unauthorized",
			Success: false,
		}, nil
	}

	_, err = server.DatabaseProvider.GetOrganizationByName(ctx, request.Body.Name)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("error getting organization: %w", err)
	}
	if err == nil {
		return &generated.PostOrganizations400JSONResponse{
			Success: false,
			Message: "Organization already exists",
		}, nil
	}

	organization, err := server.DatabaseProvider.CreateOrganization(ctx, request.Body.Name)
	if err != nil {
		return nil, fmt.Errorf("error creating organization: %w", err)
	}

	err = server.DatabaseProvider.AddUserToOrganization(ctx, queries.AddUserToOrganizationParams{
		OrganizationID: organization.ID,
		UserID:         user.ID,
		Role:           int32(authorization.Owner),
	})
	if err != nil {
		return nil, fmt.Errorf("error adding user to organization: %w", err)
	}

	return &generated.PostOrganizations201JSONResponse{
		Organization: generated.Organization{
			Id:   int64(organization.ID),
			Name: organization.Name,
		},
		Success: true,
	}, nil
}
