//go:build generate
// +build generate

package v1

import _ "github.com/deepmap/oapi-codegen/cmd/oapi-codegen"

//go:generate oapi-codegen -package generated -generate types,spec,fiber,strict-server -o generated/gen.go spec.yaml
