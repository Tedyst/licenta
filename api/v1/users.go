package v1

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/tedyst/licenta/db"
)

type publicUserAPIResponse struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

func HandleGetUsers(c *fiber.Ctx) error {
	users, err := c.Locals("queries").(*db.Queries).ListUsers(c.Context())
	if err != nil {
		log.Println(err)
		return err
	}
	var publicUsers []publicUserAPIResponse
	for _, user := range users {
		publicUsers = append(publicUsers, publicUserAPIResponse{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		})
	}
	return c.JSON(publicUsers)
}

func HandleCreateUser(c *fiber.Ctx) error {
	user, err := c.Locals("queries").(*db.Queries).CreateUser(c.Context(), db.CreateUserParams{
		Username: "tedyst1",
		Password: "1234561",
		Email:    "stoicatedy1@gmail.com",
	})
	if err != nil {
		log.Println(err)
		return err
	}
	return c.JSON(publicUserAPIResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
	})
}
