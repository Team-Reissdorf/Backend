package authMiddleware

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
	"strings"
)

var (
	NoAuthorizationHeaderError      = errors.New("No Authorization header provided")
	InvalidAuthorizationHeaderError = errors.New("Invalid Authorization header format")
	UnexpectedSigningMethodError    = errors.New("Unexpected signing method")
	InvalidTokenSignatureError      = errors.New("Invalid token signature")
	TokenUnverifiableError          = errors.New("Token is unverifiable, likely problems with the keyFunc")
	TokenProblemError               = errors.New("Token problem occurred")
)

func parseAuthorizationHeader(ctx context.Context, authHeader string) (*jwt.Token, error) {
	ctx, span := tracer.Start(ctx, "ParseAuthorizationHeader")
	defer span.End()

	// Check if the Authorization header is provided
	if authHeader == "" {
		return nil, NoAuthorizationHeaderError
	}

	// Check if the Authorization header is in the correct format
	bearer := strings.Split(authHeader, "Bearer ")
	if len(bearer) < 2 || len(bearer) > 2 || bearer[0] != "" || bearer[1] == "" {
		logger.Debug(ctx, InvalidAuthorizationHeaderError)
		return nil, InvalidAuthorizationHeaderError
	}
	tokenString := bearer[1]

	// Validate and parse the token
	token, err1 := jwt.ParseWithClaims(tokenString, &CustomClaims{}, getSecretKey)
	if errors.Is(err1, jwt.ErrTokenSignatureInvalid) || errors.Is(err1, jwt.ErrTokenMalformed) ||
		errors.Is(err1, InvalidTokenClaimsError) || errors.Is(err1, TokenTypeNotSupportedError) {
		// jwt.ErrTokenSignatureInvalid: The token algorithm is invalid or the secret key doesn't match
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			err := errors.Wrap(UnexpectedSigningMethodError, fmt.Sprintf("%v", token.Header["alg"]))
			logger.Debug(ctx, err)
			return nil, UnexpectedSigningMethodError
		} else { // Secret key doesn't match or the signature is malformed
			err1 = errors.Wrap(err1, InvalidTokenSignatureError.Error())
			logger.Debug(ctx, err1)
			return nil, InvalidTokenSignatureError
		}
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
