package auth

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v5"
	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/models"
	"github.com/volatiletech/authboss/v3"
)

type authbossStorer struct {
	querier db.TransactionQuerier
}

var _ authboss.ServerStorer = (*authbossStorer)(nil)

func newAuthbossStorer(querier db.TransactionQuerier) *authbossStorer {
	return &authbossStorer{
		querier: querier,
	}
}

func (a *authbossStorer) Load(ctx context.Context, key string) (authboss.User, error) {
	user, err := a.querier.GetUserByUsernameOrEmail(ctx, queries.GetUserByUsernameOrEmailParams{
		Username: key,
		Email:    key,
	})
	if err != nil && err != pgx.ErrNoRows {
		return nil, err
	}
	if err == pgx.ErrNoRows {
		return nil, authboss.ErrUserNotFound
	}

	return &authbossUser{
		user: user,
	}, nil
}

func (a *authbossStorer) Save(ctx context.Context, user authboss.User) error {
	return a.querier.UpdateUser(ctx, queries.UpdateUserParams{
		Username:          sql.NullString{String: user.(*authbossUser).user.Username, Valid: user.(*authbossUser).user.Username != ""},
		Email:             sql.NullString{String: user.(*authbossUser).user.Email, Valid: user.(*authbossUser).user.Email != ""},
		Password:          sql.NullString{String: user.(*authbossUser).user.Password, Valid: user.(*authbossUser).user.Password != ""},
		ID:                user.(*authbossUser).user.ID,
		RecoveryCodes:     user.(*authbossUser).user.RecoveryCodes,
		TotpSecret:        user.(*authbossUser).user.TotpSecret,
		RecoverSelector:   user.(*authbossUser).user.RecoverSelector,
		RecoverVerifier:   user.(*authbossUser).user.RecoverVerifier,
		RecoverExpiry:     user.(*authbossUser).user.RecoverExpiry,
		LoginAttemptCount: sql.NullInt32{Int32: user.(*authbossUser).user.LoginAttemptCount, Valid: true},
		LoginLastAttempt:  user.(*authbossUser).user.LoginLastAttempt,
		Locked:            user.(*authbossUser).user.Locked,
	})
}

func (a *authbossStorer) New(ctx context.Context) authboss.User {
	return &authbossUser{
		user: &models.User{},
	}
}

func (a *authbossStorer) Create(ctx context.Context, user authboss.User) error {
	if user.(*authbossUser).user.Username == "" || user.(*authbossUser).user.Email == "" || user.(*authbossUser).user.Password == "" {
		return authboss.ErrUserNotFound
	}

	_, err := a.querier.GetUserByUsernameOrEmail(ctx, queries.GetUserByUsernameOrEmailParams{
		Username: user.(*authbossUser).user.Username,
		Email:    user.(*authbossUser).user.Email,
	})
	if err != nil && err != pgx.ErrNoRows {
		return err
	}
	if err == nil {
		return authboss.ErrUserFound
	}

	_, err = a.querier.CreateUser(ctx, queries.CreateUserParams{
		Username: user.(*authbossUser).user.Username,
		Email:    user.(*authbossUser).user.Email,
		Password: user.(*authbossUser).user.Password,
	})
	return err
}

func (a *authbossStorer) AddRememberToken(ctx context.Context, pid, token string) error {
	user, err := a.querier.GetUserByUsernameOrEmail(ctx, queries.GetUserByUsernameOrEmailParams{
		Username: pid,
		Email:    pid,
	})
	if err != nil {
		return err
	}
	_, err = a.querier.CreateRememberMeToken(ctx, queries.CreateRememberMeTokenParams{
		UserID: user.ID,
		Token:  token,
	})
	return err
}

func (a *authbossStorer) DelRememberTokens(ctx context.Context, pid string) error {
	user, err := a.querier.GetUserByUsernameOrEmail(ctx, queries.GetUserByUsernameOrEmailParams{
		Username: pid,
		Email:    pid,
	})
	if err != nil {
		return err
	}
	return a.querier.DeleteRememberMeTokensForUser(ctx, user.ID)
}

func (a *authbossStorer) UseRememberToken(ctx context.Context, pid, token string) error {
	user, err := a.querier.GetUserByUsernameOrEmail(ctx, queries.GetUserByUsernameOrEmailParams{
		Username: pid,
		Email:    pid,
	})
	if err != nil {
		return err
	}
	return a.querier.DeleteRememberMeTokenByUserAndToken(ctx, queries.DeleteRememberMeTokenByUserAndTokenParams{
		UserID: user.ID,
		Token:  token,
	})
}

func (a *authbossStorer) LoadByRecoverSelector(ctx context.Context, selector string) (authboss.RecoverableUser, error) {
	user, err := a.querier.GetUserByRecoverSelector(ctx, sql.NullString{String: selector, Valid: true})
	if err != nil && err != pgx.ErrNoRows {
		return nil, err
	}
	if err == pgx.ErrNoRows {
		return nil, authboss.ErrUserNotFound
	}

	return &authbossUser{
		user: user,
	}, nil
}
