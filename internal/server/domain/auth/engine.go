package auth

import "github.com/golang-jwt/jwt/v4"

// Claims user information
type Claims struct {
	jwt.RegisteredClaims
	UserID int64
}

// Engine Authentification engine
type Engine interface {
	GenerateToken(userID int64) (string, error)
	ReadToken(tokenString string) (*Claims, error)
}
