package v1

import (
	"github.com/tedyst/licenta/api/v1/generated"
	"github.com/tedyst/licenta/api/v1/handlers"
)

func GetServerHandler() generated.ServerInterface {
	return &handlers.ServerHandler{}
}
