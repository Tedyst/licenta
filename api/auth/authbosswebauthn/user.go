package authbosswebauthn

import (
	"context"
	"fmt"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/volatiletech/authboss/v3"
)

type Credential = webauthn.Credential

// WebauthnUser is an interface that can be implemented by a user.
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
	GetUserByCredentialID(ctx context.Context, credentialID []byte) (authboss.User, error)
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

type WebauthnUserValuer interface {
	authboss.Validator

	GetPID() string
}

func MustBeWebauthnUserValuer(validator authboss.Validator) WebauthnUserValuer {
	if au, ok := validator.(WebauthnUserValuer); ok {
		return au
	}
	panic(fmt.Sprintf("could not upgrade validator to an authable validator, type: %T", validator))
}

type FinishWebauthnCreationUserValuer interface {
	authboss.Validator

	GetCreationCredential() protocol.ParsedCredentialCreationData
}

func MustBeFinishWebauthnCreationUserValuer(validator authboss.Validator) FinishWebauthnCreationUserValuer {
	if au, ok := validator.(FinishWebauthnCreationUserValuer); ok {
		return au
	}
	panic(fmt.Sprintf("could not upgrade validator to an authable validator, type: %T", validator))
}

type FinishWebauthnLoginUserValuer interface {
	authboss.Validator

	GetCredentialAssertion() protocol.ParsedCredentialAssertionData
}

func MustBeFinishWebauthnLoginUserValuer(validator authboss.Validator) FinishWebauthnLoginUserValuer {
	if au, ok := validator.(FinishWebauthnLoginUserValuer); ok {
		return au
	}
	panic(fmt.Sprintf("could not upgrade validator to an authable validator, type: %T", validator))
}
