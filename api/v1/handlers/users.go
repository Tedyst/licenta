package handlers

import (
	"context"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/tedyst/licenta/api/v1/generated"
	"github.com/tedyst/licenta/db/queries"
)

func (server *serverHandler) GetUsers(ctx context.Context, request generated.GetUsersRequestObject) (generated.GetUsersResponseObject, error) {
	user, err := server.userAuth.GetUser(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "GetUsers: error getting user")
	}
	if user == nil {
		return generated.GetUsers401JSONResponse{
			Message: Unauthorized,
			Success: false,
		}, nil
	}

	count, err := server.DatabaseProvider.CountUsers(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "GetUsers: error cou7nting users")
	}

	var limit int32 = DefaultPaginationLimit
	var offset int32 = 0
	if request.Params.Limit != nil {
		limit = *request.Params.Limit
	}
	if request.Params.Offset != nil {
		offset = *request.Params.Offset
	}

	users, err := server.DatabaseProvider.ListUsersPaginated(ctx, queries.ListUsersPaginatedParams{
		StartID:      0,
		ItemsPerPage: limit,
	})
	if err != nil {
		return nil, errors.Wrap(err, "GetUsers: error getting users")
	}

	result := make([]generated.User, len(users))
	for i, user := range users {
		result[i] = generated.User{
			Id:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		}
	}

	var nextURL = viper.GetString("baseurl") + Prefix + "/users?limit=" + string(limit) + "&offset=" + string(offset+limit)
	if int(offset+limit) > int(count) {
		nextURL = ""
	}

	return generated.GetUsers200JSONResponse{
		Count:   int(count),
		Next:    &nextURL,
		Results: result,
	}, nil
}

func (server *serverHandler) GetUsersMe(ctx context.Context, request generated.GetUsersMeRequestObject) (generated.GetUsersMeResponseObject, error) {
	user, err := server.userAuth.GetUser(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "GetUsersMe: error getting user")
	}
	if user == nil {
		return generated.GetUsersMe401JSONResponse{
			Message: Unauthorized,
			Success: false,
		}, nil
	}

	return generated.GetUsersMe200JSONResponse{
		Success: true,
		User: generated.User{
			Id:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		},
	}, nil
}

func (server *serverHandler) GetUsersId(ctx context.Context, request generated.GetUsersIdRequestObject) (generated.GetUsersIdResponseObject, error) {
	user, err := server.userAuth.GetUser(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "GetUsersId: error getting user")
	}
	if user == nil {
		return generated.GetUsersId401JSONResponse{
			Message: Unauthorized,
			Success: false,
		}, nil
	}

	u, err := server.DatabaseProvider.GetUser(ctx, request.Id)
	if err != nil {
		return nil, errors.Wrap(err, "GetUsersId: error getting user")
	}

	return generated.GetUsersId200JSONResponse{
		Id:       u.ID,
		Username: u.Username,
		Email:    u.Email,
	}, nil
}

func (server *serverHandler) PostUsersMeChangePassword(ctx context.Context, request generated.PostUsersMeChangePasswordRequestObject) (generated.PostUsersMeChangePasswordResponseObject, error) {
	err := valid.Struct(request)
	if err != nil {
		return nil, errors.Wrap(err, "PostUsersMeChangePassword: error validating request")
	}

	user, err := server.userAuth.GetUser(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "PostUsersMeChangePassword: error getting user")
	}
	if user == nil {
		return generated.PostUsersMeChangePassword401JSONResponse{
			Message: Unauthorized,
			Success: false,
		}, nil
	}

	_, err = server.userAuth.VerifyPassword(ctx, user, request.Body.OldPassword)
	if err != nil {
		return generated.PostUsersMeChangePassword401JSONResponse{
			Message: "Old password is incorrect",
			Success: false,
		}, nil
	}

	err = server.userAuth.UpdatePassword(ctx, user, request.Body.NewPassword)
	if err != nil {
		return nil, errors.Wrap(err, "PostUsersMeChangePassword: error setting password")
	}

	return generated.PostUsersMeChangePassword200JSONResponse{
		Success: true,
	}, nil
}
