package session

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
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

type SessionStore interface {
	GetUser(ctx context.Context) *models.User
	SetUser(ctx context.Context, user *models.User)
	ClearSession(ctx context.Context)
	Handler(next http.Handler) http.Handler
}

const cookieSessionKey = "session"

type sessionData struct {
	Session        *models.Session
	User           *models.User
	Waiting2faUser *models.User
	sessionChanged bool

	request *http.Request

	initialized bool
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
	newData := sessionData{
		initialized: false,
		request:     r,
	}
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
		ID:     data.Session.ID,
		UserID: data.Session.UserID,
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

func (store *sessionStore) initializeSession(ctx context.Context) (*sessionData, error) {
	sessionData, err := store.getSessionData(ctx)
	if err != nil {
		return sessionData, errors.Wrap(err, "initializeSession: error getting session data")
	}
	if sessionData != nil && sessionData.initialized {
		return sessionData, nil
	}

	ctx, span := tracer.Start(ctx, "initializeSession")
	defer span.End()

	if sessionData.Session == nil {
		err := store.createNewSession(ctx, sessionData)
		if err != nil {
			return sessionData, errors.Wrap(err, "initializeSession: error creating new session")
		}
	} else {
		err := store.getSession(ctx, sessionData, sessionData.request)
		if err != nil {
			return sessionData, errors.Wrap(err, "initializeSession: error getting session")
		}
	}

	err = store.getUser(ctx, sessionData)
	if err != nil {
		return sessionData, errors.Wrap(err, "initializeSession: error getting user")
	}

	sessionData.initialized = true

	return sessionData, nil
}

func (store *sessionStore) getSessionData(ctx context.Context) (*sessionData, error) {
	data := ctx.Value(contextSessionKey{})
	if data == nil {
		return nil, errors.New("getSessionData: no session data")
	}
	switch newData := data.(type) {
	case *sessionData:
		return newData, nil
	default:
		return nil, errors.New("getSessionData: invalid session data")
	}
}

func (store *sessionStore) GetUser(ctx context.Context) *models.User {
	ctx, span := tracer.Start(ctx, "GetUser")
	defer span.End()

	data, err := store.initializeSession(ctx)
	if err != nil {
		return nil
	}
	return data.User
}

func (store *sessionStore) SetUser(ctx context.Context, user *models.User) {
	data, err := store.initializeSession(ctx)
	if err != nil {
		return
	}
	data.sessionChanged = true
	data.User = user

	if user == nil {
		data.Session.UserID = sql.NullInt64{
			Int64: 0,
			Valid: false,
		}
		return
	}

	data.Session.UserID = sql.NullInt64{
		Int64: user.ID,
		Valid: true,
	}
}

func (store *sessionStore) ClearSession(ctx context.Context) {
	data, err := store.initializeSession(ctx)
	if err != nil {
		return
	}
	data.Session.UserID = sql.NullInt64{
		Int64: 0,
		Valid: false,
	}
	data.User = nil
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

func (store *sessionStore) showError(w http.ResponseWriter, r *http.Request, err error) {
	tracer := trace.SpanFromContext(r.Context())
	tracer.RecordError(err)
	tracer.SetStatus(codes.Error, err.Error())
	w.WriteHeader(http.StatusInternalServerError)
	if store.config.debug {
		data, err := json.Marshal(generated.Error{
			Success: false,
			Message: err.Error(),
		})
		if err != nil {
			_, err := w.Write([]byte(`{"success":false,"message":"Internal server error"}`))
			if err != nil {
				slog.Error("Error writing response", "error", err)
			}
			return
		}
		_, err = w.Write([]byte(data))
		if err != nil {
			slog.Error("Error writing response", "error", err)
		}
	} else {
		_, err := w.Write([]byte(`{"success":false,"message":"Internal server error"}`))
		if err != nil {
			slog.Error("Error writing response", "error", err)
		}
	}
}

func (store *sessionStore) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r, data := store.initSessionData(r.Context(), w, r)
		next.ServeHTTP(w, r)
		if data.sessionChanged {
			err := store.saveSession(r.Context(), data)
			if err != nil {
				store.showError(w, r, err)
				return
			}
		}
		if data.initialized {
			http.SetCookie(w, &http.Cookie{
				Name:     cookieSessionKey,
				Value:    data.Session.ID.String(),
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteStrictMode,
			})
		}
	})
}
