package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/jackc/pgx/v5"
	"github.com/tedyst/authbosswebauthn"
	"github.com/tedyst/licenta/cache"
	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/db/queries"
	"github.com/volatiletech/authboss/v3"
)

type authbossStorer struct {
	cache   cache.CacheProvider[queries.User]
	querier db.TransactionQuerier
}

var _ authboss.ServerStorer = (*authbossStorer)(nil)

func newAuthbossStorer(querier db.TransactionQuerier, cache cache.CacheProvider[queries.User]) *authbossStorer {
	return &authbossStorer{
		querier: querier,
		cache:   cache,
	}
}

func (a *authbossStorer) Load(ctx context.Context, key string) (authboss.User, error) {
	cachedUser, ok, err := a.cache.Get(key)
	if err == nil && ok {
		return &authbossUser{
			user: &cachedUser,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting user from cache: %w", err)
	}

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
	u, ok := user.(*authbossUser)
	if !ok {
		return errors.New("invalid user type")
	}

	err := a.cache.Set(u.user.Username, *u.user)
	if err != nil {
		return fmt.Errorf("error setting user in cache: %w", err)
	}

	return a.querier.UpdateUser(ctx, queries.UpdateUserParams{
		ID:                u.user.ID,
		Username:          u.user.Username,
		Email:             u.user.Email,
		Password:          u.user.Password,
		RecoveryCodes:     u.user.RecoveryCodes,
		TotpSecret:        u.user.TotpSecret,
		RecoverSelector:   u.user.RecoverSelector,
		RecoverVerifier:   u.user.RecoverVerifier,
		RecoverExpiry:     u.user.RecoverExpiry,
		LoginAttemptCount: u.user.LoginAttemptCount,
		LoginLastAttempt:  u.user.LoginLastAttempt,
		Locked:            u.user.Locked,
		ConfirmSelector:   u.user.ConfirmSelector,
		ConfirmVerifier:   u.user.ConfirmVerifier,
		Confirmed:         u.user.Confirmed,
	})
}

func (a *authbossStorer) New(ctx context.Context) authboss.User {
	return &authbossUser{
		user: &queries.User{},
	}
}

func (a *authbossStorer) Create(ctx context.Context, user authboss.User) error {
	u, ok := user.(*authbossUser)
	if !ok {
		return errors.New("invalid user type")
	}
	if u.user.Username == "" || u.user.Email == "" || u.user.Password == "" {
		return authboss.ErrUserNotFound
	}

	_, err := a.querier.GetUserByUsernameOrEmail(ctx, queries.GetUserByUsernameOrEmailParams{
		Username: u.user.Username,
		Email:    u.user.Email,
	})
	if err != nil && err != pgx.ErrNoRows {
		return err
	}
	if err == nil {
		return authboss.ErrUserFound
	}

	newUser, err := a.querier.CreateUser(ctx, queries.CreateUserParams{
		Username:          u.user.Username,
		Email:             u.user.Email,
		Password:          u.user.Password,
		RecoveryCodes:     u.user.RecoveryCodes,
		TotpSecret:        u.user.TotpSecret,
		RecoverSelector:   u.user.RecoverSelector,
		RecoverVerifier:   u.user.RecoverVerifier,
		RecoverExpiry:     u.user.RecoverExpiry,
		LoginAttemptCount: u.user.LoginAttemptCount,
		LoginLastAttempt:  u.user.LoginLastAttempt,
		Locked:            u.user.Locked,
		ConfirmSelector:   u.user.ConfirmSelector,
		ConfirmVerifier:   u.user.ConfirmVerifier,
		Confirmed:         u.user.Confirmed,
	})
	u.user = newUser

	err2 := a.cache.Set(u.user.Username, *u.user)

	return errors.Join(err, err2)
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

func (a *authbossStorer) LoadByConfirmSelector(ctx context.Context, selector string) (authboss.ConfirmableUser, error) {
	user, err := a.querier.GetUserByConfirmSelector(ctx, sql.NullString{String: selector, Valid: true})
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

func (a *authbossStorer) GetWebauthnCredentials(ctx context.Context, pid string) ([]authbosswebauthn.Credential, error) {
	user, err := a.querier.GetUserByUsernameOrEmail(ctx, queries.GetUserByUsernameOrEmailParams{
		Username: pid,
		Email:    pid,
	})
	if err != nil {
		return nil, err
	}

	creds, err := a.querier.GetWebauthnCredentialsByUserID(ctx, user.ID)
	if err != nil && err != pgx.ErrNoRows {
		return nil, err
	}
	if err == pgx.ErrNoRows {
		return nil, authboss.ErrUserNotFound
	}

	credentials := make([]authbosswebauthn.Credential, len(creds))
	for i, cred := range creds {
		transport := make([]protocol.AuthenticatorTransport, len(cred.Transport))
		for i, t := range cred.Transport {
			transport[i] = protocol.AuthenticatorTransport(t)
		}

		credentials[i] = authbosswebauthn.Credential{
			ID:              cred.CredentialID,
			PublicKey:       cred.PublicKey,
			AttestationType: cred.AttestationType,
			Transport:       transport,
			Flags: webauthn.CredentialFlags{
				UserPresent:    cred.UserPresent,
				UserVerified:   cred.UserVerified,
				BackupEligible: cred.BackupEligible,
				BackupState:    cred.BackupState,
			},
			Authenticator: webauthn.Authenticator{
				AAGUID:       cred.AaGuid,
				SignCount:    uint32(cred.SignCount),
				CloneWarning: cred.CloneWarning,
				Attachment:   protocol.AuthenticatorAttachment(cred.Attachment),
			},
		}
	}

	return credentials, nil
}

func (a *authbossStorer) CreateWebauthnCredential(ctx context.Context, pid string, credential authbosswebauthn.Credential) error {
	user, err := a.querier.GetUserByUsernameOrEmail(ctx, queries.GetUserByUsernameOrEmailParams{
		Username: pid,
		Email:    pid,
	})
	if err != nil {
		return err
	}

	transports := make([]string, len(credential.Transport))
	for i, transport := range credential.Transport {
		transports[i] = string(transport)
	}

	_, err = a.querier.CreateWebauthnCredential(ctx, queries.CreateWebauthnCredentialParams{
		UserID:          user.ID,
		CredentialID:    credential.ID,
		PublicKey:       credential.PublicKey,
		AttestationType: credential.AttestationType,
		Transport:       transports,
		UserPresent:     credential.Flags.UserPresent,
		UserVerified:    credential.Flags.UserVerified,
		BackupEligible:  credential.Flags.BackupEligible,
		BackupState:     credential.Flags.BackupState,
		AaGuid:          credential.Authenticator.AAGUID,
		SignCount:       int32(credential.Authenticator.SignCount),
		CloneWarning:    credential.Authenticator.CloneWarning,
		Attachment:      string(credential.Authenticator.Attachment),
	})
	return err
}

func (a *authbossStorer) GetUserByCredentialID(ctx context.Context, credentialID []byte) (authboss.User, error) {
	user, err := a.querier.GetUserByWebauthnCredentialID(ctx, credentialID)
	if err != nil && err != pgx.ErrNoRows {
		return nil, err
	}
	if err == pgx.ErrNoRows {
		return nil, authboss.ErrUserNotFound
	}

	return &authbossUser{
		user: &user.User,
	}, nil
}
