package v1

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/tedyst/licenta/config"
	"github.com/tedyst/licenta/db"
)

type publicUserAPIResponse struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type userCreateAPIRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

// @Summary		Get users
// @Description	Get all users
// @Tags			users
// @Accept			json
// @Produce		json
// @Success		200	{array}	publicUserAPIResponse
// @Router			/api/v1/users [get]
func HandleGetUsers(c *fiber.Ctx) error {
	ctx, span := config.Tracer.Start(c.UserContext(), "HandleGetUsers")
	defer span.End()

	users, err := config.DatabaseQueries.ListUsers(ctx)
	if err != nil {
		log.Println(err)
		return err
	}
	span.AddEvent("Creating response")
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

// @Summary		Create user
// @Description	Create a new user
// @Tags			users
// @Accept			json
// @Produce		json
// @Param 			user body userCreateAPIRequest true "User"
// @Success		200	{object}	publicUserAPIResponse
// @Router			/api/v1/users [post]
func HandleCreateUser(c *fiber.Ctx) error {
	request := userCreateAPIRequest{}
	err := c.BodyParser(&request)
	if err != nil {
		log.Println(err)
		return err
	}
	user, err := config.DatabaseQueries.CreateUser(c.Context(), db.CreateUserParams{
		Username: request.Username,
		Email:    request.Email,
	})
	if err != nil {
		log.Println(err)
		return err
	}

	err = user.SetPassword(request.Password)
	if err != nil {
		log.Println(err)
		return err
	}

	sess, err := config.SessionStore.Get(c)
	if err != nil {
		log.Println(err)
		return err
	}
	sess.Set("user_id", 1)
	log.Print(sess)
	err = sess.Save()
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

// @Summary		Get user
// @Description	Get a user
// @Tags			users
// @Accept			json
// @Produce		json
// @Param 			user body userCreateAPIRequest true "User"
// @Success		200	{object}	publicUserAPIResponse
// @Router			/api/v1/users/{id} [get]
func HandleGetUser(c *fiber.Ctx) error {
	user, err := config.DatabaseQueries.GetUser(c.Context(), 1)
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
