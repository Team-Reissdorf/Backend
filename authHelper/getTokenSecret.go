package authHelper

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
)

var (
	InvalidTokenClaimsError    = errors.New("Invalid token claims")
	TokenTypeNotSupportedError = errors.New("Token type not supported")
)

// getSecretKey returns the secret key for the HMAC algorithm based on the TokenType claim
func getSecretKey(token *jwt.Token) (interface{}, error) {
	ctx := context.Background()
	ctx, span := tracer.Start(ctx, "GetJWTSecretKey")
	defer span.End()

	// Check if the claims are valid
	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		logger.Error(ctx, fmt.Sprintf("Invalid token claims: %v", token.Claims))
		return nil, InvalidTokenClaimsError
	}

	// Get the secret key of the token type
	secretKey, err := getSecretKeyByTokenType(ctx, claims.Name)
	if err != nil {
		return nil, err
	}

	return secretKey, nil
}

// getSecretKeyByTokenType returns the secret key for the HMAC algorithm based on the token type
func getSecretKeyByTokenType(ctx context.Context, tokenType string) ([]byte, error) {
	_, span := tracer.Start(ctx, "GetJWTSecretKeyByType")
	defer span.End()

	switch tokenType {
	case string(AccessToken):
		return accessTokenSecretKey, nil
	case string(RefreshToken):
		return refreshTokenSecretKey, nil
	case string(SettingsAccessToken):
		return settingsAccessTokenSecretKey, nil
	default:
		return nil, TokenTypeNotSupportedError
	}
}
