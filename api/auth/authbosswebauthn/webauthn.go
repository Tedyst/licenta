package authbosswebauthn

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/volatiletech/authboss/v3"
	"github.com/volatiletech/authboss/v3/otp/twofactor/totp2fa"
)

const (
	PageWebauthn            = "webauthn"
	PageWebauthnSetup       = "webauthn_setup"
	PageWebauthnSetupFinish = "webauthn_setup_finish"
	PageWebauthnLogin       = "webauthn_login"
	PageWebauthnLoginFinish = "webauthn_login_finish"

	PageChoose2FA = "choose_2fa"

	WebauthnSetupResponseKey = "response"
	WebauthnLoginResponseKey = "response"

	WebauthnSessionKey      = "webauthn"
	WebauthnLoginSessionKey = "webauthn_login"
)

// webAuthn module
type webAuthn struct {
	Authboss *authboss.Authboss
	WebAuthn *webauthn.WebAuthn

	RegisterOptions []webauthn.RegistrationOption
}

type saveSessionStruct struct {
	Discoverable bool                  `json:"discoverable"`
	Data         *webauthn.SessionData `json:"data"`
	PID          string                `json:"pid"`
}

func New(ab *authboss.Authboss, webauthn *webauthn.WebAuthn, registrationOptions []webauthn.RegistrationOption) *webAuthn {
	return &webAuthn{
		WebAuthn:        webauthn,
		Authboss:        ab,
		RegisterOptions: registrationOptions,
	}
}

// Init module
func (a *webAuthn) Setup() (err error) {
	if err = a.Authboss.Config.Core.ViewRenderer.Load(PageWebauthn); err != nil {
		return err
	}

	var unauthedResponse authboss.MWRespondOnFailure
	if a.Authboss.Config.Modules.ResponseOnUnauthed != 0 {
		unauthedResponse = a.Authboss.Config.Modules.ResponseOnUnauthed
	} else if a.Authboss.Config.Modules.RoutesRedirectOnUnauthed {
		unauthedResponse = authboss.RespondRedirect
	}
	abmw := authboss.MountedMiddleware2(a.Authboss, true, authboss.RequireFullAuth, unauthedResponse)

	a.Authboss.Config.Core.Router.Get("/webauthn/register/begin", abmw(a.Authboss.Core.ErrorHandler.Wrap(a.BeginRegistrationGet)))
	a.Authboss.Config.Core.Router.Post("/webauthn/register/begin", abmw(a.Authboss.Core.ErrorHandler.Wrap(a.BeginRegistrationPost)))

	a.Authboss.Config.Core.Router.Get("/webauthn/register/finish", abmw(a.Authboss.Core.ErrorHandler.Wrap(a.FinishRegistrationGet)))
	a.Authboss.Config.Core.Router.Post("/webauthn/register/finish", abmw(a.Authboss.Core.ErrorHandler.Wrap(a.FinishRegistrationPost)))

	a.Authboss.Config.Core.Router.Get("/webauthn/login/begin", a.Authboss.Core.ErrorHandler.Wrap(a.BeginLoginGet))
	a.Authboss.Config.Core.Router.Post("/webauthn/login/begin", a.Authboss.Core.ErrorHandler.Wrap(a.BeginLoginPost))

	a.Authboss.Config.Core.Router.Get("/webauthn/login/finish", a.Authboss.Core.ErrorHandler.Wrap(a.FinishLoginGet))
	a.Authboss.Config.Core.Router.Post("/webauthn/login/finish", a.Authboss.Core.ErrorHandler.Wrap(a.FinishLoginPost))

	a.Authboss.Events.Before(authboss.EventAuthHijack, a.HijackAuth)

	return nil
}

// BeginRegistrationGet renders the webauthn page
func (webn *webAuthn) BeginRegistrationGet(w http.ResponseWriter, r *http.Request) error {
	data := authboss.HTMLData{}
	if redir := r.URL.Query().Get(authboss.FormValueRedirect); len(redir) != 0 {
		data[authboss.FormValueRedirect] = redir
	}
	return webn.Authboss.Core.Responder.Respond(w, r, http.StatusOK, PageWebauthn, data)
}

// BeginRegistrationPost starts the webauthn registration process
func (webn *webAuthn) BeginRegistrationPost(w http.ResponseWriter, r *http.Request) error {
	logger := webn.Authboss.RequestLogger(r)

	abUser, err := webn.Authboss.CurrentUser(r)
	if err != nil {
		return err
	}

	authUser := MustBeWebauthnUser(abUser)
	storer := MustBeWebauthnStorer(webn.Authboss.Config.Storage.Server)

	currentCreds, err := storer.GetWebauthnCredentials(r.Context(), authUser.GetPID())
	if err != nil {
		return err
	}

	excludeCreds := make([]protocol.CredentialDescriptor, len(currentCreds))
	for i, cred := range currentCreds {
		excludeCreds[i] = protocol.CredentialDescriptor{
			Type:            protocol.PublicKeyCredentialType,
			CredentialID:    cred.ID,
			Transport:       cred.Transport,
			AttestationType: cred.AttestationType,
		}
	}

	webAuthnUser := &webauthnUser{user: authUser, credentials: nil}
	options, session, err := webn.WebAuthn.BeginRegistration(webAuthnUser, webauthn.WithExclusions(excludeCreds), webauthn.WithResidentKeyRequirement(protocol.ResidentKeyRequirementPreferred))
	if err != nil {
		return err
	}

	sessionData, err := json.Marshal(&session)
	if err != nil {
		return err
	}
	authboss.PutSession(w, WebauthnSessionKey, string(sessionData))

	logger.Infof("Starting registration using webauthn for %s", authUser.GetPID())

	data := authboss.HTMLData{
		WebauthnSetupResponseKey: options.Response,
	}
	return webn.Authboss.Core.Responder.Respond(w, r, http.StatusOK, PageWebauthnSetup, data)
}

