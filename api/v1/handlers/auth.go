package handlers

import (
	"database/sql"

	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/tedyst/licenta/api/v1/generated"
	"github.com/tedyst/licenta/config"
	db "github.com/tedyst/licenta/db/generated"
	"github.com/tedyst/licenta/middleware/session"
	"github.com/tedyst/licenta/models"
)

func (*ServerHandler) PostLogin(c *fiber.Ctx, request generated.PostLoginRequestObject) (generated.PostLoginResponseObject, error) {
	err := valid.Struct(request)
	if err != nil {
		return generated.PostLogin400JSONResponse{
			Message: err.Error(),
			Success: false,
		}, nil
	}

	user, err := config.DatabaseQueries.GetUserByUsernameOrEmail(c.UserContext(), request.Body.Username)
	if err != nil {
		traceError(c, errors.Wrap(err, "PostLogin: error getting user"))
		return generated.PostLogin401JSONResponse{
			Message: InvalidCredentials,
			Success: false,
		}, nil
	}

	ok, err := models.VerifyPassword(user, request.Body.Password)
	if err != nil {
		return nil, errors.Wrapf(err, "PostLogin: error verifying password for user `%s`", request.Body.Username)
	}

	if !ok {
		return generated.PostLogin401JSONResponse{
			Message: InvalidCredentials,
			Success: false,
		}, nil
	}

	sess, err := session.GetSession(c.UserContext(), c)
	if err != nil || sess == nil {
		return nil, errors.Wrap(err, "PostLogin: error getting session")
	}

	if user.TotpSecret.Valid {
		sess.Waiting2fa = sql.NullInt64{
			Int64: user.ID,
			Valid: true,
		}
	} else {
		sess.UserID = sql.NullInt64{
			Int64: user.ID,
			Valid: true,
		}
		sess.Waiting2fa = sql.NullInt64{}
	}
	sess.TotpKey = sql.NullString{}

	err = session.SaveSession(c.UserContext(), c, sess)
	if err != nil {
		return nil, errors.Wrapf(err, "PostLogin: error saving session `%s`", sess.ID)
	}

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

func (*ServerHandler) PostLogout(c *fiber.Ctx) error {
	sess, err := session.GetSession(c.UserContext(), c)
	if err != nil || sess == nil {
		return errors.Wrap(err, "Error getting session")
	}
	sess.UserID = sql.NullInt64{}
	sess.Waiting2fa = sql.NullInt64{}
	sess.TotpKey = sql.NullString{}

	err = session.ClearSession(c.UserContext(), c)

	return errors.Wrap(err, "PostLogout: error saving session")
}

func (*ServerHandler) PostRegister(c *fiber.Ctx) error {
	var body generated.PostRegisterJSONRequestBody
	if err := c.BodyParser(&body); err != nil {
		return sendError(c, fiber.StatusBadRequest, err.Error())
	}

	err := valid.Struct(body)
	if err != nil {
		return sendError(c, fiber.StatusBadRequest, err.Error())
	}

	user, err := config.DatabaseQueries.CreateUser(c.UserContext(), db.CreateUserParams{
		Username: body.Username,
		Email:    body.Email,
	})
	if err != nil {
		return sendError(c, fiber.StatusBadRequest, "Username or email already exists")
	}

	err = models.SetPassword(user, body.Password)
	if err != nil {
		return errors.Wrap(err, "Error setting password")
	}

	err = config.DatabaseQueries.UpdateUserPassword(c.UserContext(), db.UpdateUserPasswordParams{
		ID:       user.ID,
		Password: user.Password,
	})

	if err != nil {
		return errors.Wrap(err, "Error updating password")
	}

	sess, err := getSession(c)
	if err != nil || sess == nil {
		return errors.Wrap(err, "Error getting session")
	}

	sess.UserID = sql.NullInt64{}
	sess.Waiting2fa = sql.NullInt64{}
	sess.TotpKey = sql.NullString{}

	err = session.SaveSession(c.UserContext(), c, sess)

	return errors.Wrap(err, "PostRegister: error saving session")
}

func (*ServerHandler) Post2faTotpFirstStep(c *fiber.Ctx) error {
	return nil
}

func (*ServerHandler) Post2faTotpSecondStep(c *fiber.Ctx) error {
	return nil
}
