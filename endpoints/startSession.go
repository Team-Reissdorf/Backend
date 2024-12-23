package endpoints

import (
	"github.com/Team-Reissdorf/Backend/endpoints/authMiddleware"
	"github.com/Team-Reissdorf/Backend/endpoints/standardJsonAnswers"
	"github.com/gin-gonic/gin"
	"net/http"
)

// StartSession generates an access token for the user.
// @Summary StartSession generates an access token for the user.
// @Description Uses the refresh token to generate a new access-JWT for the user, if the refresh token is still valid.
// @Tags User Management
// @Accept json
// @Produce json
// @Param Authorization  header  string  true "Refresh JWT"
// @Success 200 {object} TokenHolder "Session start successful"
// @Failure 401 {object} standardJsonAnswers.ErrorResponse "The token is invalid"
// @Failure 500 {object} standardJsonAnswers.ErrorResponse "Internal server error"
// @Router /v1/user/start-session [post]
func StartSession(c *gin.Context) {
	ctx, span := tracer.Start(c.Request.Context(), "StartSession")
	defer span.End()

	// Get the user id from the context
	userId, exists := c.Get(authMiddleware.UserIdContextKey)
	if !exists || userId == nil {
		logger.Debug(ctx, "User ID not found in the context")
		c.JSON(
			http.StatusInternalServerError,
			standardJsonAnswers.ErrorResponse{
				Error: "User ID not found in the context",
			},
		)
		c.Abort()
		return
	}

	// ToDo: Validate the refresh token

	// ToDo: Generate the access token
	accessJWT := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiI8dXNlci1pZD4iLCJuYW1lIjoiPHRva2VuLXR5cGU-IiwiaWF0IjoxNzM0Njk4NzEwfQ.hzvbcP77EO8dnEyy5i-OgoOp8MYYwslfwKx32ZKgrH8"

	c.JSON(
		http.StatusOK,
		TokenHolder{
			Token: accessJWT,
		},
	)
}
