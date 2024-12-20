package endpoints

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"net/http"
)

type UserBody struct {
	Email    string `json:"email" binding:"required,email" example:"bob.alice@example.com"`
	Password string `json:"password" binding:"required" example:"<password>"`
}

// Register handles the user registration process.
// @Summary Register a new user
// @Description Creates a new user account in the database with the provided details and starts the verification process.
// @Tags User Management
// @Accept json
// @Produce json
// @Param User body UserBody true "Email address and password of the user"
// @Success 200 {object} SuccessResponse "Registration successful"
// @Failure 400 {object} ErrorResponse "Invalid request body"
// @Failure 409 {object} ErrorResponse "User already exists"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /v1/user/register [post]
func Register(c *gin.Context) {
	ctx, span := tracer.Start(c.Request.Context(), "RegisterUser")
	defer span.End()

	// Bind JSON body to struct
	var body UserBody
	if err := c.ShouldBindJSON(&body); err != nil {
		err = errors.Wrap(err, "Failed to bind JSON body")
		logger.Debug(ctx, err)
		c.JSON(
			http.StatusBadRequest,
			ErrorResponse{
				Error: "Invalid request body",
			},
		)
		return
	}

	// ToDo: Implement the registration process

	c.JSON(
		http.StatusOK,
		SuccessResponse{
			Message: "Registration successful",
		},
	)
}
