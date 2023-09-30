package handlers_test

import (
	"context"
	"testing"

	"github.com/tedyst/licenta/api/v1/generated"
	"github.com/tedyst/licenta/api/v1/handlers"
	sessionmock "github.com/tedyst/licenta/api/v1/middleware/session/mock"
	dbmock "github.com/tedyst/licenta/db/mock"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/models"
	"go.uber.org/mock/gomock"
)

func TestLogin200(t *testing.T) {
	var tests = []struct {
		user       queries.User
		username   string
		password   string
		totp       *string
		statusCode int
	}{
		{
			user: queries.User{
				ID:       1,
				Username: "test",
				Email:    "asd@asd.com",
				Password: "asd123123",
			},
			username:   "test",
			password:   "asd123123",
			totp:       nil,
			statusCode: 200,
		},
	}
	t.Parallel()
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			querier := dbmock.NewMockTransactionQuerier(ctrl)
			sessionStore := sessionmock.NewMockSessionStore(ctrl)
			server := handlers.NewServerHandler(querier, sessionStore)

			hash, err := models.GenerateHash(context.Background(), test.user.Password)
			if err != nil {
				t.Fatal(err)
			}
			newUser := test.user
			newUser.Password = hash

			querier.EXPECT().GetUserByUsernameOrEmail(gomock.Any(), test.username).Return(&newUser, nil)

			sessionStore.EXPECT().SetUser(gomock.Any(), gomock.Any()).Do(func(ctx context.Context, user *models.User) {
				if user != &newUser {
					t.Errorf("expected user %v, got %v", newUser, user)
				}
			})

			resp, err := server.PostLogin(context.Background(), generated.PostLoginRequestObject{
				Body: &generated.LoginUser{
					Username: test.username,
					Password: test.password,
					TotpCode: test.totp,
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
