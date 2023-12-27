package auth

import (
	"context"
	"net/http"
	"os"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/tedyst/licenta/api/auth/authbosswebauthn"
	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/models"
	abclientstate "github.com/volatiletech/authboss-clientstate"
	"github.com/volatiletech/authboss/v3"
	_ "github.com/volatiletech/authboss/v3/auth"
	"github.com/volatiletech/authboss/v3/confirm"
	"github.com/volatiletech/authboss/v3/defaults"
	"github.com/volatiletech/authboss/v3/lock"
	_ "github.com/volatiletech/authboss/v3/logout"
	"github.com/volatiletech/authboss/v3/otp/twofactor"
	"github.com/volatiletech/authboss/v3/otp/twofactor/totp2fa"
	_ "github.com/volatiletech/authboss/v3/recover"
	_ "github.com/volatiletech/authboss/v3/register"
	"github.com/volatiletech/authboss/v3/remember"
)

const sessionCookieName = "session"

type requestStorer struct{}

type authenticationProvider struct {
	authboss *authboss.Authboss
	querier  db.TransactionQuerier
}

func NewAuthenticationProvider(baseurl string, querier db.TransactionQuerier, authKey []byte, sessionKey []byte) (*authenticationProvider, error) {
	ab := authboss.New()

	ab.Config.Storage.Server = newAuthbossStorer(querier)
	ab.Config.Storage.SessionState = abclientstate.NewSessionStorer(sessionCookieName, authKey, sessionKey)
	ab.Config.Storage.CookieState = abclientstate.NewCookieStorer(authKey, sessionKey)

	ab.Config.Core.ViewRenderer = jsonRenderer{}
	ab.Config.Core.MailRenderer = defaults.JSONRenderer{}
	ab.Config.Core.Logger = &authbossLogger{}
	ab.Config.Core.Router = defaults.NewRouter()
	ab.Config.Core.ErrorHandler = &authbossErrorHandler{LogWriter: ab.Config.Core.Logger}
	ab.Config.Core.Responder = defaults.NewResponder(ab.Config.Core.ViewRenderer)

	redirector := defaults.NewRedirector(ab.Config.Core.ViewRenderer, authboss.FormValueRedirect)
	redirector.CorceRedirectTo200 = true
	ab.Config.Core.Redirector = redirector

	ab.Config.Core.Mailer = defaults.NewLogMailer(os.Stdout)

	ab.Config.Modules.TwoFactorEmailAuthRequired = false
	ab.Config.Modules.RoutesRedirectOnUnauthed = false

	ab.Config.Core.BodyReader = newAuthbossBodyReader()

	ab.Config.Paths.Mount = "/auth"
	ab.Config.Paths.RootURL = baseurl

	ab.Events.Before(authboss.EventRegister, func(w http.ResponseWriter, r *http.Request, handled bool) (bool, error) {
		return true, nil
	})

	tf := twofactor.Recovery{Authboss: ab}
	if err := tf.Setup(); err != nil {
		return nil, err
	}

	ab.Config.Modules.TOTP2FAIssuer = "licenta"

	totp := totp2fa.TOTP{Authboss: ab}
	if err := totp.Setup(); err != nil {
		return nil, err
	}

	webn, err := webauthn.New(&webauthn.Config{
		RPDisplayName:         "Licenta",
		RPID:                  "laptop.tedyst.ro",
		RPOrigins:             []string{"https://laptop.tedyst.ro"},
		AttestationPreference: protocol.PreferNoAttestation,
	})
	if err != nil {
		return nil, err
	}
	wa := authbosswebauthn.New(ab, webn, nil)

	if err := wa.Setup(); err != nil {
		return nil, err
	}

	if err := ab.Init(); err != nil {
		return nil, err
	}

	return &authenticationProvider{
		querier:  querier,
		authboss: ab,
	}, nil
}

func (auth *authenticationProvider) Middleware(next http.Handler) http.Handler {
	rememberMiddleware := remember.Middleware(auth.authboss)
	next = rememberMiddleware(next)
	loadClientStateMiddleware := auth.authboss.LoadClientStateMiddleware(next)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = context.WithValue(ctx, requestStorer{}, r)
		r = r.WithContext(ctx)

		loadClientStateMiddleware.ServeHTTP(w, r)
	})
}

func (auth *authenticationProvider) APIMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lockMiddleware := lock.Middleware(auth.authboss)(next)
		confirmMiddleware := confirm.Middleware(auth.authboss)(lockMiddleware)
		user, err := auth.authboss.CurrentUser(r)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		if user != nil {
			confirmMiddleware.ServeHTTP(w, r)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

func (auth *authenticationProvider) Handler() http.Handler {
	return auth.authboss.Config.Core.Router
}

func (auth *authenticationProvider) GetUser(ctx context.Context) (*models.User, error) {
	r := ctx.Value(requestStorer{}).(*http.Request)
	user, err := auth.authboss.CurrentUser(r)
	if err != nil {
		return nil, err
	}

	return user.(*authbossUser).user, nil
}

func (auth *authenticationProvider) UpdatePassword(ctx context.Context, user *models.User, newPassword string) error {
	return auth.authboss.UpdatePassword(ctx, &authbossUser{
		user: user,
	}, newPassword)
}

func (auth *authenticationProvider) VerifyPassword(ctx context.Context, user *models.User, password string) (bool, error) {
	err := auth.authboss.VerifyPassword(&authbossUser{
		user: user,
	}, password)
	if err != nil {
		return false, err
	}
	return true, nil
}
