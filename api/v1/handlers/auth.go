package handlers

import (
	"context"

	"github.com/pkg/errors"
	"github.com/tedyst/licenta/api/v1/generated"
	database "github.com/tedyst/licenta/db"
	db "github.com/tedyst/licenta/db/generated"
	"github.com/tedyst/licenta/middleware/session"
	"github.com/tedyst/licenta/models"
)

func (*ServerHandler) PostLogin(ctx context.Context, request generated.PostLoginRequestObject) (generated.PostLoginResponseObject, error) {
	err := valid.Struct(request)
	if err != nil {
		return generated.PostLogin400JSONResponse{
			Message: err.Error(),
			Success: false,
		}, nil
	}

	user, err := database.DatabaseQueries.GetUserByUsernameOrEmail(ctx, request.Body.Username)
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

	if models.Requires2FA(ctx, user) {
		err = session.SetWaiting2FA(ctx, user)
		if err != nil {
			return nil, errors.Wrap(err, "PostLogin: error setting waiting2fa")
		}

		return generated.PostLogin401JSONResponse{
			Success: false,
			Message: TwoFactorRequired,
		}, nil
	}

	session.SetUser(ctx, user)

	return generated.PostLogin200JSONResponse{
		Success: true,
		User: generated.User{
			Id:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		},
	}, nil
}

func (*ServerHandler) PostLogout(ctx context.Context, _ generated.PostLogoutRequestObject) (generated.PostLogoutResponseObject, error) {
	session.ClearSession(ctx)

	return generated.PostLogout200JSONResponse{
		Success: true,
	}, nil

}

func (*ServerHandler) PostRegister(ctx context.Context, request generated.PostRegisterRequestObject) (generated.PostRegisterResponseObject, error) {
	err := valid.Struct(request)
	if err != nil {
		return generated.PostRegister400JSONResponse{
			Message: err.Error(),
			Success: false,
		}, nil
	}

	user, err := database.DatabaseQueries.CreateUser(ctx, db.CreateUserParams{
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

	err = database.DatabaseQueries.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		ID:       user.ID,
		Password: user.Password,
	})

	if err != nil {
		return nil, errors.Wrap(err, "PostRegister: error updating user password")
	}

	err = session.SetUser(ctx, user)
	if err != nil {
		return nil, errors.Wrap(err, "PostRegister: error setting user")
	}

	return generated.PostRegister200JSONResponse{
		Success: true,
		User: generated.User{
			Id:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		},
	}, nil
}

func (*ServerHandler) Post2faTotpFirstStep(ctx context.Context, request generated.Post2faTotpFirstStepRequestObject) (generated.Post2faTotpFirstStepResponseObject, error) {
	user, err := session.GetUser(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "Post2faTotpFirstStep: error getting user")
	}

	if user == nil {
		return generated.Post2faTotpFirstStep401JSONResponse{
			Message: Unauthorized,
			Success: false,
		}, nil
	}

	key, err := models.GenerateTOTP(ctx, user)
	if err != nil {
		return nil, errors.Wrap(err, "Post2faTotpFirstStep: error generating totp key")
	}

	err = session.SetTOTPKey(ctx, key)
	if err != nil {
		return nil, errors.Wrap(err, "Post2faTotpFirstStep: error setting totp key")
	}

	return generated.Post2faTotpFirstStep200JSONResponse{
		TotpSecret: key,
	}, nil
}

func (*ServerHandler) Post2faTotpSecondStep(ctx context.Context, request generated.Post2faTotpSecondStepRequestObject) (generated.Post2faTotpSecondStepResponseObject, error) {
	err := valid.Struct(request)
	if err != nil {
		return generated.Post2faTotpSecondStep400JSONResponse{
			Message: err.Error(),
			Success: false,
		}, nil
	}

	user, err := session.GetWaiting2FA(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "Post2faTotpSecondStep: error getting waiting 2fa user")
	}

	if user == nil {
		return generated.Post2faTotpSecondStep401JSONResponse{
			Message: Unauthorized,
			Success: false,
		}, nil
	}

	ok := models.Verify2FA(ctx, user, request.Body.TotpCode)
	if !ok {
		return generated.Post2faTotpSecondStep401JSONResponse{
			Message: InvalidCredentials,
			Success: false,
		}, nil
	}

	err = session.SetUser(ctx, user)
	if err != nil {
		return nil, errors.Wrap(err, "Post2faTotpSecondStep: error setting user")
	}

	return generated.Post2faTotpSecondStep200JSONResponse{
		Success: true,
		User: generated.User{
			Id:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		},
	}, nil
}

func (*ServerHandler) PostLogin2faTotp(ctx context.Context, request generated.PostLogin2faTotpRequestObject) (generated.PostLogin2faTotpResponseObject, error) {
	err := valid.Struct(request)
	if err != nil {
		return generated.PostLogin2faTotp400JSONResponse{
			Message: err.Error(),
			Success: false,
		}, nil
	}

	user, err := session.GetWaiting2FA(ctx)
	if err != nil {
		traceError(ctx, errors.Wrap(err, "PostLogin2faTotp: error getting user"))
		return generated.PostLogin2faTotp401JSONResponse{
			Message: InvalidCredentials,
			Success: false,
		}, nil
	}

	ok := models.VerifyTOTP(ctx, user, request.Body.TotpCode)
	if !ok {
		return generated.PostLogin2faTotp401JSONResponse{
			Message: InvalidCredentials,
			Success: false,
		}, nil
	}

	session.SetUser(ctx, user)

	return generated.PostLogin2faTotp200JSONResponse{
		Success: true,
		User: generated.User{
			Id:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		},
	}, nil
}
