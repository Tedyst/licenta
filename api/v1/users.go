package v1

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/tedyst/licenta/config"
	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/middleware/session"
	"go.opentelemetry.io/otel/codes"
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
// @Param 			offset query int false "Page" default(0)
// @Param 			limit query int false "Limit" default(10)
// @Success		200	{object}	PaginationResponse[[]publicUserAPIResponse]
// @Router			/api/v1/users [get]
func HandleGetUsers(c *fiber.Ctx) error {
	ctx, span := config.Tracer.Start(c.UserContext(), "HandleGetUsers")
	defer span.End()

	_, _, err := verifyIfAdmin(ctx, c)
	if err != nil {
		return err
	}

	offset, limit, err := GetOffsetAndLimit(c)
	if err != nil {
		return handleError(c, span, err)
	}

	users, err := config.DatabaseQueries.ListUsersPaginated(ctx, db.ListUsersPaginatedParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return handleError(c, span, err)
	}
	count, err := config.DatabaseQueries.CountUsers(ctx)
	if err != nil {
		return handleError(c, span, err)
	}
	var publicUsers []publicUserAPIResponse
	for _, user := range users {
		publicUsers = append(publicUsers, publicUserAPIResponse{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		})
	}
	return c.JSON(NewPaginationResponse(publicUsers, int32(count), offset, limit))
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
	ctx, span := config.Tracer.Start(c.UserContext(), "HandleCreateUser")
	defer span.End()

	_, _, err := verifyIfAdmin(ctx, c)
	if err != nil {
		return err
	}

	request := userCreateAPIRequest{}
	err = c.BodyParser(&request)
	if err != nil {
		span.SetStatus(codes.Error, "Error parsing body")
		span.RecordError(err)
		return err
	}
	user, err := config.DatabaseQueries.CreateUser(ctx, db.CreateUserParams{
		Username: request.Username,
		Email:    request.Email,
	})
	if err != nil {
		span.SetStatus(codes.Error, "Error creating user")
		span.RecordError(err)
		return err
	}

	err = user.SetPassword(request.Password)
	if err != nil {
		span.SetStatus(codes.Error, "Error setting password")
		span.RecordError(err)
		return err
	}

	err = config.DatabaseQueries.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		ID:       user.ID,
		Password: user.Password,
	})
	if err != nil {
		span.SetStatus(codes.Error, "Error updating password")
		span.RecordError(err)
		return err
	}

	sess, err := session.GetSession(ctx, c)
	if err != nil {
		span.SetStatus(codes.Error, "Error getting session")
		span.RecordError(err)
		return err
	}
	sess.UserID = pgtype.Int8{Int64: user.ID, Valid: true}
	err = session.SaveSession(ctx, c, sess)
	if err != nil {
		span.SetStatus(codes.Error, "Error saving session")
		span.RecordError(err)
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
	ctx, span := config.Tracer.Start(c.UserContext(), "HandleGetUser")
	defer span.End()

	user, err := config.DatabaseQueries.GetUser(ctx, 1)
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
