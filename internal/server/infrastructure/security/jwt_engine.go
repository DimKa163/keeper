package security

import (
	"errors"
	"strings"
	"time"

	"github.com/DimKa163/keeper/internal/server/domain/auth"
	"github.com/golang-jwt/jwt/v4"
)

type JWTConfig struct {
	TokenExpiration time.Duration
	SecretKey       []byte
}
type JWTEngine struct {
	*JWTConfig
}

func NewJWTEngine(config *JWTConfig) *JWTEngine {
	return &JWTEngine{config}
}

func (engine *JWTEngine) GenerateToken(userID int64) (string, error) {
	if engine.SecretKey == nil {
		return "", errors.New("secret key is required")
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(engine.TokenExpiration)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString(engine.SecretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (engine *JWTEngine) ReadToken(tokenString string) (*auth.Claims, error) {
	if engine.SecretKey == nil {
		return nil, errors.New("secret key is required")
	}
	var claims auth.Claims
	token, err := jwt.ParseWithClaims(strings.ReplaceAll(tokenString, "Bearer ", ""), &claims, func(token *jwt.Token) (interface{}, error) {
		return engine.SecretKey, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	return &claims, nil
}
