//go:build generate
// +build generate

package db_generate

import (
	_ "github.com/sqlc-dev/sqlc/cmd/sqlc"
)

//go:generate sqlc generate
