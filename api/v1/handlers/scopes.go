package handlers

import (
	"context"
)

type Scope string

var (
	UserScope   Scope = "user"
	WorkerScope Scope = "worker"
)

func (server *serverHandler) IsScopeAllowed(ctx context.Context, s string, scope Scope) bool {
	sc := server.SessionStore.GetScope(ctx)
	if sc == nil {
		return false
	}
	return s == string(scope)
}
