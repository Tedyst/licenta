package handlers

import (
	"context"

	"github.com/pkg/errors"
	"github.com/tedyst/licenta/api/v1/generated"
	"github.com/tedyst/licenta/config"
	db "github.com/tedyst/licenta/db/generated"
	"github.com/tedyst/licenta/middleware/session"
	"github.com/tedyst/licenta/models"
	"go.opentelemetry.io/otel/trace"
)

func (*ServerHandler) PostLogin(ctx context.Context, request generated.PostLoginRequestObject) (generated.PostLoginResponseObject, error) {
	span := trace.SpanFromContext(ctx)

	err := valid.Struct(request)
	if err != nil {
		return generated.PostLogin400JSONResponse{
			Message: err.Error(),
			Success: false,
		}, nil
	}

	user, err := config.DatabaseQueries.GetUserByUsernameOrEmail(ctx, request.Body.Username)
	if err != nil {
		traceError(ctx, errors.Wrap(err, "PostLogin: error getting user"))
		return generated.PostLogin401JSONResponse{
			Message: InvalidCredentials,
			Success: false,
		}, nil
	}

	span.AddEvent("Start verifying password")
	ok, err := models.VerifyPassword(ctx, user, request.Body.Password)
	if err != nil {
		return nil, errors.Wrapf(err, "PostLogin: error verifying password for user `%s`", request.Body.Username)
	}
	span.AddEvent("Finished verifying password")

	if !ok {
		return generated.PostLogin401JSONResponse{
			Message: InvalidCredentials,
			Success: false,
		}, nil
	}

	session.SetUser(ctx, user)

	var SuccessTrue = true
	return generated.PostLogin200JSONResponse{
		Success: &SuccessTrue,
		User: &generated.User{
			Id:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		},
	}, nil
}

func (*ServerHandler) PostLogout(ctx context.Context, _ generated.PostLogoutRequestObject) (generated.PostLogoutResponseObject, error) {
	session.ClearSession(ctx)

	var SuccessTrue = true
	return generated.PostLogout200JSONResponse{
		Success: &SuccessTrue,
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

	user, err := config.DatabaseQueries.CreateUser(ctx, db.CreateUserParams{
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

	err = config.DatabaseQueries.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		ID:       user.ID,
		Password: user.Password,
	})

	if err != nil {
		return nil, errors.Wrap(err, "PostRegister: error updating user password")
	}

	// sess, err := session.GetSession(c.UserContext(), c)
	// if err != nil || sess == nil {
	// 	return nil, errors.Wrap(err, "PostRegister: error getting session")
	// }

	// sess.UserID = sql.NullInt64{
	// 	Int64: user.ID,
	// 	Valid: true,
	// }
	// sess.Waiting2fa = sql.NullInt64{}
	// sess.TotpKey = sql.NullString{}

	// err = session.SaveSession(c.UserContext(), c, sess)
	// if err != nil {
	// 	return nil, errors.Wrapf(err, "PostRegister: error saving session `%s`", sess.ID)
	// }

	var SuccessTrue = true
	return generated.PostRegister200JSONResponse{
		Success: &SuccessTrue,
		User: &generated.User{
			Id:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		},
	}, nil
}

func (*ServerHandler) Post2faTotpFirstStep(ctx context.Context, request generated.Post2faTotpFirstStepRequestObject) (generated.Post2faTotpFirstStepResponseObject, error) {
	return nil, nil
}

func (*ServerHandler) Post2faTotpSecondStep(ctx context.Context, request generated.Post2faTotpSecondStepRequestObject) (generated.Post2faTotpSecondStepResponseObject, error) {
	return nil, nil
}
