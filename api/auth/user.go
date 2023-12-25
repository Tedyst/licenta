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
