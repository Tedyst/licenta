package handlers

import (
	"context"

	scopes "github.com/SonicRoshan/scope"
)

var (
	UsersReadScope    = "users.read"
	UsersWriteScope   = "users.write"
	UsersMeReadScope  = "users.me.read"
	UsersMeWriteScope = "users.me.write"
)

func (server *serverHandler) IsScopeAllowed(ctx context.Context, s string) bool {
	sc := server.SessionStore.GetScope(ctx)
	return scopes.ScopeInAllowed(s, sc) || scopes.ScopeInAllowed("user", sc)
}
