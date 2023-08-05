package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"github.com/tedyst/licenta/api/v1/generated"
	"github.com/tedyst/licenta/config"
	db "github.com/tedyst/licenta/db/generated"
	"github.com/tedyst/licenta/middleware/session"
)

func (*ServerHandler) PostLogin(c *fiber.Ctx) error {
	var body generated.PostLoginJSONRequestBody
	if err := c.BodyParser(&body); err != nil {
		return errors.Wrap(err, "Error parsing body")
	}

	user, err := config.DatabaseQueries.GetUserByUsernameOrEmail(c.UserContext(), body.Username)
	if err != nil {
		return sendError(c, fiber.StatusUnauthorized, "Invalid credentials")
	}

	ok, err := user.VerifyPassword(body.Password)
	if err != nil {
		return errors.Wrap(err, "Error verifying password")
	}

	if !ok {
		return sendError(c, fiber.StatusUnauthorized, "Invalid credentials")
	}

	sess, err := getSession(c)
	if err != nil || sess == nil {
		return errors.Wrap(err, "Error getting session")
	}
	sess.UserID = pgtype.Int8{Int64: user.ID, Valid: true}
	sess.Waiting2fa = pgtype.Int8{Valid: false}
	sess.TotpKey = pgtype.Text{Valid: false}

	session.SaveSession(c.UserContext(), c, sess)

	return nil
}

func (*ServerHandler) PostLogout(c *fiber.Ctx) error {
	sess, err := getSession(c)
	if err != nil || sess == nil {
		return errors.Wrap(err, "Error getting session")
	}
	sess.UserID = pgtype.Int8{Valid: false}
	sess.Waiting2fa = pgtype.Int8{Valid: false}
	sess.TotpKey = pgtype.Text{Valid: false}

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

	err = user.SetPassword(body.Password)
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

	sess.UserID = pgtype.Int8{Int64: user.ID, Valid: true}
	sess.Waiting2fa = pgtype.Int8{Valid: false}
	sess.TotpKey = pgtype.Text{Valid: false}

	session.SaveSession(c.UserContext(), c, sess)

	return nil
}

func (*ServerHandler) Post2faTotpFirstStep(c *fiber.Ctx) error {
	return nil
}

func (*ServerHandler) Post2faTotpSecondStep(c *fiber.Ctx) error {
	return nil
}
