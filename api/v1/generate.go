//go:build generate
// +build generate

package v1

import _ "github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen"

//go:generate go run github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen -config oapi-codegen.yaml spec.yaml
