package userManagement

import (
	"context"
	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"os"
	"strconv"
)

var (
	domain, path                                         string
	secure                                               bool
	refreshTokenDurationDays, accessTokenDurationMinutes int
)

func init() {
	ctx := context.Background()

	// Load the environment variables
	if err := godotenv.Load(".env"); err != nil {
		endpoints.Logger.Fatal(ctx, "Failed to load environment variables")
	}

	// Get the domain name for the cookies
	domain = os.Getenv("DOMAIN")
	if domain == "" {
		err := errors.New("DOMAIN not set, using default")
		endpoints.Logger.Error(ctx, err)
		domain = "localhost"
	}

	// Get the path for the cookies
	path = os.Getenv("REFRESH_TOKEN_USAGE_PATH")
	if path == "" {
		err := errors.New("REFRESH_TOKEN_USAGE_PATH not set, using default")
		endpoints.Logger.Error(ctx, err)
		path = "/"
	}

	// Get the secure flag for the cookies
	var err1 error
	secure, err1 = strconv.ParseBool(os.Getenv("TOKEN_SECURE_FLAG"))
	if err1 != nil {
		err1 = errors.Wrap(err1, "Failed to parse TOKEN_SECURE_FLAG, using default")
		endpoints.Logger.Error(ctx, err1)
		secure = false
	}

	// Get the refresh token duration in days
	var err2 error
	refreshTokenDurationDays, err2 = strconv.Atoi(os.Getenv("REFRESH_TOKEN_DURATION_DAYS"))
	if err2 != nil {
		err2 = errors.Wrap(err2, "Failed to parse REFRESH_TOKEN_DURATION_DAYS, using default")
		endpoints.Logger.Error(ctx, err2)
		refreshTokenDurationDays = 30
	}

	// Get the access token duration in minutes
	var err3 error
	accessTokenDurationMinutes, err3 = strconv.Atoi(os.Getenv("ACCESS_TOKEN_DURATION_MINUTES"))
	if err3 != nil {
		err3 = errors.Wrap(err3, "Failed to parse ACCESS_TOKEN_DURATION_MINUTES, using default")
		endpoints.Logger.Error(ctx, err3)
		accessTokenDurationMinutes = 15
	}
}
