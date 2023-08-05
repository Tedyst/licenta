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

func (*ServerHandler) PostLogin(c *fiber.Ctx) error {
	var body generated.PostLoginJSONRequestBody
	if err := c.BodyParser(&body); err != nil {
		return err
	}

	user, err := config.DatabaseQueries.GetUserByUsernameOrEmail(c.UserContext(), body.Username)
	if err != nil {
		traceError(c, errors.Wrap(err, "PostLogin: error getting user"))
		return sendError(c, fiber.StatusUnauthorized, InvalidCredentials)
	}

	ok, err := models.VerifyPassword(user, body.Password)
	if err != nil {
		return errors.Wrapf(err, "PostLogin: error verifying password for user `%s`", body.Username)
	}

	if !ok {
		return sendError(c, fiber.StatusUnauthorized, InvalidCredentials)
	}

	sess, err := getSession(c)
	if err != nil || sess == nil {
		return errors.Wrap(err, ErrorGettingSession)
	}
	sess.UserID = sql.NullInt64{}
	sess.Waiting2fa = sql.NullInt64{}
	sess.TotpKey = sql.NullString{}

	err = session.SaveSession(c.UserContext(), c, sess)
	if err != nil {
		return errors.Wrapf(err, "PostLogin: error saving session `%s`", sess.ID)
	}

	return c.JSON(generated.User{
		Id:       user.ID,
		Username: user.Username,
		Email:    user.Email,
	})
}

func (*ServerHandler) PostLogout(c *fiber.Ctx) error {
	sess, err := getSession(c)
	if err != nil || sess == nil {
		return errors.Wrap(err, "Error getting session")
	}
	sess.UserID = sql.NullInt64{}
	sess.Waiting2fa = sql.NullInt64{}
	sess.TotpKey = sql.NullString{}

	session.SaveSession(c.UserContext(), c, sess)

	return nil
}

func (*ServerHandler) PostRegister(c *fiber.Ctx) error {
	var body generated.PostRegisterJSONRequestBody
	if err := c.BodyParser(&body); err != nil {
		return errors.Wrap(err, "Error parsing body")
	}

	user, err := config.DatabaseQueries.CreateUser(c.UserContext(), db.CreateUserParams{
		Username: body.Username,
		Email:    body.Email,
	})
	if err != nil {
		return errors.Wrap(err, "Error creating user")
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

	session.SaveSession(c.UserContext(), c, sess)

	return nil
}

func (*ServerHandler) Post2faTotpFirstStep(c *fiber.Ctx) error {
	return nil
}

func (*ServerHandler) Post2faTotpSecondStep(c *fiber.Ctx) error {
	return nil
}
