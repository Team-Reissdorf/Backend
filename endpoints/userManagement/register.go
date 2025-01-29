package userManagement

import (
	"github.com/Team-Reissdorf/Backend/authHelper"
	"github.com/Team-Reissdorf/Backend/databaseModels"
	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/Team-Reissdorf/Backend/formatHelper"
	"github.com/Team-Reissdorf/Backend/hashingHelper"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"net/http"
)

type UserBody struct {
	Email      string `json:"email" binding:"required,email" example:"bob.alice@example.com"`
	Password   string `json:"password" binding:"required" example:"<password>"`
	RememberMe bool   `json:"remember_me" example:"true"`
}

// Register handles the user registration process.
// @Summary Register a new user
// @Description Creates a new user account in the database with the provided details.
// @Tags User Management
// @Accept json
// @Produce json
// @Param User body UserBody true "The user's email address and password, along with a 'remember_me' field. If set to false or left empty, the refresh token cookie will not have a maxAge flag, causing the browser to automatically delete it when the session ends."
// @Success 201 {object} DoubleTokenHolder "Registration successful"
// @Failure 400 {object} endpoints.ErrorResponse "Invalid request body"
// @Failure 409 {object} endpoints.ErrorResponse "User already exists"
// @Failure 500 {object} endpoints.ErrorResponse "Internal server error"
// @Router /v1/user/register [post]
func Register(c *gin.Context) {
	ctx, span := endpoints.Tracer.Start(c.Request.Context(), "RegisterUser")
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

	// Hash the given password
	hash, err1 := hashingHelper.DefaultHashParams.HashPassword(ctx, body.Password)
	if err1 != nil {
		err1 = errors.Wrap(err1, "Failed to hash password")
		endpoints.Logger.Error(ctx, err1)
		c.JSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Internal server error"})
		c.Abort()
		return
	}

	// Create the user
	trainer := databaseModels.Trainer{
		Email:    body.Email,
		Password: hash,
	}

	// Write the user to the database
	err2 := createTrainer(ctx, trainer)
	if errors.Is(err2, TrainerAlreadyExistsErr) {
		endpoints.Logger.Debug(ctx, err2)
		c.JSON(http.StatusConflict, endpoints.ErrorResponse{Error: "Trainer already exists"})
		c.Abort()
		return
	} else if err2 != nil {
		err2 = errors.Wrap(err2, "Failed to create the trainer")
		endpoints.Logger.Error(ctx, err2)
		c.JSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Internal server error"})
		c.Abort()
		return
	}

	// Generate the refresh token
	refreshJWT, err3 := authHelper.GenerateToken(ctx, trainer.Email, authHelper.RefreshToken, body.RememberMe)
	if err3 != nil {
		err3 = errors.Wrap(err3, "Failed to generate refresh token")
		endpoints.Logger.Error(ctx, err3)
		c.JSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Internal server error"})
		c.Abort()
		return
	}

	// Generate an access token
	accessJWT, err4 := authHelper.GenerateToken(ctx, trainer.Email, authHelper.AccessToken, body.RememberMe)
	if err4 != nil {
		err4 = errors.Wrap(err4, "Failed to generate access token")
		endpoints.Logger.Error(ctx, err4)
		c.JSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Internal server error"})
		c.Abort()
		return
	}

	// Set cookies for the client to store the tokens
	SetCookies(c, &accessJWT, &refreshJWT, body.RememberMe)

	c.JSON(
		http.StatusCreated,
		DoubleTokenHolder{
			RefreshToken: refreshJWT,
			AccessToken:  accessJWT,
		},
	)
}
