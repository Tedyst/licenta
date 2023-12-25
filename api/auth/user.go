package auth

import (
	"github.com/tedyst/licenta/models"
	"github.com/volatiletech/authboss/v3"
)

type authbossUser struct {
	user *models.User
}

var _ authboss.User = (*authbossUser)(nil)

func (a *authbossUser) GetPID() string {
	return a.user.Username
}

func (a *authbossUser) PutPID(pid string) {
	a.user.Username = pid
}

func (a *authbossUser) GetPassword() string {
	return a.user.Password
}

func (a *authbossUser) PutPassword(password string) {
	a.user.Password = password
}

func (a *authbossUser) GetArbitrary() map[string]string {
	return nil
}

func (a *authbossUser) PutArbitrary(values map[string]string) {
	if val, ok := values["username"]; ok {
		a.user.Username = val
	}
	if val, ok := values["email"]; ok {
		a.user.Email = val
	}
}

func (a *authbossUser) GetEmail() string {
	return a.user.Email
}

func (a *authbossUser) PutEmail(email string) {
	a.user.Email = email
}

func (a *authbossUser) GetRecoveryCodes() string {
	return a.user.RecoveryCodes.String
}

func (a *authbossUser) PutRecoveryCodes(codes string) {
	a.user.RecoveryCodes.String = codes
	a.user.RecoveryCodes.Valid = true
}

func (a *authbossUser) GetTOTPSecretKey() string {
	return a.user.TotpSecret.String
}

func (a *authbossUser) PutTOTPSecretKey(key string) {
	a.user.TotpSecret.String = key
	a.user.TotpSecret.Valid = true
}
