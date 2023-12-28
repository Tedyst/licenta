package authbosswebauthn

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/volatiletech/authboss/v3"
)

const (
	PageWebauthn             = "webauthn"
	PageWebauthnSetup        = "webauthn_setup"
	PageWebauthnSetupSuccess = "webauthn_setup_success"
	PageWebauthnLogin        = "webauthn_login"

	WebauthnSetupResponseKey = "response"
	WebauthnLoginResponseKey = "response"

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

	a.Authboss.Config.Core.Router.Get("/webauthn/register/begin", abmw(a.Authboss.Core.ErrorHandler.Wrap(a.BeginRegistrationGet)))
	a.Authboss.Config.Core.Router.Post("/webauthn/register/begin", abmw(a.Authboss.Core.ErrorHandler.Wrap(a.BeginRegistrationPost)))

	a.Authboss.Config.Core.Router.Get("/webauthn/register/finish", abmw(a.Authboss.Core.ErrorHandler.Wrap(a.FinishRegistrationGet)))
	a.Authboss.Config.Core.Router.Post("/webauthn/register/finish", abmw(a.Authboss.Core.ErrorHandler.Wrap(a.FinishRegistrationPost)))

	a.Authboss.Config.Core.Router.Get("/webauthn/login/begin", a.Authboss.Core.ErrorHandler.Wrap(a.BeginLoginGet))
	a.Authboss.Config.Core.Router.Post("/webauthn/login/begin", a.Authboss.Core.ErrorHandler.Wrap(a.BeginLoginPost))

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
	options, session, err := webn.WebAuthn.BeginRegistration(webAuthnUser, webauthn.WithExclusions(excludeCreds))
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

	reader, err := webn.Authboss.Core.BodyReader.Read(PageWebauthnSetup, r)
	if err != nil {
		return err
	}
	wnReader := MustBeWebauthnCreationUserValuer(reader)
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
	return webn.Authboss.Core.Responder.Respond(w, r, http.StatusOK, PageWebauthnSetupSuccess, data)
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

		sessionData, err := json.Marshal(&session)
		if err != nil {
			return err
		}

		authboss.PutSession(w, WebauthnSessionKey, string(sessionData))

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

	sessionData, err := json.Marshal(&session)
	if err != nil {
		return err
	}

	authboss.PutSession(w, WebauthnSessionKey, string(sessionData))

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

// func (webn *webAuthn) FinishLoginPost(w http.ResponseWriter, r *http.Request) error {
// 	webn.WebAuthn.FinishDiscoverableLogin()
// }
