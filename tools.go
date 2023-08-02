//go:build tools
// +build tools

package main

import (
	_ "github.com/sqlc-dev/sqlc/cmd/sqlc"
	_ "github.com/swaggo/swag/cmd/swag"
	_ "github.com/valyala/quicktemplate/qtc"
)
