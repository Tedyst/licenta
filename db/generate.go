//go:build generate
// +build generate

package db

import (
	_ "github.com/sqlc-dev/sqlc/cmd/sqlc"
	_ "go.uber.org/mock/mockgen"
)

//go:generate go run github.com/sqlc-dev/sqlc/cmd/sqlc generate
//go:generate go run go.uber.org/mock/mockgen -source=db.go -package mock -typed -destination mock/mock.go
