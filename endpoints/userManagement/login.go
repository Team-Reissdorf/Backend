package userManagement

import (
	"github.com/Team-Reissdorf/Backend/authHelper"
	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/Team-Reissdorf/Backend/hashingHelper"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"net/http"
)

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
// @Param User body UserBody true "The user's email address and password, along with a 'remember_me' field. If set to false or left empty, the refresh token cookie will not have a maxAge flag, causing the browser to automatically delete it when the session ends."
// @Success 200 {object} DoubleTokenHolder "Login successful"
// @Failure 400 {object} endpoints.ErrorResponse "Invalid request body"
// @Failure 401 {object} endpoints.ErrorResponse "Wrong credentials"
// @Failure 404 {object} endpoints.ErrorResponse "User not found"
// @Failure 500 {object} endpoints.ErrorResponse "Internal server error"
// @Router /v1/user/login [post]
func Login(c *gin.Context) {
	ctx, span := endpoints.Tracer.Start(c.Request.Context(), "LoginUser")
	defer span.End()

	// Bind JSON body to struct
	var body UserBody
	if err := c.ShouldBindJSON(&body); err != nil {
		err = errors.Wrap(err, "Failed to bind JSON body")
		endpoints.Logger.Debug(ctx, err)
		c.JSON(
			http.StatusBadRequest,
			endpoints.ErrorResponse{
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
		endpoints.Logger.Error(ctx, err1)
		c.JSON(
			http.StatusInternalServerError,
			endpoints.ErrorResponse{
				Error: "Internal server error",
			},
		)
		return
	}
	if !verified {
		c.JSON(
			http.StatusUnauthorized,
			endpoints.ErrorResponse{
				Error: "Wrong credentials",
			},
		)
		return
	}

	// Generate the refresh token
	refreshJWT, err2 := authHelper.GenerateToken(ctx, userId, authHelper.RefreshToken, body.RememberMe)
	if err2 != nil {
		err2 = errors.Wrap(err2, "Failed to generate refresh token")
		endpoints.Logger.Error(ctx, err2)
		c.JSON(
			http.StatusInternalServerError,
			endpoints.ErrorResponse{
				Error: "Internal server error",
			},
		)
		return
	}

	// Generate an access token
	accessJWT, err3 := authHelper.GenerateToken(ctx, userId, authHelper.AccessToken, body.RememberMe)
	if err3 != nil {
		err3 = errors.Wrap(err3, "Failed to generate access token")
		endpoints.Logger.Error(ctx, err3)
		c.JSON(
			http.StatusInternalServerError,
			endpoints.ErrorResponse{
				Error: "Internal server error",
			},
		)
		return
	}

	// Set cookies for the client to store the tokens
	accessToken := &http.Cookie{
		Name:     string(authHelper.AccessToken),
		Value:    accessJWT,
		MaxAge:   accessTokenDurationMinutes * 60,
		Path:     "/",
		Domain:   domain,
		Secure:   secure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	refreshToken := &http.Cookie{
		Name:     string(authHelper.RefreshToken),
		Value:    refreshJWT,
		MaxAge:   refreshTokenDurationDays * 24 * 60 * 60,
		Path:     path,
		Domain:   domain,
		Secure:   secure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(c.Writer, accessToken)
	http.SetCookie(c.Writer, refreshToken)

	c.JSON(
		http.StatusOK,
		DoubleTokenHolder{
			RefreshToken: refreshJWT,
			AccessToken:  accessJWT,
		},
	)
}
