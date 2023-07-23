package v1

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/tedyst/licenta/config"
	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/handlers"
	"github.com/tedyst/licenta/middleware/session"
	"go.opentelemetry.io/otel/codes"
)

type loginAPIRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

const (
	Success             = "success"
	RequireTOTP         = "require_totp"
	InvalidCredentials  = "invalid_credentials"
	TOTPSetupNotStarted = "totp_setup_not_started"
)

type loginAPIResponse struct {
	Status string `json:"status"`
}

var (
	SuccessResponse        = loginAPIResponse{Status: Success}
	ErrInvalidCredentials  = loginAPIResponse{Status: InvalidCredentials}
	ErrRequireTOTOP        = loginAPIResponse{Status: RequireTOTP}
	ErrTotpSetupNotStarted = loginAPIResponse{Status: TOTPSetupNotStarted}
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
	ctx, span := config.Tracer.Start(c.UserContext(), "HandleLoginAPI")
	defer span.End()

	var req loginAPIRequest
	if err := c.BodyParser(&req); err != nil {
		span.SetStatus(codes.Error, "Error parsing body")
		span.RecordError(err)
		return err
	}

	sess, err := session.GetSession(ctx, c)
	if err != nil {
		return handleError(c, span, err)
	}

	status, err := handlers.HandleFirstStepLogin(ctx, sess, req.Username, req.Password)
	if err != nil {
		return handleError(c, span, err)
	}

	err = session.SaveSession(ctx, c, sess)
	if err != nil {
		return handleError(c, span, err)
	}

	switch status {
	case handlers.Success:
		return c.JSON(SuccessResponse)
	case handlers.RequireTOTP:
		return c.JSON(ErrRequireTOTOP)
	case handlers.InvalidCredentials:
		return c.JSON(ErrInvalidCredentials)
	}
	return c.JSON(ErrInvalidCredentials)
}

// @Summary		Logout
// @Description	Logout
// @Tags			auth
// @Accept			json
// @Produce		json
// @Success		200	{object}	loginAPIResponse
// @Router			/api/v1/logout [post]
func HandleLogoutAPI(c *fiber.Ctx) error {
	ctx, span := config.Tracer.Start(c.UserContext(), "HandleLogoutAPI")
	defer span.End()

	sess, err := session.GetSession(ctx, c)
	if err != nil {
		return handleError(c, span, err)
	}

	handlers.HandleLogout(ctx, sess)

	err = session.SaveSession(ctx, c, sess)
	if err != nil {
		return handleError(c, span, err)
	}
	return c.JSON(SuccessResponse)
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
	ctx, span := config.Tracer.Start(c.UserContext(), "HandleTOTPAPI")
	defer span.End()

	var req totpAPIRequest
	if err := c.BodyParser(&req); err != nil {
		return handleError(c, span, err)
	}

	sess, user, err := getSessionAndUser(ctx, c)
	if err != nil {
		return handleError(c, span, err)
	}

	status, err := handlers.HandleTOTPVerify(ctx, sess, user, req.Totp)
	if err != nil {
		return handleError(c, span, err)
	}

	err = session.SaveSession(ctx, c, sess)
	if err != nil {
		return handleError(c, span, err)
	}

	switch status {
	case handlers.Success:
		return c.JSON(SuccessResponse)
	case handlers.InvalidCredentials:
		return c.JSON(ErrInvalidCredentials)
	}
	return c.JSON(ErrInvalidCredentials)

}

type setupTOTPResponse struct {
	TotpSecret string `json:"totp_secret"`
}

// @Summary		Generate TOTP
// @Description	Generate TOTP
// @Tags			auth
// @Accept			json
// @Produce		json
// @Success		200	{object}	setupTOTPResponse
// @Router			/api/v1/totp/setup [post]
func HandleTOTPSetup(c *fiber.Ctx) error {
	ctx, span := config.Tracer.Start(c.UserContext(), "HandleTOTPSetup")
	defer span.End()

	sess, user, err := getSessionAndUser(ctx, c)
	if err != nil {
		return handleError(c, span, err)
	}

	status, err := handlers.HandleTOTPSetup(ctx, sess, user)
	if err != nil {
		return handleError(c, span, err)
	}

	err = session.SaveSession(ctx, c, sess)
	if err != nil {
		return handleError(c, span, err)
	}

	err = config.DatabaseQueries.UpdateUserTOTPSecret(ctx, db.UpdateUserTOTPSecretParams{
		ID:         user.ID,
		TotpSecret: user.TotpSecret,
	})

	if err != nil {
		return handleError(c, span, err)
	}

	switch status {
	case handlers.Success:
		return c.JSON(setupTOTPResponse{
			TotpSecret: user.TotpSecret.String,
		})
	}
	return c.JSON(ErrTotpSetupNotStarted)
}

type enableTotpAPIRequest struct {
	Totp string `json:"totp"`
}

// @Summary		Enable TOTP
// @Description	Enable TOTP
// @Tags			auth
// @Accept			json
// @Produce		json
// @Param 			user body enableTotpAPIRequest true "User"
// @Success		200	{object}	loginAPIResponse
// @Router			/api/v1/enable_totp [post]
func HandleEnableTotp(c *fiber.Ctx) error {
	ctx, span := config.Tracer.Start(c.UserContext(), "HandleEnableTOTP")
	defer span.End()

	var req enableTotpAPIRequest
	if err := c.BodyParser(&req); err != nil {
		span.SetStatus(codes.Error, "Error parsing body")
		span.RecordError(err)
		return err
	}

	sess, err := session.GetSession(ctx, c)
	if err != nil {
		span.SetStatus(codes.Error, "Error getting session")
		span.RecordError(err)
		return err
	}
	if sess.UserID.Valid == false {
		return fiber.ErrUnauthorized
	}
	if sess.TotpKey.Valid == false {
		return fiber.ErrUnauthorized
	}
	user, err := config.DatabaseQueries.GetUser(c.Context(), sess.UserID.Int64)
	if err != nil {
		span.SetStatus(codes.Error, "Error getting user")
		span.RecordError(err)
		return err
	}
	user.TotpSecret = sess.TotpKey

	ok := user.VerifyTOTP(req.Totp)
	if !ok {
		return c.JSON(ErrInvalidCredentials)
	}

	sess.TotpKey = pgtype.Text{Valid: false}
	err = session.SaveSession(ctx, c, sess)
	if err != nil {
		span.SetStatus(codes.Error, "Error saving session")
		span.RecordError(err)
		return err
	}

	err = config.DatabaseQueries.UpdateUserTOTPSecret(ctx, db.UpdateUserTOTPSecretParams{
		ID:         user.ID,
		TotpSecret: user.TotpSecret,
	})
	if err != nil {
		span.SetStatus(codes.Error, "Error updating user")
		span.RecordError(err)
		return err
	}

	return c.JSON(SuccessResponse)
}
