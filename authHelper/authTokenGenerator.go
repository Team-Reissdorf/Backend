package authHelper

import (
	"context"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
	"time"
)

// GenerateToken generates a new token for the given user ID and token type
func GenerateToken(ctx context.Context, userId string, tokenType TokenType, rememberMe bool) (string, error) {
	ctx, span := tracer.Start(ctx, "GenerateToken")
	defer span.End()

	// Create the claims for the token
	claims := CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:  userId,
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
		Name:       string(tokenType),
		RememberMe: rememberMe,
	}

	// Create the token
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	// Get the secret key of the token type
	secretKey, err1 := getSecretKeyByTokenType(ctx, string(tokenType))
	if err1 != nil {
		err1 = errors.Wrap(err1, "Failed to get the secret key")
		return "", err1
	}

	// Sign the token
	tokenString, err2 := token.SignedString(secretKey)
	if err2 != nil {
		err2 = errors.Wrap(err2, "Failed to sign the token")
		return "", err2
	}

	return tokenString, nil
}
