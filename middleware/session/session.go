package session

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/tedyst/licenta/config"
	db "github.com/tedyst/licenta/db/generated"
)

const (
	cookieSessionKey  = "session"
	contextUserKey    = "user"
	contextSessionKey = "session"
)

func createNewSession(ctx context.Context, c *fiber.Ctx) (*db.Session, error) {
	sess, err := config.DatabaseQueries.CreateSession(ctx, db.CreateSessionParams{
		ID: uuid.New(),
	})
	if err != nil {
		return nil, err
	}
	c.Cookie(&fiber.Cookie{
		Name:     cookieSessionKey,
		Value:    sess.ID.String(),
		HTTPOnly: true,
		SameSite: "Strict",
	})
	c.Locals(contextSessionKey, sess)
	return sess, nil
}

func GetSession(ctx context.Context, c *fiber.Ctx) (*db.Session, error) {
	ctx, span := config.Tracer.Start(ctx, "GetSession")
	defer span.End()

	if c.Locals(contextSessionKey) != nil {
		return c.Locals(contextSessionKey).(*db.Session), nil
	}

	sess_id := c.Cookies(cookieSessionKey)
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
	c.Locals(contextSessionKey, sess)
	return sess, nil
}

func GetSessionAndUser(ctx context.Context, c *fiber.Ctx) (*db.Session, *db.User, error) {
	ctx, span := config.Tracer.Start(ctx, "GetSessionAndUser")
	defer span.End()

	if c.Locals(contextSessionKey) != nil && c.Locals(contextUserKey) != nil {
		return c.Locals(contextSessionKey).(*db.Session), c.Locals(contextUserKey).(*db.User), nil
	} else if c.Locals(contextSessionKey) != nil {
		return c.Locals(contextSessionKey).(*db.Session), nil, nil
	}

	sess_id := c.Cookies(cookieSessionKey)
	var u uuid.UUID
	if len(sess_id) == 0 {
		sess, err := createNewSession(ctx, c)
		return sess, nil, errors.Wrap(err, "GetSessionAndUser: error creating new session")
	}
	var err error
	u, err = uuid.Parse(sess_id)
	if err != nil {
		sess, err := createNewSession(ctx, c)
		return sess, nil, errors.Wrap(err, "GetSessionAndUser: error parsing session id")
	}
	row, err := config.DatabaseQueries.GetUserAndSessionBySessionID(ctx, u)
	if err != nil {
		sess, err := createNewSession(ctx, c)
		return sess, nil, errors.Wrap(err, "GetSessionAndUser: error getting user and session")
	}

	c.Locals(contextSessionKey, &row.Session)
	c.Locals(contextUserKey, &row.User)
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
		return errors.Wrap(err, "SaveSession: error updating session")
	}
	c.Cookie(&fiber.Cookie{
		Name:     cookieSessionKey,
		Value:    sess.ID.String(),
		HTTPOnly: true,
		SameSite: "Strict",
	})

	c.Locals(contextSessionKey, sess)
	return nil
}

func deleteSession(ctx context.Context, sess *db.Session) error {
	ctx, span := config.Tracer.Start(ctx, "deleteSession")
	defer span.End()

	return config.DatabaseQueries.DeleteSession(ctx, sess.ID)
}

func ClearSession(ctx context.Context, c *fiber.Ctx) error {
	ctx, span := config.Tracer.Start(ctx, "ClearSession")
	defer span.End()

	if c.Locals(contextSessionKey) == nil {
		return nil
	}

	sess := c.Locals(contextSessionKey).(*db.Session)
	err := deleteSession(ctx, sess)
	if err != nil {
		return errors.Wrap(err, "ClearSession: error deleting session")
	}

	c.Cookie(&fiber.Cookie{
		Name:     cookieSessionKey,
		Value:    "",
		HTTPOnly: true,
		SameSite: "Strict",
		Expires:  time.Now().Add(-1 * time.Hour),
	})
	uuid, err := uuid.ParseBytes(c.Request().Header.Peek(cookieSessionKey))
	if err != nil {
		return errors.Wrap(err, "ClearSession: error parsing session id")
	}

	c.Locals(contextSessionKey, nil)
	c.Locals(contextUserKey, nil)
	return config.DatabaseQueries.DeleteSession(ctx, uuid)
}
