package v1

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tedyst/licenta/config"
	"github.com/tedyst/licenta/db"
)

const (
	UserIDKey    = "user_id"
	UserID2FAKey = "user_id_totp"
)

func loginUser(c *fiber.Ctx, user *db.User) error {
	sess, err := config.SessionStore.Get(c)
	if err != nil {
		return err
	}
	sess.Delete(UserID2FAKey)
	sess.Set(UserIDKey, user.ID)
	return sess.Save()
}

func logoutUser(c *fiber.Ctx) error {
	sess, err := config.SessionStore.Get(c)
	if err != nil {
		return err
	}
	sess.Delete(UserIDKey)
	sess.Delete(UserID2FAKey)
	return sess.Save()
}

func verifyIfLoggedIn(c *fiber.Ctx) error {
	sess, err := config.SessionStore.Get(c)
	if err != nil {
		return err
	}
	userID := sess.Get(UserIDKey)
	if userID == nil {
		return fiber.ErrUnauthorized
	}
	return nil
}

func verifyIfAdmin(c *fiber.Ctx) error {
	sess, err := config.SessionStore.Get(c)
	if err != nil {
		return err
	}
	userID := sess.Get(UserIDKey)
	if userID == nil {
		return fiber.ErrUnauthorized
	}
	user, err := config.DatabaseQueries.GetUser(c.Context(), userID.(int64))
	if err != nil {
		return err
	}
	if !user.Admin {
		return fiber.ErrForbidden
	}
	return nil
}
