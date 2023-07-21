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

func verifyIfLoggedIn(ctx context.Context, c *fiber.Ctx) error {
	_, span := config.Tracer.Start(ctx, "verifyIfLoggedIn")
	defer span.End()

	sess, err := session.GetSession(ctx, c)
	if err != nil {
		span.AddEvent("Error getting session")
		span.RecordError(err)
		return err
	}
	if sess.UserID.Valid == true {
		span.AddEvent("User not logged in")
		return fiber.ErrUnauthorized
	}
	return nil
}

func verifyIfAdmin(ctx context.Context, c *fiber.Ctx) error {
	ctx, span := config.Tracer.Start(ctx, "verifyIfAdmin")
	defer span.End()

	sess, err := session.GetSession(ctx, c)
	if err != nil {
		span.SetStatus(codes.Error, "Error getting session")
		span.RecordError(err)
		return err
	}
	if sess.UserID.Valid == false {
		span.AddEvent("User not logged in")
		return fiber.ErrUnauthorized
	}
	user, err := config.DatabaseQueries.GetUser(ctx, sess.UserID.Int64)
	if err != nil {
		span.SetStatus(codes.Error, "Error getting user")
		span.RecordError(err)
		return err
	}
	if !user.Admin {
		span.AddEvent("User is not admin")
		return fiber.ErrForbidden
	}
	return nil
}
