//go:build generate
// +build generate

package database

import (
	_ "github.com/sqlc-dev/sqlc/cmd/sqlc"
)

//go:generate sqlc generate
