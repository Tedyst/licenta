package handlers

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/tedyst/licenta/api/v1/generated"
	database "github.com/tedyst/licenta/db"
	db "github.com/tedyst/licenta/db/generated"
	"github.com/tedyst/licenta/middleware/session"
	"github.com/tedyst/licenta/models"
)

func (*ServerHandler) GetUsers(ctx context.Context, request generated.GetUsersRequestObject) (generated.GetUsersResponseObject, error) {
	user, err := session.GetUser(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "GetUsers: error getting user")
	}
	if user == nil || !user.Admin {
		return generated.GetUsers401JSONResponse{
			Message: Unauthorized,
			Success: false,
		}, nil
	}

	count, err := database.DatabaseQueries.CountUsers(ctx)
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

	users, err := database.DatabaseQueries.ListUsersPaginated(ctx, db.ListUsersPaginatedParams{
		Limit:  limit,
		Offset: offset,
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
	var previousURL = viper.GetString("baseurl") + Prefix + "/users?limit=" + string(limit) + "&offset=" + string(offset-limit)
	if int(offset-limit) < 0 {
		previousURL = ""
	}

	return generated.GetUsers200JSONResponse{
		Count:    int(count),
		Next:     nextURL,
		Previous: previousURL,
		Results:  result,
	}, nil
}

func (*ServerHandler) GetUsersMe(ctx context.Context, request generated.GetUsersMeRequestObject) (generated.GetUsersMeResponseObject, error) {
	user, err := session.GetUser(ctx)
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

func (*ServerHandler) GetUsersId(ctx context.Context, request generated.GetUsersIdRequestObject) (generated.GetUsersIdResponseObject, error) {
	user, err := session.GetUser(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "GetUsersId: error getting user")
	}
	if user == nil || !user.Admin {
		return generated.GetUsersId401JSONResponse{
			Message: Unauthorized,
			Success: false,
		}, nil
	}

	u, err := database.DatabaseQueries.GetUser(ctx, request.Id)
	if err != nil {
		return nil, errors.Wrap(err, "GetUsersId: error getting user")
	}

	return generated.GetUsersId200JSONResponse{
		Id:       u.ID,
		Username: u.Username,
		Email:    u.Email,
	}, nil
}

func (*ServerHandler) PostUsersMeChangePassword(ctx context.Context, request generated.PostUsersMeChangePasswordRequestObject) (generated.PostUsersMeChangePasswordResponseObject, error) {
	err := valid.Struct(request)
	if err != nil {
		return nil, errors.Wrap(err, "PostUsersMeChangePassword: error validating request")
	}

	user, err := session.GetUser(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "PostUsersMeChangePassword: error getting user")
	}
	if user == nil {
		return generated.PostUsersMeChangePassword401JSONResponse{
			Message: Unauthorized,
			Success: false,
		}, nil
	}

	ok, err := models.VerifyPassword(ctx, user, request.Body.OldPassword)
	if err != nil {
		return nil, errors.Wrap(err, "PostUsersMeChangePassword: error verifying password")
	}
	if !ok {
		return generated.PostUsersMeChangePassword401JSONResponse{
			Message: InvalidCredentials,
			Success: false,
		}, nil
	}

	err = models.SetPassword(ctx, user, request.Body.NewPassword)
	if err != nil {
		return nil, errors.Wrap(err, "PostUsersMeChangePassword: error setting password")
	}

	err = database.DatabaseQueries.UpdateUser(ctx, db.UpdateUserParams{
		ID:       user.ID,
		Password: sql.NullString{Valid: true, String: user.Password},
	})

	return generated.PostUsersMeChangePassword200JSONResponse{
		Success: true,
	}, nil
}
