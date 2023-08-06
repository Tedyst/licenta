package session

import (
	"context"
	"database/sql"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/tedyst/licenta/config"
	db "github.com/tedyst/licenta/db/generated"
)

const cookieSessionKey = "session"

type contextSessionKey struct{}
type contextUserKey struct{}
type contextOriginalContextKey struct{}

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
	c.Locals(contextSessionKey{}, sess)
	return sess, nil
}

func getSessionAndUser(ctx context.Context, c *fiber.Ctx) (*db.Session, *db.User, error) {
	ctx, span := config.Tracer.Start(ctx, "getSessionAndUser")
	defer span.End()

	if c.Locals(contextSessionKey{}) != nil && c.Locals(contextUserKey{}) != nil {
		return c.Locals(contextSessionKey{}).(*db.Session), c.Locals(contextUserKey{}).(*db.User), nil
	} else if c.Locals(contextSessionKey{}) != nil {
		return c.Locals(contextSessionKey{}).(*db.Session), nil, nil
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
	sess, err := config.DatabaseQueries.GetSession(ctx, u)
	if err != nil {
		sess, err := createNewSession(ctx, c)
		return sess, nil, errors.Wrap(err, "GetSessionAndUser: error getting session")
	}

	var user *db.User
	if sess.UserID.Valid {
		user, err = config.DatabaseQueries.GetUser(ctx, sess.UserID.Int64)
		if err != nil {
			return sess, nil, errors.Wrap(err, "GetSessionAndUser: error getting user")
		}
		c.Locals(contextUserKey{}, user)
	}

	c.Locals(contextSessionKey{}, sess)
	return sess, user, nil
}

func saveSession(ctx context.Context, c *fiber.Ctx, sess *db.Session) error {
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

	c.Locals(contextSessionKey{}, sess)
	return nil
}

func GetUser(ctx context.Context) (*db.User, error) {
	ctx, span := config.Tracer.Start(ctx, "GetUser")
	defer span.End()

	c := ctx.Value(contextOriginalContextKey{})
	if c == nil {
		return nil, errors.New("GetUser: cannot find context")
	}
	cc := c.(*fiber.Ctx)
	u := cc.Locals(contextUserKey{})
	if u != nil {
		return u.(*db.User), nil
	}

	_, user, err := getSessionAndUser(ctx, cc)
	if err != nil {
		return nil, errors.Wrap(err, "GetUser: error getting session and user")
	}
	return user, nil
}

func getSessionFromContext(ctx context.Context) (*db.Session, *fiber.Ctx, error) {
	c := ctx.Value(contextOriginalContextKey{})
	if c == nil {
		return nil, nil, errors.New("getSessionFromContext: cannot find context")
	}
	cc := c.(*fiber.Ctx)
	sess := cc.Locals(contextSessionKey{})
	sesss := sess.(*db.Session)
	return sesss, cc, nil
}

func SetUser(ctx context.Context, user *db.User) error {
	ctx, span := config.Tracer.Start(ctx, "SetUser")
	defer span.End()

	sess, c, err := getSessionFromContext(ctx)
	if err != nil {
		return errors.Wrap(err, "SetUser: error getting session from context")
	}

	c.Locals(contextUserKey{}, user)

	sess.UserID = sql.NullInt64{
		Int64: user.ID,
		Valid: true,
	}
	sess.TotpKey = sql.NullString{}
	sess.Waiting2fa = sql.NullInt64{}

	return saveSession(ctx, c, sess)
}

func ClearSession(ctx context.Context) error {
	ctx, span := config.Tracer.Start(ctx, "ClearSession")
	defer span.End()

	sess, c, err := getSessionFromContext(ctx)
	if err != nil {
		return errors.Wrap(err, "ClearSession: error getting session from context")
	}

	sess.UserID = sql.NullInt64{}
	sess.TotpKey = sql.NullString{}
	sess.Waiting2fa = sql.NullInt64{}

	return saveSession(ctx, c, sess)
}

func SessionMiddleware() func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		sess, user, err := getSessionAndUser(c.UserContext(), c)
		if err != nil {
			return errors.Wrap(err, "SessionMiddleware: error getting session and user")
		}

		ctx := context.WithValue(c.UserContext(), contextOriginalContextKey{}, c)
		c.SetUserContext(ctx)

		c.Locals(contextSessionKey{}, sess)
		c.Locals(contextUserKey{}, user)

		err = c.Next()
		if err != nil {
			return errors.Wrap(err, "SessionMiddleware: error calling next")
		}

		return nil
	}
}
