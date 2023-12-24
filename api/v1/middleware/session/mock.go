//go:build generate
// +build generate

package session

import _ "go.uber.org/mock/mockgen"

//go:generate go run go.uber.org/mock/mockgen -source=session.go -package mock -typed -destination mock/mock.go
