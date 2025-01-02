package endpoints

import (
	"context"
	"github.com/Team-Reissdorf/Backend/authHelper"
	"github.com/Team-Reissdorf/Backend/endpoints/standardJsonAnswers"
	"github.com/Team-Reissdorf/Backend/hashingHelper"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"net/http"
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
		logger.Fatal(ctx, "Failed to load environment variables")
	}

	// Get the domain name for the cookies
	domain = os.Getenv("DOMAIN")
	if domain == "" {
		err := errors.New("DOMAIN not set, using default")
		logger.Error(ctx, err)
		domain = "localhost"
	}

	// Get the path for the cookies
	path = os.Getenv("REFRESH_TOKEN_USAGE_PATH")
	if path == "" {
		err := errors.New("REFRESH_TOKEN_USAGE_PATH not set, using default")
		logger.Error(ctx, err)
		path = "/"
	}

	// Get the secure flag for the cookies
	var err1 error
	secure, err1 = strconv.ParseBool(os.Getenv("TOKEN_SECURE_FLAG"))
	if err1 != nil {
		err1 = errors.Wrap(err1, "Failed to parse TOKEN_SECURE_FLAG, using default")
		logger.Error(ctx, err1)
		secure = false
	}

	// Get the refresh token duration in days
	var err2 error
	refreshTokenDurationDays, err2 = strconv.Atoi(os.Getenv("REFRESH_TOKEN_DURATION_DAYS"))
	if err2 != nil {
		err2 = errors.Wrap(err2, "Failed to parse REFRESH_TOKEN_DURATION_DAYS, using default")
		logger.Error(ctx, err2)
		refreshTokenDurationDays = 30
	}

	// Get the access token duration in minutes
	var err3 error
	accessTokenDurationMinutes, err3 = strconv.Atoi(os.Getenv("ACCESS_TOKEN_DURATION_MINUTES"))
	if err3 != nil {
		err3 = errors.Wrap(err3, "Failed to parse ACCESS_TOKEN_DURATION_MINUTES, using default")
		logger.Error(ctx, err3)
		accessTokenDurationMinutes = 15
	}
}

type DoubleTokenHolder struct {
	RefreshToken string `json:"refresh-token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiI8dXNlci1pZD4iLCJuYW1lIjoiPHRva2VuLXR5cGU-IiwiaWF0IjoxNzM0Njk4NzEwfQ.hzvbcP77EO8dnEyy5i-OgoOp8MYYwslfwKx32ZKgrH8"`
	AccessToken  string `json:"access-token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiI8dXNlci1pZD4iLCJuYW1lIjoiPHRva2VuLXR5cGU-IiwiaWF0IjoxNzM0Njk4NzEwfQ.hzvbcP77EO8dnEyy5i-OgoOp8MYYwslfwKx32ZKgrH8"`
}

// Login handles the user login process.
// @Summary Login to an existing user account
// @Description Logs the user into the system with the provided credentials and returns a refresh-JWT.
// @Tags User Management
// @Accept json
// @Produce json
// @Param User body UserBody true "Email address and password of the user"
// @Success 200 {object} DoubleTokenHolder "Login successful"
// @Failure 400 {object} standardJsonAnswers.ErrorResponse "Invalid request body"
// @Failure 401 {object} standardJsonAnswers.ErrorResponse "Wrong credentials"
// @Failure 404 {object} standardJsonAnswers.ErrorResponse "User not found"
// @Failure 500 {object} standardJsonAnswers.ErrorResponse "Internal server error"
// @Router /v1/user/login [post]
func Login(c *gin.Context) {
	ctx, span := tracer.Start(c.Request.Context(), "LoginUser")
	defer span.End()

	// Bind JSON body to struct
	var body UserBody
	if err := c.ShouldBindJSON(&body); err != nil {
		err = errors.Wrap(err, "Failed to bind JSON body")
		logger.Debug(ctx, err)
		c.JSON(
			http.StatusBadRequest,
			standardJsonAnswers.ErrorResponse{
				Error: "Invalid request body",
			},
		)
		return
	}

	// ToDo: Implement the login process
	userId := "<user-id>"                                                                                                            // ToDo: Get from database
	hash := "$argon2id$v=19$m=65536,t=2,p=4$PL26GfocVx8cCYyUnYWJei5ihyAqS0snyTwtqdH4YT8$fxZMiVwi9F/1BCEFieYc9QAHiaOZbNxp6AsnIBJm9xY" // ToDo: Get from database

	// Verify the password
	verified, err1 := hashingHelper.VerifyHash(ctx, hash, body.Password)
	if err1 != nil {
		err1 = errors.Wrap(err1, "Failed to verify password")
		logger.Error(ctx, err1)
		c.JSON(
			http.StatusInternalServerError,
			standardJsonAnswers.ErrorResponse{
				Error: "Internal server error",
			},
		)
		return
	}
	if !verified {
		c.JSON(
			http.StatusUnauthorized,
			standardJsonAnswers.ErrorResponse{
				Error: "Wrong credentials",
			},
		)
		return
	}

	// Generate the refresh token
	refreshJWT, err2 := authHelper.GenerateToken(ctx, userId, authHelper.RefreshToken)
	if err2 != nil {
		err2 = errors.Wrap(err2, "Failed to generate refresh token")
		logger.Error(ctx, err2)
		c.JSON(
			http.StatusInternalServerError,
			standardJsonAnswers.ErrorResponse{
				Error: "Internal server error",
			},
		)
		return
	}

	// Generate an access token
	accessJWT, err3 := authHelper.GenerateToken(ctx, userId, authHelper.AccessToken)
	if err3 != nil {
		err3 = errors.Wrap(err3, "Failed to generate access token")
		logger.Error(ctx, err3)
		c.JSON(
			http.StatusInternalServerError,
			standardJsonAnswers.ErrorResponse{
				Error: "Internal server error",
			},
		)
		return
	}

	// Set cookies for the client to store the tokens
	c.SetCookie(string(authHelper.RefreshToken), refreshJWT, refreshTokenDurationDays*24*60*60, path, domain, secure, true)
	c.SetCookie(string(authHelper.AccessToken), accessJWT, accessTokenDurationMinutes*60, "/", domain, secure, true)

	c.JSON(
		http.StatusOK,
		DoubleTokenHolder{
			RefreshToken: refreshJWT,
			AccessToken:  accessJWT,
		},
	)
}
