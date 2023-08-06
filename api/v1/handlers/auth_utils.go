package handlers

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/tedyst/licenta/config"
	db "github.com/tedyst/licenta/db/generated"
	"github.com/tedyst/licenta/middleware/session"
)

const (
	ContextUserKey    = "user"
	ContextSessionKey = "session"
	UserIDKey         = "user_id"
	UserID2FAKey      = "user_id_totp"
)

var valid = validator.New()

//lint:ignore U1000 Ignore unused function temporarily for debugging
func verifyIfLoggedIn(c *fiber.Ctx) (*db.Session, *db.User, error) {
	_, span := config.Tracer.Start(c.UserContext(), "verifyIfLoggedIn")
	defer span.End()

	user := c.Locals(ContextUserKey).(*db.User)
	sess := c.Locals(ContextSessionKey).(*db.Session)
	if sess != nil {
		return sess, user, nil
	}

	sess, user, err := session.GetSessionAndUser(c.UserContext(), c)
	if err != nil {
		newErr := errors.Wrap(err, "verifyIfLoggedIn: error getting session")
		span.RecordError(newErr)
		return nil, nil, newErr
	}
	if user == nil {
		span.AddEvent("User is not logged in")
		return nil, nil, fiber.ErrUnauthorized
	}
	c.Locals(ContextUserKey, user)
	c.Locals(ContextSessionKey, sess)
	return sess, user, nil
}

//lint:ignore U1000 Ignore unused function temporarily for debugging
func verifyIfAdmin(ctx context.Context, c *fiber.Ctx) (*db.Session, *db.User, error) {
	_, span := config.Tracer.Start(ctx, "verifyIfAdmin")
	defer span.End()

	sess, user, err := verifyIfLoggedIn(c)
	if err != nil {
		return nil, nil, err
	}
	if !user.Admin {
		span.AddEvent("User is not admin")
		return nil, nil, fiber.ErrForbidden
	}
	return sess, user, nil
}
