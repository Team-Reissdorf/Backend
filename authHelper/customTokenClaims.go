package authHelper

import "github.com/golang-jwt/jwt/v5"

// CustomClaims is the struct to define the claim structure used by the JWT tokens
type CustomClaims struct {
	jwt.RegisteredClaims
	Name       string `json:"name"`
	RememberMe bool   `json:"remember_me"`
}
