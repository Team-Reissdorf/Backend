package authHelper

import (
	"context"
	"net/http"

	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/gin-gonic/gin"
)

// GetUserIdFromContext gets the user id from the gin context and returns it as a string.
// Swag-Annotations to use in the endpoint handlers:
// @Failure 500 {object} endpoints.ErrorResponse "Internal server error"
// nolint:staticcheck
func GetUserIdFromContext(ctx context.Context, c *gin.Context) string {
	ctx, span := tracer.Start(c.Request.Context(), "GetUserIdFromContext")
	defer span.End()

	// Get the user id from the context
	userId, exists := c.Get(UserIdContextKey)
	if !exists || userId == nil {
		logger.Error(ctx, "User ID not found in the context")
		c.JSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Internal server error"})
		c.Abort()
		return ""
	}

	// Assert the user id to a string
	userIdString, ok := userId.(string)
	if !ok {
		logger.Error(ctx, "Failed to assert user id to string: ", userIdString)
		c.JSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Internal server error"})
		c.Abort()
		return ""
	}

	// Check if the user id is empty
	if len(userIdString) == 0 {
		logger.Error(ctx, "User ID is empty")
		c.JSON(
			http.StatusInternalServerError,
			endpoints.ErrorResponse{
				Error: "Internal server error",
			},
		)
		c.Abort()
		return ""
	}

	return userIdString
}

// GetBooleanFromContext gets a boolean value from the gin context according to the given contextKey
// and returns it as a boolean.
// Swag-Annotations to use in the endpoint handlers:
// @Failure 500 {object} endpoints.ErrorResponse "Internal server error"
func GetBooleanFromContext(ctx context.Context, c *gin.Context, contextKey string) bool {
	ctx, span := tracer.Start(ctx, "GetBooleanFromContext")
	defer span.End()

	// Get the value from the context
	value, exists := c.Get(contextKey)
	if !exists || value == nil {
		logger.Error(ctx, "Value not found in the context")
		c.JSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Internal server error"})
		c.Abort()
		return false
	}

	// Assert the value to a bool
	boolValue, ok := value.(bool)
	if !ok {
		logger.Error(ctx, "Failed to assert value to boolean")
		c.JSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Internal server error"})
		c.Abort()
		return false
	}

	return boolValue
}
