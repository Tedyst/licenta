package models_test

import (
	"context"
	"testing"

	db "github.com/tedyst/licenta/db/generated"
	"github.com/tedyst/licenta/models"
)

func TestUserVerifyPassword(t *testing.T) {
	user := db.User{}
	err := models.SetPassword(context.Background(), &user, "test")
	if err != nil {
		t.Error(err)
	}
	ok, err := models.VerifyPassword(context.Background(), &user, "test")
	if err != nil {
		t.Error(err)
	}
	if !ok {
		t.Error("Password verification failed")
	}
}

func TestUserWrongPassword(t *testing.T) {
	user := db.User{}
	err := models.SetPassword(context.Background(), &user, "test")
	if err != nil {
		t.Error(err)
	}
	ok, err := models.VerifyPassword(context.Background(), &user, "test2")
	if err != nil {
		t.Error(err)
	}
	if ok {
		t.Error("Password verification should have failed")
	}
}

func TestUserVerifyPasswordFromDB(t *testing.T) {
	user := db.User{
		Password: "$argon2id$v=19$m=65536,t=3,p=2$GenWczla9FZ9Ub77I1zYXQ$RgiRBtL8oJp7X/gReYHhJcZfvXYKvrv0uV4ZiTVJqo8",
	}
	ok, err := models.VerifyPassword(context.Background(), &user, "test")
	if err != nil {
		t.Error(err)
	}
	if !ok {
		t.Error("Password verification failed")
	}
}
