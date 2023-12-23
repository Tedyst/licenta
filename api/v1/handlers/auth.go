package handlers

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"
	"github.com/tedyst/licenta/api/v1/generated"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/models"
)

func (server *serverHandler) PostLogin(ctx context.Context, request generated.PostLoginRequestObject) (generated.PostLoginResponseObject, error) {
	err := valid.Struct(request)
	if err != nil {
		return generated.PostLogin400JSONResponse{
			Message: err.Error(),
			Success: false,
		}, nil
	}

	user, err := server.Queries.GetUserByUsernameOrEmail(ctx, request.Body.Username)
	if err != nil {
		traceError(ctx, errors.Wrap(err, "PostLogin: error getting user"))
		return generated.PostLogin401JSONResponse{
			Message: InvalidCredentials,
			Success: false,
		}, nil
	}

	ok, err := models.VerifyPassword(ctx, user, request.Body.Password)
	if err != nil {
		return nil, errors.Wrapf(err, "PostLogin: error verifying password for user `%s`", request.Body.Username)
	}

	if !ok {
		return generated.PostLogin401JSONResponse{
			Message: InvalidCredentials,
			Success: false,
		}, nil
	}

	totpSecretToken, err := server.Queries.GetTOTPSecretForUser(ctx, user.ID)
	if err != nil && err != sql.ErrNoRows {
		return nil, errors.Wrap(err, "PostLogin: error getting totp secret for user")
	}
	if totpSecretToken != nil {
		res := models.VerifyTOTP(ctx, totpSecretToken, *request.Body.Totp)
		if !res {
			return generated.PostLogin401JSONResponse{
				Message: ErrInvalidTOTP,
				Success: false,
			}, nil
		}
	}

	server.SessionStore.SetUser(ctx, user)

	return generated.PostLogin200JSONResponse{
		Success: true,
		User: generated.User{
			Id:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		},
	}, nil
}

func (server *serverHandler) PostLogout(ctx context.Context, _ generated.PostLogoutRequestObject) (generated.PostLogoutResponseObject, error) {
	server.SessionStore.ClearSession(ctx)

	return generated.PostLogout200JSONResponse{
		Success: true,
	}, nil
}

func (server *serverHandler) PostRegister(ctx context.Context, request generated.PostRegisterRequestObject) (generated.PostRegisterResponseObject, error) {
	err := valid.Struct(request)
	if err != nil {
		return generated.PostRegister400JSONResponse{
			Message: err.Error(),
			Success: false,
		}, nil
	}

	user, err := server.Queries.CreateUser(ctx, queries.CreateUserParams{
		Username: request.Body.Username,
		Email:    request.Body.Email,
	})
	if err != nil {
		return nil, errors.Wrap(err, "PostRegister: error creating user")
	}

	err = models.SetPassword(ctx, user, request.Body.Password)
	if err != nil {
		return nil, errors.Wrap(err, "PostRegister: error setting password")
	}

	err = server.Queries.UpdateUserPassword(ctx, queries.UpdateUserPasswordParams{
		ID:       user.ID,
		Password: user.Password,
	})

	if err != nil {
		return nil, errors.Wrap(err, "PostRegister: error updating user password")
	}

	server.SessionStore.SetUser(ctx, user)

	return generated.PostRegister200JSONResponse{
		Success: true,
		User: generated.User{
			Id:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		},
	}, nil
}

func (server *serverHandler) Post2faTotpFirstStep(ctx context.Context, request generated.Post2faTotpFirstStepRequestObject) (generated.Post2faTotpFirstStepResponseObject, error) {
	user := server.SessionStore.GetUser(ctx)
	if user == nil {
		return generated.Post2faTotpFirstStep401JSONResponse{
			Message: Unauthorized,
			Success: false,
		}, nil
	}

	key, err := models.GenerateTOTPSecret(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "Post2faTotpFirstStep: error generating totp key")
	}

	totpSecret, err := server.Queries.CreateTOTPSecretForUser(ctx, queries.CreateTOTPSecretForUserParams{
		UserID:     user.ID,
		TotpSecret: key,
		Valid:      false,
	})
	if err != nil {
		return nil, errors.Wrap(err, "Post2faTotpFirstStep: error creating totp secret for user")
	}

	return generated.Post2faTotpFirstStep200JSONResponse{
		TotpSecret: totpSecret.TotpSecret,
	}, nil
}

func (server *serverHandler) Post2faTotpSecondStep(ctx context.Context, request generated.Post2faTotpSecondStepRequestObject) (generated.Post2faTotpSecondStepResponseObject, error) {
	err := valid.Struct(request)
	if err != nil {
		return generated.Post2faTotpSecondStep400JSONResponse{
			Message: err.Error(),
			Success: false,
		}, nil
	}
	user := server.SessionStore.GetUser(ctx)
	if user == nil {
		return generated.Post2faTotpSecondStep401JSONResponse{
			Message: Unauthorized,
			Success: false,
		}, nil
	}

	totpSecret, err := server.Queries.GetInvalidTOTPSecretForUser(ctx, user.ID)
	if err != nil && err != sql.ErrNoRows {
		return nil, errors.Wrap(err, "Post2faTotpSecondStep: error getting invalid totp secret for user")
	}

	if totpSecret == nil {
		return generated.Post2faTotpSecondStep401JSONResponse{
			Message: ErrNotTryingToSetupTOTP,
			Success: false,
		}, nil
	}

	if models.VerifyTOTP(ctx, totpSecret, request.Body.TotpCode) {
		err = server.Queries.ValidateTOTPSecretForUser(ctx, user.ID)
		if err != nil {
			return nil, errors.Wrap(err, "Post2faTotpSecondStep: error updating totp secret for user")
		}
		return generated.Post2faTotpSecondStep200JSONResponse{
			Success: true,
		}, nil
	}

	return generated.Post2faTotpSecondStep401JSONResponse{
		Message: ErrInvalidTOTP,
		Success: false,
	}, nil
}
