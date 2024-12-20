package endpoints

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"net/http"
)

// StartSession generates an access token for the user.
// @Summary StartSession generates an access token for the user.
// @Description Uses the refresh token to generate a new access-JWT for the user, if the refresh token is still valid.
// @Tags User Management
// @Accept json
// @Produce json
// @Param Refresh-Token body TokenHolder true "Refresh token of the user"
// @Success 200 {object} TokenHolder "Session start successful"
// @Failure 400 {object} ErrorResponse "Invalid request body"
// @Failure 401 {object} ErrorResponse "Refresh token expired"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /v1/user/start-session [post]
func StartSession(c *gin.Context) {
	ctx, span := tracer.Start(c.Request.Context(), "StartSession")
	defer span.End()

	// Bind JSON body to struct
	var body TokenHolder
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

	// ToDo: Implement the login process
	refreshJWT := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiI8dXNlci1pZD4iLCJuYW1lIjoiPHRva2VuLXR5cGU-IiwiaWF0IjoxNzM0Njk4NzEwfQ.hzvbcP77EO8dnEyy5i-OgoOp8MYYwslfwKx32ZKgrH8"

	c.JSON(
		http.StatusOK,
		TokenHolder{
			Token: refreshJWT,
		},
	)
}
