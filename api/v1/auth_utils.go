package v1

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/tedyst/licenta/config"
	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/middleware/session"
)

const (
	UserIDKey    = "user_id"
	UserID2FAKey = "user_id_totp"
)

func verifyIfLoggedIn(ctx context.Context, c *fiber.Ctx) (*db.Session, *db.User, error) {
	_, span := config.Tracer.Start(ctx, "verifyIfLoggedIn")
	defer span.End()

	sess, user, err := session.GetSessionAndUser(ctx, c)
	if err != nil {
		span.AddEvent("Error getting session")
		span.RecordError(err)
		return nil, nil, err
	}
	if user == nil {
		span.AddEvent("User not logged in")
		return nil, nil, fiber.ErrUnauthorized
	}
	return sess, user, nil
}

func verifyIfAdmin(ctx context.Context, c *fiber.Ctx) (*db.Session, *db.User, error) {
	ctx, span := config.Tracer.Start(ctx, "verifyIfAdmin")
	defer span.End()

	sess, user, err := session.GetSessionAndUser(ctx, c)
	if err != nil {
		span.AddEvent(err.Error())
		span.RecordError(err)
		return nil, nil, err
	}
	if user == nil {
		span.AddEvent("User not logged in")
		return nil, nil, fiber.ErrUnauthorized
	}
	if !user.Admin {
		span.AddEvent("User is not admin")
		return nil, nil, fiber.ErrForbidden
	}
	return sess, user, nil
}
