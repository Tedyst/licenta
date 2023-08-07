//go:build generate
// +build generate

package v1

import _ "github.com/deepmap/oapi-codegen/cmd/oapi-codegen"

//go:generate oapi-codegen -config oapi-codegen.yaml spec.yaml
