//go:build generate
// +build generate

package templates

import _ "github.com/valyala/quicktemplate/qtc"

//go:generate go run github.com/valyala/quicktemplate/qtc -dir=mail
