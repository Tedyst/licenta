package v1

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/tedyst/licenta/config"
	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/middleware/session"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func handleError(c *fiber.Ctx, span trace.Span, err error) error {
	span.SetStatus(codes.Error, err.Error())
	span.RecordError(err)
	c.Status(fiber.StatusInternalServerError)
	return c.JSON(fiber.Map{
		"error": "Internal server error",
	})
}

func getSessionAndUser(ctx context.Context, c *fiber.Ctx) (*db.Session, *db.User, error) {
	sess, err := session.GetSession(ctx, c)
	if err != nil {
		return nil, nil, err
	}
	user, err := config.DatabaseQueries.GetUser(ctx, sess.UserID.Int64)
	if err != nil {
		return nil, nil, err
	}
	return sess, &user, nil
}
