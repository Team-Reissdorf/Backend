package userManagement

import (
	"github.com/LucaSchmitz2003/DatabaseFlow"
	"github.com/Team-Reissdorf/Backend/authHelper"
	"github.com/Team-Reissdorf/Backend/databaseModels"
	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/Team-Reissdorf/Backend/formatHelper"
	"github.com/Team-Reissdorf/Backend/hashingHelper"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"
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
		c.JSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Invalid request body"})
		c.Abort()
		return
	}

	// Validate inputs
	if err := formatHelper.IsEmail(body.Email); err != nil {
		endpoints.Logger.Debug(ctx, err)
		c.JSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Invalid email address"})
		c.Abort()
		return
	}

	// Get trainer from the database
	var trainer databaseModels.Trainer
	err1 := DatabaseFlow.TransactionHandler(ctx, func(tx *gorm.DB) error {
		err := tx.Model(&databaseModels.Trainer{}).Where("email = ?", body.Email).First(&trainer).Error
		return err
	})
	if errors.Is(err1, gorm.ErrRecordNotFound) {
		err1 = errors.Wrap(err1, "User not found")
		endpoints.Logger.Debug(ctx, err1)
		c.JSON(http.StatusNotFound, endpoints.ErrorResponse{Error: "Trainer could not be found"})
		c.Abort()
		return
	} else if err1 != nil {
		err1 = errors.Wrap(err1, "Failed to find the trainer account")
		endpoints.Logger.Error(ctx, err1)
		c.JSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to login"})
		c.Abort()
		return
	}

	// Verify the password
	verified, err1 := hashingHelper.VerifyHash(ctx, trainer.Password, body.Password)
	if err1 != nil {
		err1 = errors.Wrap(err1, "Failed to verify password")
		endpoints.Logger.Error(ctx, err1)
		c.JSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Internal server error"})
		c.Abort()
		return
	}
	if !verified {
		c.JSON(http.StatusUnauthorized, endpoints.ErrorResponse{Error: "Wrong credentials"})
		c.Abort()
		return
	}

	// Generate the refresh token
	refreshJWT, err2 := authHelper.GenerateToken(ctx, trainer.Email, authHelper.RefreshToken, body.RememberMe)
	if err2 != nil {
		err2 = errors.Wrap(err2, "Failed to generate refresh token")
		endpoints.Logger.Error(ctx, err2)
		c.JSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Internal server error"})
		c.Abort()
		return
	}

	// Generate an access token
	accessJWT, err3 := authHelper.GenerateToken(ctx, trainer.Email, authHelper.AccessToken, body.RememberMe)
	if err3 != nil {
		err3 = errors.Wrap(err3, "Failed to generate access token")
		endpoints.Logger.Error(ctx, err3)
		c.JSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Internal server error"})
		c.Abort()
		return
	}

	// Set cookies for the client to store the tokens
	SetCookies(c, &accessJWT, &refreshJWT, body.RememberMe)

	c.JSON(
		http.StatusOK,
		DoubleTokenHolder{
			RefreshToken: refreshJWT,
			AccessToken:  accessJWT,
		},
	)
}
