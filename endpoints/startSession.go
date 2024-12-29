package endpoints

import (
	"github.com/Team-Reissdorf/Backend/authHelper"
	"github.com/Team-Reissdorf/Backend/endpoints/standardJsonAnswers"
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
// @Param Authorization  header  string  true "Refresh JWT"
// @Success 200 {object} TokenHolder "Session start successful"
// @Failure 401 {object} standardJsonAnswers.ErrorResponse "The token is invalid"
// @Failure 500 {object} standardJsonAnswers.ErrorResponse "Internal server error"
// @Router /v1/user/start-session [post]
func StartSession(c *gin.Context) {
	ctx, span := tracer.Start(c.Request.Context(), "StartSession")
	defer span.End()

	// Get the user id from the context
	userId := authHelper.GetUserIdFromContext(ctx, c)

	// Generate the access token
	accessJWT, err2 := authHelper.GenerateToken(ctx, userId, authHelper.AccessToken)
	if err2 != nil {
		err2 = errors.Wrap(err2, "Failed to generate access token")
		logger.Error(ctx, err2)
		c.JSON(
			http.StatusInternalServerError,
			standardJsonAnswers.ErrorResponse{
				Error: "Internal server error",
			},
		)
		return
	}

	c.JSON(
		http.StatusOK,
		TokenHolder{
			Token: accessJWT,
		},
	)
}
