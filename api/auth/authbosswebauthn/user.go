package authbosswebauthn

import (
	"context"
	"fmt"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/volatiletech/authboss/v3"
)

type Credential = webauthn.Credential

// WebauthnUser is an interface that can be implemented by a user
type WebauthnUser interface {
	authboss.User

	GetName() (name string)
	GetDisplayName() (displayName string)
	GetIcon() (icon string)
}

type WebauthnStorer interface {
	authboss.ServerStorer

	GetWebauthnCredentials(ctx context.Context, pid string) ([]Credential, error)
	CreateWebauthnCredential(ctx context.Context, pid string, credential Credential) error
}

func MustBeWebauthnUser(user authboss.User) WebauthnUser {
	if au, ok := user.(WebauthnUser); ok {
		return au
	}
	panic(fmt.Sprintf("could not upgrade user to an authable user, type: %T", user))
}

func MustBeWebauthnStorer(storer authboss.ServerStorer) WebauthnStorer {
	if au, ok := storer.(WebauthnStorer); ok {
		return au
	}
	panic(fmt.Sprintf("could not upgrade storer to an authable storer, type: %T", storer))
}

type webauthnUser struct {
	user WebauthnUser

	credentials []Credential
}

var _ webauthn.User = (*webauthnUser)(nil)

func (w *webauthnUser) WebAuthnID() []byte {
	return []byte(w.user.GetPID())
}

func (w *webauthnUser) WebAuthnName() string {
	return w.user.GetName()
}

func (w *webauthnUser) WebAuthnDisplayName() string {
	return w.user.GetDisplayName()
}

func (w *webauthnUser) WebAuthnIcon() string {
	return w.user.GetIcon()
}

func (w *webauthnUser) WebAuthnCredentials() []webauthn.Credential {
	return w.credentials
}
