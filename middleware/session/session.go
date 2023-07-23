package session

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/tedyst/licenta/config"
	"github.com/tedyst/licenta/db"
)

const sessionKey = "session"

func createNewSession(ctx context.Context, c *fiber.Ctx) (*db.Session, error) {
	sess, err := config.DatabaseQueries.CreateSession(ctx, db.CreateSessionParams{
		ID: uuid.New(),
	})
	if err != nil {
		return nil, err
	}
	c.Cookie(&fiber.Cookie{
		Name:     sessionKey,
		Value:    sess.ID.String(),
		HTTPOnly: true,
		SameSite: "Strict",
	})
	return &sess, nil
}

func GetSession(ctx context.Context, c *fiber.Ctx) (*db.Session, error) {
	ctx, span := config.Tracer.Start(ctx, "GetSession")
	defer span.End()

	sess_id := c.Cookies(sessionKey)
	var u uuid.UUID
	if len(sess_id) == 0 {
		return createNewSession(ctx, c)
	}
	var err error
	u, err = uuid.Parse(sess_id)
	if err != nil {
		return createNewSession(ctx, c)
	}
	sess, err := config.DatabaseQueries.GetSession(ctx, u)
	if err != nil {
		return createNewSession(ctx, c)
	}
	return &sess, nil
}

func GetSessionAndUser(ctx context.Context, c *fiber.Ctx) (*db.Session, *db.User, error) {
	ctx, span := config.Tracer.Start(ctx, "GetSessionAndUser")
	defer span.End()

	sess_id := c.Cookies(sessionKey)
	var u uuid.UUID
	if len(sess_id) == 0 {
		sess, err := createNewSession(ctx, c)
		return sess, nil, err
	}
	var err error
	u, err = uuid.Parse(sess_id)
	if err != nil {
		sess, err := createNewSession(ctx, c)
		return sess, nil, err
	}
	row, err := config.DatabaseQueries.GetUserAndSessionBySessionID(ctx, u)
	if err != nil {
		sess, err := createNewSession(ctx, c)
		return sess, nil, err
	}
	return &row.Session, &row.User, nil
}

func SaveSession(ctx context.Context, c *fiber.Ctx, sess *db.Session) error {
	ctx, span := config.Tracer.Start(ctx, "SaveSession")
	defer span.End()

	err := config.DatabaseQueries.UpdateSession(ctx, db.UpdateSessionParams{
		ID:      sess.ID,
		UserID:  sess.UserID,
		TotpKey: sess.TotpKey,
	})
	if err != nil {
		return err
	}
	c.Cookie(&fiber.Cookie{
		Name:     sessionKey,
		Value:    sess.ID.String(),
		HTTPOnly: true,
		SameSite: "Strict",
	})
	return nil
}

func ClearSession(ctx context.Context, c *fiber.Ctx) error {
	ctx, span := config.Tracer.Start(ctx, "ClearSession")
	defer span.End()

	c.Cookie(&fiber.Cookie{
		Name:     sessionKey,
		Value:    "",
		HTTPOnly: true,
		SameSite: "Strict",
		Expires:  time.Now().Add(-1 * time.Hour),
	})
	uuid, err := uuid.ParseBytes(c.Request().Header.Peek(sessionKey))
	if err != nil {
		return err
	}
	return config.DatabaseQueries.DeleteSession(ctx, uuid)
}
