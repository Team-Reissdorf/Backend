package authHelper

import (
	"context"
	"github.com/Team-Reissdorf/Backend/endpoints/standardJsonAnswers"
	"github.com/gin-gonic/gin"
	"net/http"
)

// GetUserIdFromContext gets the user id from the gin context and returns it as a string.
// Swag-Annotations to use in the endpoint handlers:
// @Failure 500 {object} standardJsonAnswers.ErrorResponse "Internal server error"
func GetUserIdFromContext(ctx context.Context, c *gin.Context) string {
	ctx, span := tracer.Start(c.Request.Context(), "GetUserIdFromContext")
	defer span.End()

	// Get the user id from the context
	userId, exists := c.Get(UserIdContextKey)
	if !exists || userId == nil {
		logger.Error(ctx, "User ID not found in the context")
		c.JSON(
			http.StatusInternalServerError,
			standardJsonAnswers.ErrorResponse{
				Error: "Internal server error",
			},
		)
		c.Abort()
		return ""
	}

	// Assert the user id to a string
	userIdString, ok := userId.(string)
	if !ok {
		logger.Error(ctx, "Failed to assert user id to string")
		c.JSON(
			http.StatusInternalServerError,
			standardJsonAnswers.ErrorResponse{
				Error: "Internal server error",
			},
		)
		c.Abort()
		return ""
	}

	// Check if the user id is empty
	if len(userIdString) == 0 {
		logger.Error(ctx, "User ID is empty")
		c.JSON(
			http.StatusInternalServerError,
			standardJsonAnswers.ErrorResponse{
				Error: "Internal server error",
			},
		)
		c.Abort()
		return ""
	}

	return userIdString
}
