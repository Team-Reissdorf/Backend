package endpoints

import (
	"github.com/Team-Reissdorf/Backend/endpoints/standardJsonAnswers"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"net/http"
)

type TokenHolder struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiI8dXNlci1pZD4iLCJuYW1lIjoiPHRva2VuLXR5cGU-IiwiaWF0IjoxNzM0Njk4NzEwfQ.hzvbcP77EO8dnEyy5i-OgoOp8MYYwslfwKx32ZKgrH8"`
}

// Login handles the user login process.
// @Summary Login to an existing user account
// @Description Logs the user into the system with the provided credentials and returns a refresh-JWT.
// @Tags User Management
// @Accept json
// @Produce json
// @Param User body UserBody true "Email address and password of the user"
// @Success 200 {object} TokenHolder "Login successful"
// @Failure 400 {object} standardJsonAnswers.ErrorResponse "Invalid request body"
// @Failure 404 {object} standardJsonAnswers.ErrorResponse "User not found"
// @Failure 500 {object} standardJsonAnswers.ErrorResponse "Internal server error"
// @Router /v1/user/login [post]
func Login(c *gin.Context) {
	ctx, span := tracer.Start(c.Request.Context(), "LoginUser")
	defer span.End()

	// Bind JSON body to struct
	var body UserBody
	if err := c.ShouldBindJSON(&body); err != nil {
		err = errors.Wrap(err, "Failed to bind JSON body")
		logger.Debug(ctx, err)
		c.JSON(
			http.StatusBadRequest,
			standardJsonAnswers.ErrorResponse{
				Error: "Invalid request body",
			},
		)
		return
	}

	// ToDo: Implement the login process
	refreshJWT := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiI8dXNlci1pZD4iLCJuYW1lIjoiPHRva2VuLXR5cGU-IiwiaWF0IjoxNzM0Njk4NzEwfQ.hzvbcP77EO8dnEyy5i-OgoOp8MYYwslfwKx32ZKgrH8"

	c.JSON(
		http.StatusOK,
		TokenHolder{
			Token: refreshJWT,
		},
	)
}
