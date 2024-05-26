package auth

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"errors"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/tedyst/authbosswebauthn"
	"github.com/tedyst/licenta/cache"
	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/db/queries"
	abclientstate "github.com/volatiletech/authboss-clientstate"
	"github.com/volatiletech/authboss/v3"
	_ "github.com/volatiletech/authboss/v3/auth"

	// "github.com/volatiletech/authboss/v3/confirm"
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
	cache    cache.CacheProvider[queries.User]
	authboss *authboss.Authboss
	querier  db.TransactionQuerier
}

func NewAuthenticationProvider(baseURL string, querier db.TransactionQuerier, authKey []byte, sessionKey []byte, emailTaskRunner emailTaskRunner, cache cache.CacheProvider[queries.User]) (*authenticationProvider, error) {
	ab := authboss.New()

	ab.Config.Storage.Server = newAuthbossStorer(querier, cache)
	ab.Config.Storage.SessionState = abclientstate.NewSessionStorer(sessionCookieName, authKey, sessionKey)
	ab.Config.Storage.CookieState = abclientstate.NewCookieStorer(authKey, sessionKey)

	ab.Config.Core.ViewRenderer = jsonRenderer{}
	ab.Config.Core.MailRenderer = defaults.JSONRenderer{}
	ab.Config.Core.Logger = &authbossLogger{}
	ab.Config.Core.Router = defaults.NewRouter()
	ab.Config.Core.ErrorHandler = &authbossErrorHandler{LogWriter: ab.Config.Core.Logger}
	ab.Config.Core.Responder = defaults.NewResponder(ab.Config.Core.ViewRenderer)
	ab.Config.Core.Mailer = &authbossMailer{
		runner: emailTaskRunner,
	}

	redirector := defaults.NewRedirector(ab.Config.Core.ViewRenderer, authboss.FormValueRedirect)
	redirector.CorceRedirectTo200 = true
	ab.Config.Core.Redirector = redirector

	ab.Config.Modules.LogoutMethod = "POST"

	ab.Config.Core.Mailer = defaults.NewLogMailer(os.Stdout)

	ab.Config.Modules.TwoFactorEmailAuthRequired = false
	ab.Config.Modules.RoutesRedirectOnUnauthed = false

	ab.Config.Core.BodyReader = newAuthbossBodyReader()

	ab.Config.Paths.Mount = "/auth"
	ab.Config.Paths.RootURL = baseURL

	webn, err := webauthn.New(&webauthn.Config{
		RPDisplayName:         "Licenta",
		RPID:                  baseURL,
		RPOrigins:             []string{"https://" + baseURL},
		AttestationPreference: protocol.PreferNoAttestation,
	})
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(baseURL, "localhost") {
		webn.Config.RPOrigins = []string{"http://" + baseURL, "https://" + baseURL}
	}

	wa := authbosswebauthn.New(ab, webn, nil)

	if err := wa.Setup(); err != nil {
		return nil, err
	}

	tf := twofactor.Recovery{Authboss: ab}
	if err := tf.Setup(); err != nil {
		return nil, err
	}

	ab.Config.Modules.TOTP2FAIssuer = baseURL

	totp := totp2fa.TOTP{Authboss: ab}
	if err := totp.Setup(); err != nil {
		return nil, err
	}

	if err := ab.Init(); err != nil {
		return nil, err
	}

	return &authenticationProvider{
		querier:  querier,
		authboss: ab,
		cache:    cache,
	}, nil
}

func (auth *authenticationProvider) Middleware(next http.Handler) http.Handler {
	rememberMiddleware := remember.Middleware(auth.authboss)
	next = rememberMiddleware(next)

	return auth.authboss.LoadClientStateMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		shouldRemember, ok := authboss.GetSession(r, "should_remember")
		r = r.WithContext(context.WithValue(r.Context(), authboss.CTXKeyValues, rememberer{remember: ok && shouldRemember == "true"}))
		next.ServeHTTP(w, r)
	}))
}

func (auth *authenticationProvider) APIMiddleware(next http.Handler) http.Handler {
	lockMiddleware := lock.Middleware(auth.authboss)(next)
	// confirmMiddleware := confirm.Middleware(auth.authboss)(lockMiddleware)
	confirmMiddleware := lockMiddleware
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := auth.authboss.LoadCurrentUser(&r)

		ctx := r.Context()
		ctx = context.WithValue(ctx, requestStorer{}, r)
		r = r.WithContext(ctx)

		if err != nil && err != authboss.ErrUserNotFound {
			slog.Error("Error while loading current user", "error", err.Error())
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"success": false, "message": "Internal server error"}`))
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

func (auth *authenticationProvider) GetUser(ctx context.Context) (*queries.User, error) {
	r, ok := ctx.Value(requestStorer{}).(*http.Request)
	if !ok {
		return nil, errors.New("request not found in context")
	}
	user, err := auth.authboss.CurrentUser(r)
	if err != nil && !errors.Is(err, authboss.ErrUserNotFound) {
		return nil, err
	}
	if errors.Is(err, authboss.ErrUserNotFound) {
		return nil, nil
	}

	u, ok := user.(*authbossUser)
	if !ok {
		return nil, errors.New("user not found")
	}
	return u.user, nil
}

func (auth *authenticationProvider) UpdatePassword(ctx context.Context, user *queries.User, newPassword string) error {
	return auth.authboss.UpdatePassword(ctx, &authbossUser{
		user: user,
	}, newPassword)
}

func (auth *authenticationProvider) VerifyPassword(ctx context.Context, user *queries.User, password string) (bool, error) {
	err := auth.authboss.VerifyPassword(&authbossUser{
		user: user,
	}, password)
	if err != nil {
		return false, err
	}
	return true, nil
}
