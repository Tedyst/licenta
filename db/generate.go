//go:build generate
// +build generate

package db

import (
	_ "github.com/sqlc-dev/sqlc/cmd/sqlc"
	_ "go.uber.org/mock/mockgen"
)

//go:generate sqlc generate
//go:generate mockgen -source=db.go -package mock -typed -destination mock/mock.go