// FinishRegistrationGet renders the webauthn success page
func (webn *webAuthn) FinishRegistrationGet(w http.ResponseWriter, r *http.Request) error {
	data := authboss.HTMLData{}
	return webn.Authboss.Core.Responder.Respond(w, r, http.StatusOK, PageWebauthnSetupFinish, data)
}

// FinishRegistrationPost finishes the webauthn registration process
func (webn *webAuthn) FinishRegistrationPost(w http.ResponseWriter, r *http.Request) error {
	logger := webn.Authboss.RequestLogger(r)

	storer := MustBeWebauthnStorer(webn.Authboss.Config.Storage.Server)

	abUser, err := webn.Authboss.CurrentUser(r)
	if err != nil {
		return err
	}

	authUser := MustBeWebauthnUser(abUser)

	abSession, ok := authboss.GetSession(r, WebauthnSessionKey)
	if !ok {
		return errors.New("webauthn session not found")
	}

	var session webauthn.SessionData
	if err := json.Unmarshal([]byte(abSession), &session); err != nil {
		return err
	}

	reader, err := webn.Authboss.Core.BodyReader.Read(PageWebauthnSetupFinish, r)
	if err != nil {
		return err
	}
	wnReader := MustBeFinishWebauthnCreationUserValuer(reader)
	creationCredential := wnReader.GetCreationCredential()

	credential, err := webn.WebAuthn.CreateCredential(&webauthnUser{user: authUser, credentials: nil}, session, &creationCredential)
	if err != nil {
		return err
	}

	if err := storer.CreateWebauthnCredential(r.Context(), authUser.GetPID(), *credential); err != nil {
		return err
	}
	if err = webn.Authboss.Config.Storage.Server.Save(r.Context(), abUser); err != nil {
		return err
	}

	logger.Infof("Finished registration using webauthn for %s", authUser.GetPID())

	data := authboss.HTMLData{}
	return webn.Authboss.Core.Responder.Respond(w, r, http.StatusOK, PageWebauthnSetupFinish, data)
}

func (webn *webAuthn) BeginLoginGet(w http.ResponseWriter, r *http.Request) error {
	data := authboss.HTMLData{}
	if redir := r.URL.Query().Get(authboss.FormValueRedirect); len(redir) != 0 {
		data[authboss.FormValueRedirect] = redir
	}
	return webn.Authboss.Core.Responder.Respond(w, r, http.StatusOK, PageWebauthnLogin, data)
}

func (webn *webAuthn) BeginLoginPost(w http.ResponseWriter, r *http.Request) error {
	validatable, err := webn.Authboss.Core.BodyReader.Read(PageWebauthnLogin, r)
	if err != nil {
		return err
	}

	userValuer := MustBeWebauthnUserValuer(validatable)

	if userValuer.GetPID() == "" {
		assertion, session, err := webn.WebAuthn.BeginDiscoverableLogin()
		if err != nil {
			return err
		}

		sessionData, err := json.Marshal(&saveSessionStruct{
			Discoverable: true,
			Data:         session,
		})
		if err != nil {
			return err
		}

		authboss.PutSession(w, WebauthnLoginSessionKey, string(sessionData))

		data := authboss.HTMLData{
			WebauthnLoginResponseKey: assertion.Response,
		}

		return webn.Authboss.Core.Responder.Respond(w, r, http.StatusOK, PageWebauthnLogin, data)
	}

	user, err := webn.Authboss.Config.Storage.Server.Load(r.Context(), userValuer.GetPID())
	if err != nil {
		return err
	}

	webnUser := MustBeWebauthnUser(user)
	storer := MustBeWebauthnStorer(webn.Authboss.Config.Storage.Server)

	creds, err := storer.GetWebauthnCredentials(r.Context(), webnUser.GetPID())
	if err != nil {
		return err
	}

	assertion, session, err := webn.WebAuthn.BeginLogin(&webauthnUser{user: webnUser, credentials: creds})
	if err != nil {
		return err
	}

	sessionData, err := json.Marshal(&saveSessionStruct{
		Discoverable: false,
		Data:         session,
		PID:          userValuer.GetPID(),
	})
	if err != nil {
		return err
	}

	authboss.PutSession(w, WebauthnLoginSessionKey, string(sessionData))
	authboss.DelSession(w, WebauthnSessionKey)

	data := authboss.HTMLData{
		WebauthnLoginResponseKey: assertion.Response,
	}
	return webn.Authboss.Core.Responder.Respond(w, r, http.StatusOK, PageWebauthnLogin, data)
}

