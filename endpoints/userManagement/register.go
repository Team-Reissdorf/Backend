package userManagement

import (
	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/Team-Reissdorf/Backend/hashingHelper"
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
// @Success 200 {object} endpoints.SuccessResponse "Registration successful"
// @Failure 400 {object} endpoints.ErrorResponse "Invalid request body"
// @Failure 409 {object} endpoints.ErrorResponse "User already exists"
// @Failure 500 {object} endpoints.ErrorResponse "Internal server error"
// @Router /v1/user/register [post]
func Register(c *gin.Context) {
	ctx, span := endpoints.Tracer.Start(c.Request.Context(), "RegisterUser")
	defer span.End()

	// Bind JSON body to struct
	var body UserBody
	if err := c.ShouldBindJSON(&body); err != nil {
		err = errors.Wrap(err, "Failed to bind JSON body")
		endpoints.Logger.Debug(ctx, err)
		c.JSON(
			http.StatusBadRequest,
			endpoints.ErrorResponse{
				Error: "Invalid request body",
			},
		)
		return
	}

	hash, err1 := hashingHelper.DefaultHashParams.HashPassword(ctx, body.Password)
	if err1 != nil {
		err1 = errors.Wrap(err1, "Failed to hash password")
		endpoints.Logger.Error(ctx, err1)
		c.JSON(
			http.StatusInternalServerError,
			endpoints.ErrorResponse{
				Error: "Internal server error",
			},
		)
		return
	}
	endpoints.Logger.Debug(ctx, "Hashed password: ", hash) // ToDo: Remove this line

	// ToDo: Implement the registration process

	c.JSON(
		http.StatusOK,
		endpoints.SuccessResponse{
			Message: "Registration successful",
		},
	)
}
