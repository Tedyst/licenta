package v1

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/tedyst/licenta/config"
	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/middleware/session"
	"go.opentelemetry.io/otel/codes"
)

const (
	UserIDKey    = "user_id"
	UserID2FAKey = "user_id_totp"
)

func loginUser(ctx context.Context, c *fiber.Ctx, user *db.User) error {
	_, span := config.Tracer.Start(ctx, "loginUser")
	defer span.End()

	sess, err := session.GetSession(ctx, c)
	if err != nil {
		span.SetStatus(codes.Error, "Error getting session")
		span.RecordError(err)
		return err
	}
	sess.TotpKey = pgtype.Text{Valid: false}
	sess.UserID = pgtype.Int8{Int64: user.ID, Valid: true}
	err = session.SaveSession(ctx, c, sess)
	if err != nil {
		span.SetStatus(codes.Error, "Error saving session")
		span.RecordError(err)
		return err
	}
	return nil
}

func logoutUser(ctx context.Context, c *fiber.Ctx) error {
	_, span := config.Tracer.Start(ctx, "logoutUser")
	defer span.End()

	err := session.ClearSession(ctx, c)
	if err != nil {
		span.SetStatus(codes.Error, "Error saving session")
		span.RecordError(err)
		return err
	}
	return nil
}

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
