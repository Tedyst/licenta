package authbosswebauthn

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/volatiletech/authboss/v3"
)

const (
	PageWebauthn             = "webauthn"
	PageWebauthnSetup        = "webauthn_setup"
	PageWebauthnSetupSuccess = "webauthn_setup_success"

	WebauthnSetupResponseKey = "response"

	WebauthnSessionKey = "webauthn"
)

// webAuthn module
type webAuthn struct {
	Authboss *authboss.Authboss
	WebAuthn *webauthn.WebAuthn

	RegisterOptions []webauthn.RegistrationOption
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

	a.Authboss.Config.Core.Router.Get("/webauthn/begin", abmw(a.Authboss.Core.ErrorHandler.Wrap(a.BeginRegistrationGet)))
	a.Authboss.Config.Core.Router.Post("/webauthn/begin", abmw(a.Authboss.Core.ErrorHandler.Wrap(a.BeginRegistrationPost)))

	a.Authboss.Config.Core.Router.Get("/webauthn/finish", abmw(a.Authboss.Core.ErrorHandler.Wrap(a.FinishRegistrationGet)))
	a.Authboss.Config.Core.Router.Post("/webauthn/finish", abmw(a.Authboss.Core.ErrorHandler.Wrap(a.FinishRegistrationPost)))

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

	webAuthnUser := &webauthnUser{user: authUser, credentials: nil}
	options, session, err := webn.WebAuthn.BeginRegistration(webAuthnUser, webn.RegisterOptions...)
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
	return webn.Authboss.Core.Responder.Respond(w, r, http.StatusOK, PageWebauthnSetupSuccess, data)
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

	credential, err := webn.WebAuthn.FinishRegistration(&webauthnUser{user: authUser, credentials: nil}, session, r)
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
	return webn.Authboss.Core.Responder.Respond(w, r, http.StatusOK, PageWebauthnSetupSuccess, data)
}
