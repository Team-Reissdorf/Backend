package authMiddleware

type TokenType string

const (
	AccessToken  TokenType = "ACCESS_TOKEN"
	RefreshToken TokenType = "REFRESH_TOKEN"
)
