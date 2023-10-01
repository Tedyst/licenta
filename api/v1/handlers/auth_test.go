package handlers_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
	"github.com/tedyst/licenta/api/v1/generated"
	"github.com/tedyst/licenta/api/v1/handlers"
	sessionmock "github.com/tedyst/licenta/api/v1/middleware/session/mock"
	dbmock "github.com/tedyst/licenta/db/mock"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/models"
	"go.uber.org/mock/gomock"
)

func generateTotpSecret(t *testing.T) string {
	secret, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "test",
		AccountName: "test",
	})
	if err != nil {
		t.Fatal(err)
	}
	return secret.Secret()
}

func generateTotpCode(t *testing.T, secret string, tim time.Time) *string {
	code, err := totp.GenerateCode(secret, tim)
	if err != nil {
		t.Fail()
	}
	return &code
}

func TestPostLogin(t *testing.T) {
	var tests = []struct {
		user                *queries.User
		username            string
		password            string
		statusCode          int
		shouldGetUser       bool
		shouldSetSession    bool
		shouldSetWaiting2FA bool
	}{
		{
			user: &queries.User{
				ID:       1,
				Username: "test",
				Email:    "asd@asd.com",
				Password: "asd123123",
			},
			username:         "test",
			password:         "asd123123",
			statusCode:       200,
			shouldGetUser:    true,
			shouldSetSession: true,
		},
		{
			user: &queries.User{
				ID:       1,
				Username: "test",
				Email:    "asd@asd.com",
				Password: "asd123123",
			},
			username:         "asd@asd.com",
			password:         "asd123123",
			statusCode:       200,
			shouldGetUser:    true,
			shouldSetSession: true,
		},
		{
			user:             nil,
			username:         "asd@asd.com",
			password:         "asd123123",
			statusCode:       401,
			shouldGetUser:    true,
			shouldSetSession: false,
		},
		{
			user: &queries.User{
				ID:         1,
				Username:   "test",
				Email:      "asd@asd.com",
				Password:   "asd123123",
				TotpSecret: sql.NullString{String: "asd", Valid: true},
			},
			username:            "test",
			password:            "asd123123",
			statusCode:          401,
			shouldGetUser:       true,
			shouldSetSession:    false,
			shouldSetWaiting2FA: true,
		},
		{
			user: &queries.User{
				ID:       1,
				Username: "test",
				Email:    "asd@asd.com",
				Password: "asd123123",
			},
			username:         "test",
			password:         "asd123124",
			statusCode:       401,
			shouldSetSession: false,
			shouldGetUser:    true,
		},
		{
			user: &queries.User{
				ID:       1,
				Username: "test",
				Email:    "asd@asd.com",
				Password: "asd",
			},
			username:         "test",
			password:         "asd",
			statusCode:       400,
			shouldGetUser:    false,
			shouldSetSession: false,
		},
	}
	for _, test := range tests {
		test := test
		t.Run("", func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			querier := dbmock.NewMockTransactionQuerier(ctrl)
			sessionStore := sessionmock.NewMockSessionStore(ctrl)
			server := handlers.NewServerHandler(querier, sessionStore)

			newUser := &queries.User{}
			if test.user != nil {
				*newUser = *test.user

				hash, err := models.GenerateHash(context.Background(), test.user.Password)
				if err != nil {
					t.Fatal(err)
				}

				newUser.Password = hash
			}

			if test.shouldGetUser {
				if test.user != nil {
					querier.EXPECT().GetUserByUsernameOrEmail(gomock.Any(), test.username).Return(newUser, nil)
				} else {
					querier.EXPECT().GetUserByUsernameOrEmail(gomock.Any(), test.username).Return(nil, errors.New("user not found"))
				}
			}

			if test.shouldSetSession {
				sessionStore.EXPECT().SetUser(gomock.Any(), gomock.Any()).Do(func(ctx context.Context, user *models.User) {
					if user != newUser {
						t.Errorf("expected user %v, got %v", newUser, user)
					}
				})
			}

			if test.shouldSetWaiting2FA {
				sessionStore.EXPECT().SetWaiting2FA(gomock.Any(), gomock.Any()).Do(func(ctx context.Context, user *models.User) {
					if user != newUser {
						t.Errorf("expected user %v, got %v", newUser, user)
					}
				})
			}

			resp, err := server.PostLogin(context.Background(), generated.PostLoginRequestObject{
				Body: &generated.LoginUser{
					Username: test.username,
					Password: test.password,
				},
			})
			if err != nil {
				t.Fatal(err)
			}

			switch test.statusCode {
			case 200:
				_, ok := resp.(generated.PostLogin200JSONResponse)
				if !ok {
					t.Fatalf("expected 200 response, got %T", resp)
				}
				if !resp.(generated.PostLogin200JSONResponse).Success {
					t.Errorf("expected success, got %v", resp.(generated.PostLogin200JSONResponse).Success)
				}
			case 400:
				_, ok := resp.(generated.PostLogin400JSONResponse)
				if !ok {
					t.Fatalf("expected 400 response, got %T", resp)
				}
				if resp.(generated.PostLogin400JSONResponse).Success {
					t.Errorf("expected failure, got %v", resp.(generated.PostLogin400JSONResponse).Success)
				}
			case 401:
				_, ok := resp.(generated.PostLogin401JSONResponse)
				if !ok {
					t.Fatalf("expected 401 response, got %T", resp)
				}
				if resp.(generated.PostLogin401JSONResponse).Success {
					t.Errorf("expected failure, got %v", resp.(generated.PostLogin401JSONResponse).Success)
				}
			default:
				t.Fatalf("unknown status code %d", test.statusCode)
			}
		})
	}
}

func TestPostLogout(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	querier := dbmock.NewMockTransactionQuerier(ctrl)
	sessionStore := sessionmock.NewMockSessionStore(ctrl)
	server := handlers.NewServerHandler(querier, sessionStore)

	sessionStore.EXPECT().ClearSession(gomock.Any())

	resp, err := server.PostLogout(context.Background(), generated.PostLogoutRequestObject{})
	if err != nil {
		t.Fatal(err)
	}

	_, ok := resp.(generated.PostLogout200JSONResponse)
	if !ok {
		t.Fatalf("expected 200 response, got %T", resp)
	}

	if !resp.(generated.PostLogout200JSONResponse).Success {
		t.Errorf("expected success, got %v", resp.(generated.PostLogout200JSONResponse).Success)
	}
}
