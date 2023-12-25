package auth

import (
	"context"
	"net/http"

	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/models"
	abclientstate "github.com/volatiletech/authboss-clientstate"
	"github.com/volatiletech/authboss/v3"
	_ "github.com/volatiletech/authboss/v3/auth"
	"github.com/volatiletech/authboss/v3/defaults"
	_ "github.com/volatiletech/authboss/v3/register"
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

	ab.Config.Core.ViewRenderer = defaults.JSONRenderer{}

	ab.Config.Paths.Mount = "/auth"
	ab.Config.Paths.RootURL = baseurl

	defaults.SetCore(&ab.Config, true, true)

	if err := ab.Init(); err != nil {
		return nil, err
	}

	return &authenticationProvider{
		querier:  querier,
		authboss: ab,
	}, nil
}

func (auth *authenticationProvider) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = context.WithValue(ctx, requestStorer{}, r)
		r = r.WithContext(ctx)

		auth.authboss.LoadClientStateMiddleware(next).ServeHTTP(w, r)
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
