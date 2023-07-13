package v1

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tedyst/licenta/config"
	"github.com/tedyst/licenta/db"
)

type loginAPIRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

const (
	Success            = "success"
	RequireTOTP        = "require_totp"
	InvalidCredentials = "invalid_credentials"
)

type loginAPIResponse struct {
	Status string `json:"status"`
}

var (
	SuccessResponse       = loginAPIResponse{Status: Success}
	ErrInvalidCredentials = loginAPIResponse{Status: InvalidCredentials}
	ErrRequireTOTOP       = loginAPIResponse{Status: RequireTOTP}
)

// @Summary		Login
// @Description	Login
// @Tags			auth
// @Accept			json
// @Produce		json
// @Param 			user body loginAPIRequest true "User"
// @Success		200	{object}	loginAPIResponse
// @Router			/api/v1/login [post]
func HandleLoginAPI(c *fiber.Ctx) error {
	var req loginAPIRequest
	if err := c.BodyParser(&req); err != nil {
		return err
	}
	user, err := config.DatabaseQueries.GetUserByUsernameOrEmail(c.Context(), req.Username)
	if err != nil {
		return c.JSON(ErrInvalidCredentials)
	}
	ok, err := user.VerifyPassword(req.Password)
	if err != nil {
		return err
	}
	if !ok {
		return c.JSON(ErrInvalidCredentials)
	}

	sess, err := config.SessionStore.Get(c)
	if err != nil {
		return err
	}

	if user.TotpSecret.Valid {
		sess.Set(UserID2FAKey, user.ID)
		sess.Save()
		return c.JSON(ErrRequireTOTOP)
	}
	if err := loginUser(c, &user); err != nil {
		return err
	}
	sess.Set(UserIDKey, nil)
	return c.JSON(Success)
}

// @Summary		Logout
// @Description	Logout
// @Tags			auth
// @Accept			json
// @Produce		json
// @Success		200	{object}	loginAPIResponse
// @Router			/api/v1/logout [post]
func HandleLogoutAPI(c *fiber.Ctx) error {
	if err := logoutUser(c); err != nil {
		return err
	}
	return c.JSON(Success)
}

type totpAPIRequest struct {
	Totp string `json:"totp"`
}

// @Summary		Verify TOTP
// @Description	Verify TOTP
// @Tags			auth
// @Accept			json
// @Produce		json
// @Param 			user body totpAPIRequest true "User"
// @Success		200	{object}	loginAPIResponse
// @Router			/api/v1/totp [post]
func HandleTOTPAPI(c *fiber.Ctx) error {
	var req totpAPIRequest
	if err := c.BodyParser(&req); err != nil {
		return err
	}
	sess, err := config.SessionStore.Get(c)
	if err != nil {
		return err
	}
	userID := sess.Get(UserID2FAKey)
	if userID == nil {
		return fiber.ErrUnauthorized
	}
	user, err := config.DatabaseQueries.GetUser(c.Context(), userID.(int64))
	if err != nil {
		return err
	}
	ok := user.VerifyTOTP(req.Totp)
	if !ok {
		return c.JSON(ErrInvalidCredentials)
	}
	if err := loginUser(c, &user); err != nil {
		return err
	}
	sess.Set(UserID2FAKey, nil)
	return c.JSON(Success)
}

type generateTotpAPIResponse struct {
	TotpSecret string `json:"totp_secret"`
}

// @Summary		Generate TOTP
// @Description	Generate TOTP
// @Tags			auth
// @Accept			json
// @Produce		json
// @Success		200	{object}	generateTotpAPIResponse
// @Router			/api/v1/generate_totp [post]
func HandleGenerateTOTP(c *fiber.Ctx) error {
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
	user.GenerateTOTPSecret()
	if err := config.DatabaseQueries.UpdateUserTOTPSecret(c.Context(), db.UpdateUserTOTPSecretParams{
		ID:         user.ID,
		TotpSecret: user.TotpSecret,
	}); err != nil {
		return err
	}
	return c.JSON(generateTotpAPIResponse{
		TotpSecret: user.TotpSecret.String,
	})
}
