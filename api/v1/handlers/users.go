package handlers

import (
	"context"

	"github.com/tedyst/licenta/api/v1/generated"
)

func (*ServerHandler) GetUsers(ctx context.Context, request generated.GetUsersRequestObject) (generated.GetUsersResponseObject, error) {
	return nil, nil
}

func (*ServerHandler) GetUsersMe(ctx context.Context, request generated.GetUsersMeRequestObject) (generated.GetUsersMeResponseObject, error) {
	return nil, nil
}
