package model

import "github.com/golang-jwt/jwt/v4"

// Claims описывает набор JWT-claims приложения.
//
// Структура расширяет стандартные зарегистрированные claims из пакета jwt
// пользовательским идентификатором. Используется при создании и проверке JWT,
// чтобы связать токен с конкретным пользователем.
type Claims struct {
	// RegisteredClaims содержит стандартные зарегистрированные JWT-claims,
	// такие как issuer, subject, audience, expiration time и другие.
	jwt.RegisteredClaims

	// UserID содержит идентификатор пользователя, связанного с JWT.
	UserID int
}
