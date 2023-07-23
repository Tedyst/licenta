package handlers

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pquerna/otp/totp"
	"github.com/tedyst/licenta/config"
	"github.com/tedyst/licenta/db"
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
