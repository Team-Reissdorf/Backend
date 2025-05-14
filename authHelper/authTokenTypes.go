package authHelper

type TokenType string

const (
	AccessToken         TokenType = "ACCESS_TOKEN"
	RefreshToken        TokenType = "REFRESH_TOKEN"
	SettingsAccessToken TokenType = "SETTINGS_ACCESS_TOKEN"
)
