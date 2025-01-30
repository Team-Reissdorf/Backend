package userManagement

import (
	"github.com/Team-Reissdorf/Backend/authHelper"
	"github.com/gin-gonic/gin"
	"net/http"
)

// Logout handles the user logout process.
// @Summary Logout of a user account
// @Description Logs the user out by clearing cookies from the client’s browser.
// @Tags User Management
// @Produce json
// @Success 200 {object} endpoints.SuccessResponse "Logout successful"
// @Failure 500 {object} endpoints.ErrorResponse "Internal server error"
// @Router /v1/user/logout [post]
func Logout(c *gin.Context) {
	// Generate expired cookies with empty values to override the existing ones, and thus remove them
	accessJwtCookie := &http.Cookie{
		Name:     string(authHelper.AccessToken),
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Domain:   domain,
		Secure:   secure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	refreshJwtCookie := &http.Cookie{
		Name:     string(authHelper.RefreshToken),
		Value:    "",
		Path:     path,
		MaxAge:   -1,
		Domain:   domain,
		Secure:   secure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}

	// Set the cookies to the response to override the existing ones
	http.SetCookie(c.Writer, accessJwtCookie)
	http.SetCookie(c.Writer, refreshJwtCookie)
}
