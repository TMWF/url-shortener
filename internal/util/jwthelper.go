package util

import (
	"time"

	"github.com/daixiang0/gci/pkg/config"
	"github.com/golang-jwt/jwt/v4"
)

type UserJWTBuilder interface {
	BuildJWTString(userID int) (string, error)
}

type defaultJWTHelper struct {
	config *config.Config
}

func NewJWTHelper(cfg *config.Config) *defaultJWTHelper {
	return &defaultJWTHelper{config: cfg}
}

func (helper *defaultJWTHelper) BuildJWTString(userID int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(helper.config.TokenExp)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(SECRET_KEY))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