func (webn *webAuthn) FinishLoginGet(w http.ResponseWriter, r *http.Request) error {
	data := authboss.HTMLData{}
	if redir := r.URL.Query().Get(authboss.FormValueRedirect); len(redir) != 0 {
		data[authboss.FormValueRedirect] = redir
	}
	return webn.Authboss.Core.Responder.Respond(w, r, http.StatusOK, PageWebauthnLogin, data)
}

func (webn *webAuthn) FinishLoginPost(w http.ResponseWriter, r *http.Request) error {
	validatable, err := webn.Authboss.Core.BodyReader.Read(PageWebauthnLoginFinish, r)
	if err != nil {
		return err
	}

	userValuer := MustBeFinishWebauthnLoginUserValuer(validatable)
	credential := userValuer.GetCredentialAssertion()

	abSession, ok := authboss.GetSession(r, WebauthnLoginSessionKey)
	if !ok {
		return errors.New("webauthn session not found")
	}

	var session saveSessionStruct
	if err := json.Unmarshal([]byte(abSession), &session); err != nil {
		return err
	}

	storer := MustBeWebauthnStorer(webn.Authboss.Config.Storage.Server)

	var user authboss.User
	if session.Discoverable {
		discoverableUserHandler := func(rawID, userHandle []byte) (webauthn.User, error) {
			user, err := webn.Authboss.Config.Storage.Server.Load(r.Context(), string(userHandle))
			if err != nil {
				return nil, err
			}

			creds, err := storer.GetWebauthnCredentials(r.Context(), user.GetPID())
			if err != nil {
				return nil, err
			}

			return &webauthnUser{
				user:        MustBeWebauthnUser(user),
				credentials: creds,
			}, nil
		}

		foundCred, err := webn.WebAuthn.ValidateDiscoverableLogin(discoverableUserHandler, *session.Data, &credential)
		if err != nil {
			return err
		}

		if foundCred == nil {
			return errors.New("no credential found")
		}

		user, err = storer.GetUserByCredentialID(r.Context(), foundCred.ID)
		if err != nil {
			return err
		}
	} else {
		creds, err := storer.GetWebauthnCredentials(r.Context(), session.PID)
		if err != nil {
			return err
		}

		user, err = webn.Authboss.Config.Storage.Server.Load(r.Context(), session.PID)
		if err != nil {
			return err
		}

		foundCred, err := webn.WebAuthn.ValidateLogin(&webauthnUser{
			user:        MustBeWebauthnUser(user),
			credentials: creds,
		}, *session.Data, &credential)
		if err != nil {
			return err
		}

		if foundCred == nil {
			return errors.New("no credential found")
		}
	}

	authboss.PutSession(w, authboss.SessionKey, user.GetPID())
	authboss.DelSession(w, authboss.SessionHalfAuthKey)
	authboss.DelSession(w, WebauthnSessionKey)
	authboss.DelSession(w, WebauthnLoginSessionKey)

	ro := authboss.RedirectOptions{
		Code:             http.StatusTemporaryRedirect,
		RedirectPath:     webn.Authboss.Paths.AuthLoginOK,
		FollowRedirParam: true,
	}
	return webn.Authboss.Core.Redirector.Redirect(w, r, ro)
}

// HijackAuth stores the user's pid in a special temporary session variable
// and redirects them to the validation endpoint.
func (webn *webAuthn) HijackAuth(w http.ResponseWriter, r *http.Request, handled bool) (bool, error) {
	if handled {
		return false, nil
	}

	storer := MustBeWebauthnStorer(webn.Authboss.Config.Storage.Server)
	user, ok := r.Context().Value(authboss.CTXKeyUser).(totp2fa.User)
	if !ok {
		return false, nil
	}

	rmIntf := r.Context().Value(authboss.CTXKeyValues)
	if rmIntf == nil {
		authboss.PutSession(w, "should_remember", "false")
	} else if rm, ok := rmIntf.(authboss.RememberValuer); !ok || !rm.GetShouldRemember() {
		authboss.PutSession(w, "should_remember", "false")
	} else {
		authboss.PutSession(w, "should_remember", "true")
	}

	creds, err := storer.GetWebauthnCredentials(r.Context(), user.GetPID())
	if err != nil && err != authboss.ErrUserNotFound {
		return false, err
	}

	if len(creds) == 0 && user.GetTOTPSecretKey() == "" {
		return false, nil
	}

	authboss.PutSession(w, totp2fa.SessionTOTPPendingPID, user.GetPID())

	data := authboss.HTMLData{
		"totp":     user.GetTOTPSecretKey() != "",
		"webauthn": len(creds) != 0,
		"status":   "not validated",
	}

	return true, webn.Authboss.Core.Responder.Respond(w, r, http.StatusOK, PageChoose2FA, data)
}
