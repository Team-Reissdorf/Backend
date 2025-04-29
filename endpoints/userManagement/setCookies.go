package userManagement

import (
	"net/http"

	"github.com/Team-Reissdorf/Backend/authHelper"
	"github.com/gin-gonic/gin"
)

// SetCookies sets access and refresh tokens as cookies.
// If 'rememberMe' is true, the cookies persist across sessions;
// otherwise they're deleted when the browser closes the session.
func SetCookies(c *gin.Context, accessJWT, refreshJWT *string, rememberMe bool) {

	if accessJWT != nil {
		cookie := createCookie(authHelper.AccessToken, *accessJWT, "/", rememberMe)

		// Add header to ask client to set the cookies
		http.SetCookie(c.Writer, cookie)
	}

	if refreshJWT != nil {
		cookie := createCookie(authHelper.RefreshToken, *refreshJWT, path, rememberMe)

		// Add header to ask client to set the cookies
		http.SetCookie(c.Writer, cookie)
	}
}

// createCookie generates an HTTP cookie for the given token type.
// If 'rememberMe' is true, the cookie includes a MaxAge, making it persistent for the specified duration.
func createCookie(cookieTokenType authHelper.TokenType, cookieJwt, cookiePath string, rememberMe bool) *http.Cookie {
	// Create the base structure of the cookie
	cookie := &http.Cookie{
		Name:     string(cookieTokenType),
		Value:    cookieJwt,
		Path:     cookiePath,
		Domain:   "localhost",
		Secure:   secure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}

	// Set MaxAge only if 'rememberMe' is true
	if rememberMe {
		switch cookieTokenType {
		case authHelper.AccessToken:
			cookie.MaxAge = accessTokenDurationMinutes * 60
		case authHelper.RefreshToken:
			cookie.MaxAge = refreshTokenDurationDays * 24 * 60 * 60
		}
	}

	return cookie
}
