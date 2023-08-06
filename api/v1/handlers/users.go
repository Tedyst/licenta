package handlers

import (
	"context"

	"github.com/pkg/errors"
	"github.com/tedyst/licenta/api/v1/generated"
	"github.com/tedyst/licenta/config"
	db "github.com/tedyst/licenta/db/generated"
	"github.com/tedyst/licenta/middleware/session"
)

func (*ServerHandler) GetUsers(ctx context.Context, request generated.GetUsersRequestObject) (generated.GetUsersResponseObject, error) {
	_, err := session.GetUser(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "GetUsers: error getting user")
	}
	// if user == nil || !user.Admin {
	// 	return generated.GetUsers401JSONResponse{
	// 		Message: Unauthorized,
	// 		Success: false,
	// 	}, nil
	// }

	count, err := config.DatabaseQueries.CountUsers(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "GetUsers: error counting users")
	}

	var limit int32 = DefaultPaginationLimit
	var offset int32 = 0
	if request.Params.Limit != nil {
		limit = *request.Params.Limit
	}
	if request.Params.Offset != nil {
		offset = *request.Params.Offset
	}

	users, err := config.DatabaseQueries.ListUsersPaginated(ctx, db.ListUsersPaginatedParams{
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

	var text = "asd"
	var intCount = int(count)
	return generated.GetUsers200JSONResponse{
		Count:    &intCount,
		Next:     &text,
		Previous: &text,
		Results:  &result,
	}, nil
}

func (*ServerHandler) GetUsersMe(ctx context.Context, request generated.GetUsersMeRequestObject) (generated.GetUsersMeResponseObject, error) {
	return nil, nil
}
