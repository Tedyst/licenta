package auth

import (
	"time"

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
	a.user.RecoveryCodes.Valid = codes != ""
}

func (a *authbossUser) GetTOTPSecretKey() string {
	return a.user.TotpSecret.String
}

func (a *authbossUser) PutTOTPSecretKey(key string) {
	a.user.TotpSecret.String = key
	a.user.TotpSecret.Valid = key != ""
}

func (a *authbossUser) GetRecoverSelector() string {
	return a.user.RecoverSelector.String
}

func (a *authbossUser) PutRecoverSelector(selector string) {
	a.user.RecoverSelector.String = selector
	a.user.RecoverSelector.Valid = selector != ""
}

func (a *authbossUser) GetRecoverVerifier() string {
	return a.user.RecoverVerifier.String
}

func (a *authbossUser) PutRecoverVerifier(verifier string) {
	a.user.RecoverVerifier.String = verifier
	a.user.RecoverVerifier.Valid = verifier != ""
}

func (a *authbossUser) GetRecoverExpiry() time.Time {
	return a.user.RecoverExpiry.Time
}

func (a *authbossUser) PutRecoverExpiry(expiry time.Time) {
	a.user.RecoverExpiry.Time = expiry
	a.user.RecoverExpiry.Valid = expiry != time.Time{}
}

func (a *authbossUser) GetAttemptCount() (attempts int) {
	return int(a.user.LoginAttemptCount)
}

func (a *authbossUser) PutAttemptCount(attempts int) {
	a.user.LoginAttemptCount = int32(attempts)
}

func (a *authbossUser) GetLastAttempt() (last time.Time) {
	return a.user.LoginLastAttempt.Time
}

func (a *authbossUser) PutLastAttempt(last time.Time) {
	a.user.LoginLastAttempt.Time = last
	a.user.LoginLastAttempt.Valid = last != time.Time{}
}

func (a *authbossUser) GetLocked() (locked time.Time) {
	return a.user.Locked.Time
}

func (a *authbossUser) PutLocked(locked time.Time) {
	a.user.Locked.Time = locked
	a.user.Locked.Valid = locked != time.Time{}
}
