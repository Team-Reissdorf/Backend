package userManagement

import (
	"github.com/Team-Reissdorf/Backend/authHelper"
	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"net/http"
)

type AccessTokenHolder struct {
	AccessToken string `json:"access-token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiI8dXNlci1pZD4iLCJuYW1lIjoiPHRva2VuLXR5cGU-IiwiaWF0IjoxNzM0Njk4NzEwfQ.hzvbcP77EO8dnEyy5i-OgoOp8MYYwslfwKx32ZKgrH8"`
}

// StartSession generates an access token for the user.
// @Summary StartSession generates an access token for the user.
// @Description Uses the refresh token to generate a new access-JWT for the user, if the refresh token is still valid.
// @Tags User Management
// @Accept json
// @Produce json
// @Param Authorization  header  string  false  "Refresh JWT is sent in the Authorization header or set as a http-only cookie"
// @Success 200 {object} AccessTokenHolder "Session start successful"
// @Failure 401 {object} endpoints.ErrorResponse "The token is invalid"
// @Failure 500 {object} endpoints.ErrorResponse "Internal server error"
// @Router /v1/user/start-session [post]
func StartSession(c *gin.Context) {
	ctx, span := endpoints.Tracer.Start(c.Request.Context(), "StartSession")
	defer span.End()

	// Get the user id from the context
	userId := authHelper.GetUserIdFromContext(ctx, c)

	// Generate the access token
	accessJWT, err2 := authHelper.GenerateToken(ctx, userId, authHelper.AccessToken)
	if err2 != nil {
		err2 = errors.Wrap(err2, "Failed to generate access token")
		endpoints.Logger.Error(ctx, err2)
		c.JSON(
			http.StatusInternalServerError,
			endpoints.ErrorResponse{
				Error: "Internal server error",
			},
		)
		return
	}

	// Set the access token as a cookie
	c.SetCookie(string(authHelper.AccessToken), accessJWT, accessTokenDurationMinutes*60, "/", domain, secure, true)

	c.JSON(
		http.StatusOK,
		AccessTokenHolder{
			AccessToken: accessJWT,
		},
	)
}
