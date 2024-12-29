package authHelper

import (
	"fmt"
	"github.com/LucaSchmitz2003/FlowWatch"
	"github.com/Team-Reissdorf/Backend/endpoints/standardJsonAnswers"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	"net/http"
	"time"
)

var (
	tracer = otel.Tracer("EndpointMiddlewareTracer")
	logger = FlowWatch.GetLogHelper()

	UserNotFoundError = errors.New("User could not be found in the database")
)

// UserIdContextKey is the key to get the user id from the context
const UserIdContextKey = "userId"

// GetAuthMiddlewareFor returns the middleware func for the given token type to be used in the gin router
// Usage: <router>.<Method>(<Path>, authHelper.GetAuthMiddlewareFor(authHelper.<TokenType>), <Endpoint-Handler>)
func GetAuthMiddlewareFor(tokenType TokenType) func(c *gin.Context) {

	// Parses and validates the JWT from the authorization header,
	// then sets the user ID in the request context for the next handler.
	// Swag-Annotations to use in the endpoint handlers:
	// @Param Authorization  header  string  true  "<TokenType> JWT"
	// @Failure 401 {object} standardJsonAnswers.ErrorResponse "The token is invalid"
	// @Failure 500 {object} standardJsonAnswers.ErrorResponse "Internal server error"
	return func(c *gin.Context) {
		ctx, span := tracer.Start(c.Request.Context(), "AuthMiddleware")
		defer span.End()

		// Parse the Authorization header
		authHeader := c.GetHeader("Authorization")
		token, err1 := parseAuthorizationHeader(ctx, authHeader)
		if err1 != nil {
			if errors.Is(err1, NoAuthorizationHeaderError) || errors.Is(err1, InvalidAuthorizationHeaderError) || errors.Is(err1, UnexpectedSigningMethodError) || errors.Is(err1, InvalidTokenSignatureError) {
				c.JSON(http.StatusUnauthorized, err1.Error())
			} else {
				c.JSON(http.StatusInternalServerError, standardJsonAnswers.ErrorResponse{Error: "Token is unverifiable at the moment"})
			}
			c.Abort()
			return
		}
		if token == nil {
			logger.Error(ctx, "Token is nil")
			c.JSON(http.StatusInternalServerError, standardJsonAnswers.ErrorResponse{Error: "The token is empty"})
			c.Abort()
			return
		}

		// Check if the claims are valid
		claims, ok := token.Claims.(*CustomClaims)
		if !ok || !token.Valid {
			logger.Debug(ctx, fmt.Sprintf("Invalid token claims: %v", token.Claims))
			c.JSON(http.StatusUnauthorized, standardJsonAnswers.ErrorResponse{Error: "Invalid token claims"})
			c.Abort()
			return
		}

		// Check if the token type is correct
		if claims.Name != string(tokenType) {
			logger.Debug(ctx, fmt.Sprintf("Invalid token type: %v", claims.Name))
			c.JSON(http.StatusUnauthorized, standardJsonAnswers.ErrorResponse{Error: "Invalid token type"})
			c.Abort()
			return
		}

		// Check if the token is expired
		issuedAt, err2 := claims.GetIssuedAt()
		if err2 != nil || issuedAt == nil {
			err2 = errors.Wrap(err2, "Failed to get the issued at time")
			logger.Debug(ctx, err2)
			c.JSON(http.StatusUnauthorized, standardJsonAnswers.ErrorResponse{Error: "Token is invalid"})
			c.Abort()
			return
		}
		// Use the correct duration for the token type
		var expiredThreshold time.Time
		switch tokenType {
		case AccessToken:
			expiredThreshold = issuedAt.Time.Add(accessTokenDurationMinutes)
		case RefreshToken:
			expiredThreshold = issuedAt.Time.Add(refreshTokenDurationDays)
		case SettingsAccessToken:
			expiredThreshold = issuedAt.Time.Add(settingsAccessTokenDurationMinutes)
		default:
			logger.Error(ctx, "Token type is not implemented yet")
			c.JSON(http.StatusInternalServerError, standardJsonAnswers.ErrorResponse{Error: "Token type cannot be handled"})
			c.Abort()
			return
		}
		if time.Now().After(expiredThreshold) {
			logger.Debug(ctx, "Token is expired")
			c.JSON(http.StatusUnauthorized, standardJsonAnswers.ErrorResponse{Error: "Token is expired"})
			c.Abort()
			return
		}

		// Get the user id from the claims
		userId, err3 := claims.GetSubject()
		if err3 != nil || userId == "" {
			err3 = errors.Wrap(err3, "Failed to get the user id from the claims")
			logger.Debug(ctx, err3)
			c.JSON(http.StatusUnauthorized, standardJsonAnswers.ErrorResponse{Error: "Token is invalid"})
			c.Abort()
			return
		}

		// Check if the user exists and the status is active
		active, err4 := isUserActive(ctx, userId)
		if errors.Is(err4, UserNotFoundError) {
			err4 = errors.Wrap(err4, "User not found")
			logger.Debug(ctx, err4)
			c.JSON(http.StatusUnauthorized, standardJsonAnswers.ErrorResponse{Error: "User not found"})
			c.Abort()
			return
		} else if err4 != nil {
			err4 = errors.Wrap(err4, "Failed to check if the user is active")
			logger.Debug(ctx, err4)
			c.JSON(http.StatusInternalServerError, standardJsonAnswers.ErrorResponse{Error: "Internal server error"})
			c.Abort()
			return
		}
		// Check if the user is active
		if !active {
			logger.Debug(ctx, "The user status is not marked as active")
			c.JSON(http.StatusUnauthorized, standardJsonAnswers.ErrorResponse{Error: "The user status is not marked as active"})
			c.Abort()
			return
		}

		// Set the user id in the context for the next handler
		c.Set(UserIdContextKey, userId)

		// Go to the next handler
		c.Next()
	}
}
