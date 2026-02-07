package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/TMWF/url-shortener/internal/config"
	"github.com/TMWF/url-shortener/internal/logger"
	"github.com/TMWF/url-shortener/internal/model"
	"github.com/TMWF/url-shortener/internal/util"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
)

func JwtTokenMiddleware(config config.Config) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			jwtCookie, _ := r.Cookie("AuthToken")

			if jwtCookie != nil {
				token := jwtCookie.Value
				logger.GetLogger().Info("Got token from the cookie",
					zap.String("Encrypted token", token),
				)
				userID := getUserID(token, config)
				if userID > 0 {
					ctx := context.WithValue(r.Context(), util.USER_ID, userID)
					r = r.WithContext(ctx)
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

func getUserID(tokenString string, config config.Config) int {
	claims := &model.Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(config.SecretKey), nil
		})
	if err != nil {
		return -1
	}

	if !token.Valid {
		logger.GetLogger().Error("Token is not valid")
		return -1
	}

	logger.GetLogger().Info("Token is valid")
	return claims.UserID
}
