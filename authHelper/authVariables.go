package authHelper

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/pkg/errors"
)

var accessTokenSecretKey []byte
var refreshTokenSecretKey []byte
var settingsAccessTokenSecretKey []byte

var accessTokenDurationMinutes time.Duration
var refreshTokenDurationDays time.Duration
var settingsAccessTokenDurationMinutes time.Duration

// init initializes the environment variables and the secret key for the HMAC algorithm
func init() {
	ctx := context.Background()

	// Load the environment variables
	if err := godotenv.Load(".env"); err != nil {
		logger.Fatal(ctx, "Failed to load environment variables")
	}

	// Get the secret key for the HMAC algorithm for the access token
	accessTokenSecretKey = []byte(os.Getenv("ACCESS_JWT_SECRET_KEY"))
	if len(accessTokenSecretKey) <= 12 {
		err := errors.New("ACCESS_JWT_SECRET_KEY not set or too short (should be at least 12 characters)")
		logger.Fatal(ctx, err)
	}

	// Get the secret key for the HMAC algorithm for the refresh token
	refreshTokenSecretKey = []byte(os.Getenv("REFRESH_JWT_SECRET_KEY"))
	if len(refreshTokenSecretKey) <= 12 {
		err := errors.New("REFRESH_JWT_SECRET_KEY not set or too short (should be at least 12 characters)")
		logger.Fatal(ctx, err)
	}

	// Get the secret key for the HMAC algorithm for the backendSettings access token
	settingsAccessTokenSecretKey = []byte(os.Getenv("SETTINGS_ACCESS_JWT_SECRET_KEY"))
	if len(settingsAccessTokenSecretKey) <= 12 {
		err := errors.New("SETTINGS_ACCESS_JWT_SECRET_KEY not set or too short (should be at least 12 characters)")
		logger.Fatal(ctx, err)
	}

	// Get the access token duration in minutes
	accessTokenDurationMinutesInt, err := strconv.Atoi(os.Getenv("ACCESS_TOKEN_DURATION_MINUTES"))
	if err != nil {
		err = errors.Wrap(err, "Failed to parse ACCESS_TOKEN_DURATION_MINUTES, using default")
		logger.Warn(ctx, err)
		accessTokenDurationMinutesInt = 15
	}
	accessTokenDurationMinutes = time.Duration(accessTokenDurationMinutesInt) * time.Minute

	// Get the refresh token duration in minutes
	refreshTokenDurationDaysInt, err := strconv.Atoi(os.Getenv("REFRESH_TOKEN_DURATION_DAYS"))
	if err != nil {
		err = errors.Wrap(err, "Failed to parse REFRESH_TOKEN_DURATION_DAYS, using default")
		logger.Warn(ctx, err)
		refreshTokenDurationDaysInt = 30
	}
	refreshTokenDurationDays = time.Duration(refreshTokenDurationDaysInt) * 24 * time.Hour

	// Get the backendSettings access token duration in minutes
	settingsAccessTokenDurationMinutesInt, err := strconv.Atoi(os.Getenv("SETTINGS_ACCESS_TOKEN_DURATION_MINUTES"))
	if err != nil {
		err = errors.Wrap(err, "Failed to parse SETTINGS_ACCESS_TOKEN_DURATION_MINUTES, using default")
		logger.Warn(ctx, err)
		settingsAccessTokenDurationMinutesInt = 15
	}
	settingsAccessTokenDurationMinutes = time.Duration(settingsAccessTokenDurationMinutesInt) * time.Minute
}
