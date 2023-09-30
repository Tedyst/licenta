package session

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/tedyst/licenta/api/v1/generated"
	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/models"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("github.com/tedyst/licenta/api/v1/middleware/session")

const cookieSessionKey = "session"

type sessionData struct {
	Session        *models.Session
	User           *models.User
	Waiting2faUser *models.User
	sessionChanged bool
}
type contextSessionKey struct{}

type config struct {
	debug bool
}

type sessionStore struct {
	database db.TransactionQuerier
	config   config
}

func (store *sessionStore) initSessionData(ctx context.Context, w http.ResponseWriter, r *http.Request) (*http.Request, *sessionData) {
	data := ctx.Value(contextSessionKey{})
	if data != nil {
		return r, data.(*sessionData)
	}
	newData := sessionData{}
	newCtx := context.WithValue(ctx, contextSessionKey{}, &newData)
	return r.WithContext(newCtx), &newData
}

func (store *sessionStore) createNewSession(ctx context.Context, data *sessionData) error {
	sess, err := store.database.CreateSession(ctx, queries.CreateSessionParams{
		ID: uuid.New(),
	})
	if err != nil {
		return errors.Wrap(err, "createNewSession: error creating session")
	}
	data.Session = sess
	return nil
}

func (store *sessionStore) getSession(ctx context.Context, data *sessionData, r *http.Request) error {
	cookie, err := r.Cookie(cookieSessionKey)
	if err != nil {
		return errors.Wrap(err, "getSession: error getting cookie")
	}
	u, err := uuid.Parse(cookie.Value)
	if err != nil {
		return errors.Wrap(err, "getSession: error parsing cookie")
	}
	sess, err := store.database.GetSession(ctx, u)
	if err != nil {
		return errors.Wrap(err, "getSession: error getting session")
	}
	data.Session = sess
	return nil
}

func (store *sessionStore) saveSession(ctx context.Context, data *sessionData) error {
	ctx, span := tracer.Start(ctx, "SaveSession")
	defer span.End()

	err := store.database.UpdateSession(ctx, queries.UpdateSessionParams{
		ID:      data.Session.ID,
		UserID:  data.Session.UserID,
		TotpKey: data.Session.TotpKey,
	})
	if err != nil {
		return errors.Wrap(err, "SaveSession: error updating session")
	}
	return nil
}

func (store *sessionStore) getUser(ctx context.Context, data *sessionData) error {
	if !data.Session.UserID.Valid {
		return nil
	}
	user, err := store.database.GetUser(ctx, data.Session.UserID.Int64)
	if err != nil {
		return errors.Wrap(err, "getUser: error getting user")
	}
	data.User = user
	return nil
}

func (store *sessionStore) initializeSession(ctx context.Context, r *http.Request, w http.ResponseWriter) (*http.Request, *sessionData, error) {
	ctx, span := tracer.Start(ctx, "InitializeSession")
	defer span.End()

	r, data := store.initSessionData(ctx, w, r)
	if data.Session == nil {
		err := store.createNewSession(ctx, data)
		if err != nil {
			return r, data, errors.Wrap(err, "initializeSession: error creating new session")
		}
	} else {
		err := store.getSession(ctx, data, r)
		if err != nil {
			return r, data, errors.Wrap(err, "initializeSession: error getting session")
		}
	}

	err := store.getUser(ctx, data)
	if err != nil {
		return r, data, errors.Wrap(err, "initializeSession: error getting user")
	}

	return r, data, nil
}

func (store *sessionStore) GetUser(ctx context.Context) *models.User {
	ctx, span := tracer.Start(ctx, "GetUser")
	defer span.End()

	return ctx.Value(contextSessionKey{}).(*sessionData).User
}

func (store *sessionStore) SetUser(ctx context.Context, user *models.User) {
	data := ctx.Value(contextSessionKey{}).(*sessionData)
	data.User = user

	data.Session.UserID = sql.NullInt64{
		Int64: user.ID,
		Valid: true,
	}
	data.sessionChanged = true
}

func (store *sessionStore) SetWaiting2FA(ctx context.Context, waitingUser *models.User) {
	data := ctx.Value(contextSessionKey{}).(*sessionData)
	data.Session.Waiting2fa = sql.NullInt64{
		Int64: waitingUser.ID,
		Valid: true,
	}
	data.sessionChanged = true
}

func (store *sessionStore) GetWaiting2FA(ctx context.Context) *models.User {
	data := ctx.Value(contextSessionKey{}).(*sessionData)
	return data.Waiting2faUser
}

func (store *sessionStore) SetTOTPKey(ctx context.Context, key string) {
	data := ctx.Value(contextSessionKey{}).(*sessionData)
	data.Session.TotpKey = sql.NullString{
		String: key,
		Valid:  true,
	}
	data.sessionChanged = true
}

func (store *sessionStore) GetTOTPKey(ctx context.Context) (string, error) {
	data := ctx.Value(contextSessionKey{}).(*sessionData)
	if !data.Session.TotpKey.Valid {
		return "", errors.New("GetTOTPKey: no totp key")
	}
	return data.Session.TotpKey.String, nil
}

func (store *sessionStore) ClearSession(ctx context.Context) {
	data := ctx.Value(contextSessionKey{}).(*sessionData)
	data.Session.UserID = sql.NullInt64{
		Int64: 0,
		Valid: false,
	}
	data.Session.Waiting2fa = sql.NullInt64{
		Int64: 0,
		Valid: false,
	}
	data.Session.TotpKey = sql.NullString{
		String: "",
		Valid:  false,
	}
	data.sessionChanged = true
}

func New(database db.TransactionQuerier, debug bool) *sessionStore {
	return &sessionStore{
		database: database,
		config: config{
			debug: debug,
		},
	}
}

func (store *sessionStore) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r, data, err := store.initializeSession(r.Context(), r, w)
		if err != nil {
			tracer := trace.SpanFromContext(r.Context())
			tracer.RecordError(err)
			tracer.SetStatus(codes.Error, err.Error())

			var message = "Internal server error"
			if store.config.debug {
				message = err.Error()
			}

			data, err := json.Marshal(generated.Error{
				Success: false,
				Message: message,
			})
			if err != nil {
				panic(err)
			}
			w.Write(data)
			return
		}
		next.ServeHTTP(w, r)
		http.SetCookie(w, &http.Cookie{
			Name:     cookieSessionKey,
			Value:    data.Session.ID.String(),
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		})
	})
}
