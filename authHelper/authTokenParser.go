package authHelper

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
)

var (
	NoAuthorizationTokenError    = errors.New("No Authorization token provided")
	UnexpectedSigningMethodError = errors.New("Unexpected signing method")
	InvalidTokenSignatureError   = errors.New("Invalid token signature")
	TokenUnverifiableError       = errors.New("Token is unverifiable, likely problems with the keyFunc")
	TokenProblemError            = errors.New("Token problem occurred")
)

// parseAuthorizationToken parses the Authorization token from the given string (JWT from the Authorization header or cookie)
func parseAuthorizationToken(ctx context.Context, tokenString string) (*jwt.Token, error) {
	ctx, span := tracer.Start(ctx, "ParseAuthorizationToken")
	defer span.End()

	// Check if the Authorization token is provided
	if tokenString == "" {
		return nil, NoAuthorizationTokenError
	}

	// Validate and parse the token
	token, err1 := jwt.ParseWithClaims(tokenString, &CustomClaims{}, getSecretKey)
	if errors.Is(err1, jwt.ErrTokenSignatureInvalid) || errors.Is(err1, jwt.ErrTokenMalformed) ||
		errors.Is(err1, InvalidTokenClaimsError) || errors.Is(err1, TokenTypeNotSupportedError) {
		// jwt.ErrTokenSignatureInvalid: The token algorithm is invalid or the secret key doesn't match
		if token != nil {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				err := errors.Wrap(UnexpectedSigningMethodError, fmt.Sprintf("%v", token.Header["alg"]))
				logger.Debug(ctx, err)
				return nil, UnexpectedSigningMethodError
			}
		}

		// Secret key doesn't match or the signature is malformed
		err1 = errors.Wrap(err1, InvalidTokenSignatureError.Error())
		logger.Debug(ctx, err1)
		return nil, InvalidTokenSignatureError
	} else if errors.Is(err1, jwt.ErrTokenUnverifiable) {
		err1 = errors.Wrap(err1, TokenUnverifiableError.Error())
		logger.Error(ctx, err1)
		return nil, TokenUnverifiableError
	} else if err1 != nil || !token.Valid { // Other errors or the token is invalid
		err1 = errors.Wrap(err1, "Token problem occurred")
		logger.Warn(ctx, err1)
		return nil, TokenProblemError
	}

	return token, nil
}
