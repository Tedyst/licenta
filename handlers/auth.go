package handlers

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pquerna/otp/totp"
	"github.com/tedyst/licenta/config"
	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/email"
	"github.com/tedyst/licenta/templates/mail"
)

type AuthRequestStatus struct {
	Status int
	Error  error
}

const (
	Success            int = 0
	RequireTOTP        int = 1
	InvalidCredentials int = 2
	Error              int = 3
)

func HandleFirstStepLogin(ctx context.Context, sess *db.Session, username string, password string) (int, error) {
	user, err := config.DatabaseQueries.GetUserByUsernameOrEmail(ctx, username)
	if err != nil {
		return InvalidCredentials, err
	}
	ok, err := user.VerifyPassword(password)
	if err != nil {
		return InvalidCredentials, err
	}
	if !ok {
		return InvalidCredentials, nil
	}
	if user.TotpSecret.Valid {
		sess.Waiting2fa = pgtype.Int8{Int64: user.ID, Valid: true}
		return RequireTOTP, nil
	}
	sess.UserID = pgtype.Int8{Int64: user.ID, Valid: true}
	sess.TotpKey = pgtype.Text{Valid: false}
	return Success, nil
}

func HandleLogout(ctx context.Context, sess *db.Session) {
	sess.UserID = pgtype.Int8{Valid: false}
	sess.TotpKey = pgtype.Text{Valid: false}
}

func HandleTOTPVerify(ctx context.Context, sess *db.Session, user *db.User, totp string) (int, error) {
	ok := user.VerifyTOTP(totp)
	if !ok {
		return InvalidCredentials, nil
	}
	sess.UserID = pgtype.Int8{Int64: user.ID, Valid: true}
	sess.TotpKey = pgtype.Text{Valid: false}
	sess.Waiting2fa = pgtype.Int8{Valid: false}
	return Success, nil
}

func HandleTOTPSetup(ctx context.Context, sess *db.Session, user *db.User) (int, error) {
	secret, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "licenta",
		AccountName: user.Username,
	})
	if err != nil {
		return Error, err
	}
	sess.TotpKey = pgtype.Text{String: secret.Secret(), Valid: true}
	return Success, nil
}

func HandleTOTPSetupVerify(ctx context.Context, sess *db.Session, user *db.User, totp string) (int, error) {
	ok := user.VerifyTOTP(totp)
	if !ok {
		return InvalidCredentials, nil
	}
	err := config.DatabaseQueries.UpdateUserTOTPSecret(ctx, db.UpdateUserTOTPSecretParams{
		ID:         user.ID,
		TotpSecret: user.TotpSecret,
	})
	if err != nil {
		return Error, err
	}
	sess.UserID = pgtype.Int8{Int64: user.ID, Valid: true}
	sess.TotpKey = pgtype.Text{Valid: false}
	sess.Waiting2fa = pgtype.Int8{Valid: false}
	return Success, nil
}

func HandleResetPassword(ctx context.Context, token string, password string) (int, error) {
	u, err := uuid.Parse(token)
	if err != nil {
		return InvalidCredentials, err
	}
	reset_token, err := config.DatabaseQueries.GetResetPasswordToken(ctx, u)
	if err != nil || !reset_token.Valid {
		return InvalidCredentials, err
	}
	if reset_token.CreatedAt.Time.Add(config.ResetPasswordTokenValidity).Before(time.Now()) {
		return InvalidCredentials, nil
	}
	user, err := config.DatabaseQueries.GetUser(ctx, reset_token.UserID.Int64)
	if err != nil {
		return Error, err
	}
	user.SetPassword(password)
	err = config.DatabaseQueries.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		ID:       user.ID,
		Password: user.Password,
	})
	if err != nil {
		return Error, err
	}
	err = config.DatabaseQueries.InvalidateResetPasswordToken(ctx, reset_token.ID)
	if err != nil {
		return Error, err
	}
	return Success, nil
}

func HandleRequestResetPassword(ctx context.Context, username string) (int, error) {
	user, err := config.DatabaseQueries.GetUserByUsernameOrEmail(ctx, username)
	if err != nil {
		// We don't want to leak information about the existence of the user
		return Success, err
	}
	token, err := uuid.NewRandom()
	if err != nil {
		return Error, err
	}
	_, err = config.DatabaseQueries.CreateResetPasswordToken(ctx, db.CreateResetPasswordTokenParams{
		ID:     token,
		UserID: pgtype.Int8{Int64: user.ID, Valid: true},
	})
	if err != nil {
		return Error, err
	}
	html := mail.SendResetPasswordHTML(user.Email, token.String(), config.BaseURL)
	text := mail.SendResetPasswordText(user.Email, token.String(), config.BaseURL)
	err = email.SendMultipartEmail(user.Email, html, text)
	if err != nil {
		return Error, err
	}
	return Success, nil
}
