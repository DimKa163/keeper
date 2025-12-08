// Package auth tools for authentification
package auth

import (
	"github.com/beevik/guid"
	"github.com/golang-jwt/jwt/v4"
)

// Claims user information
type Claims struct {
	jwt.RegisteredClaims
}

// Engine Authentification engine
type Engine interface {
	GenerateToken(userID guid.Guid) (string, error)
	ReadToken(tokenString string) (*Claims, error)
}
