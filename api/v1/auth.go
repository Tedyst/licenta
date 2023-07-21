package v1

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/tedyst/licenta/config"
	"github.com/tedyst/licenta/db"
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
	user, err := config.DatabaseQueries.GetUserByUsernameOrEmail(ctx, req.Username)
	if err != nil {
		span.SetStatus(codes.Error, "Error getting user")
		span.RecordError(err)
		return c.JSON(ErrInvalidCredentials)
	}
	ok, err := user.VerifyPassword(req.Password)
	if err != nil {
		span.SetStatus(codes.Error, "Error verifying password")
		span.RecordError(err)
		return err
	}
	if !ok {
		span.SetStatus(codes.Error, "Invalid credentials")
		return c.JSON(ErrInvalidCredentials)
	}

	sess, err := session.GetSession(ctx, c)
	if err != nil {
		span.SetStatus(codes.Error, "Error getting session")
		span.RecordError(err)
		return err
	}

	if user.TotpSecret.Valid {
		span.AddEvent("TOTP required")
		sess.Waiting2fa = true
		sess.UserID = pgtype.Int8{Int64: user.ID, Valid: true}
		err = session.SaveSession(ctx, c, sess)
		if err != nil {
			span.SetStatus(codes.Error, "Error saving session")
			span.RecordError(err)
			return err
		}
		return c.JSON(ErrRequireTOTOP)
	}
	if err := loginUser(ctx, c, &user); err != nil {
		span.SetStatus(codes.Error, "Error logging in user")
		span.RecordError(err)
		return err
	}
	return c.JSON(SuccessResponse)
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

	if err := logoutUser(ctx, c); err != nil {
		return err
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
	if !sess.TotpKey.Valid {
		return c.JSON(ErrInvalidCredentials)
	}
	user, err := config.DatabaseQueries.GetUser(ctx, sess.UserID.Int64)
	if err != nil {
		span.SetStatus(codes.Error, "Error getting user")
		span.RecordError(err)
		return err
	}
	ok := user.VerifyTOTP(req.Totp)
	if !ok {
		return c.JSON(ErrInvalidCredentials)
	}
	if err := loginUser(ctx, c, &user); err != nil {
		return err
	}
	session.SaveSession(ctx, c, sess)
	return c.JSON(SuccessResponse)
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
	ctx, span := config.Tracer.Start(c.UserContext(), "HandleGenerateTOTP")
	defer span.End()

	sess, err := session.GetSession(ctx, c)
	if err != nil {
		span.SetStatus(codes.Error, "Error getting session")
		span.RecordError(err)
		return err
	}
	if sess.UserID.Valid == false {
		return fiber.ErrUnauthorized
	}
	user, err := config.DatabaseQueries.GetUser(c.Context(), sess.UserID.Int64)
	if err != nil {
		span.SetStatus(codes.Error, "Error getting user")
		span.RecordError(err)
		return err
	}
	user.GenerateTOTPSecret()
	sess.TotpKey = pgtype.Text{String: user.TotpSecret.String, Valid: true}
	err = session.SaveSession(ctx, c, sess)
	if err != nil {
		span.SetStatus(codes.Error, "Error saving session")
		span.RecordError(err)
		return err
	}
	return c.JSON(generateTotpAPIResponse{
		TotpSecret: user.TotpSecret.String,
	})
}

type enableTotpAPIRequest struct {
	Totp string `json:"totp"`
}

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
