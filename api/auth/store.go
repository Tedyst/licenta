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
		Username: sql.NullString{String: user.(*authbossUser).user.Username, Valid: user.(*authbossUser).user.Username != ""},
		Email:    sql.NullString{String: user.(*authbossUser).user.Email, Valid: user.(*authbossUser).user.Email != ""},
		Password: sql.NullString{String: user.(*authbossUser).user.Password, Valid: user.(*authbossUser).user.Password != ""},
		ID:       user.(*authbossUser).user.ID,
	})
}

func (a *authbossStorer) New(ctx context.Context) authboss.User {
	return &authbossUser{
		user: &models.User{},
	}
}

func (a *authbossStorer) Create(ctx context.Context, user authboss.User) error {
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
