package v1

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

func InitV1Router(router fiber.Router) error {
	router.Get("/", func(c *fiber.Ctx) error {
		return c.Redirect(c.Path() + "/docs/")
	})

	router.Get("/docs/", swagger.HandlerDefault)

	router.Get("/docs/*", swagger.New(swagger.Config{ // custom
		URL:         "/api/v1/docs/doc.json",
		DeepLinking: false,
		// Expand ("list") or Collapse ("none") tag groups by default
		DocExpansion: "none",
		// Prefill OAuth ClientId on Authorize popup
		OAuth: &swagger.OAuthConfig{
			AppName:  "OAuth Provider",
			ClientId: "21bb4edc-05a7-4afc-86f1-2e151e4ba6e2",
		},
		// Ability to change OAuth2 redirect uri location
		OAuth2RedirectUrl: "http://localhost:8080/swagger/oauth2-redirect.html",
	}))

	router.Get("/users", HandleGetUsers)
	router.Get("/users/:id", HandleGetUser)
	router.Post("/users", HandleCreateUser)
	return nil
}
