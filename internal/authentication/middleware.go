// Package authentication - middleware для проверки JWT
package authentication

import (
	"context"
	"net/http"
	"strings"

	"Rest-user-agregator/pkg/logger"
)

type contextKey string

const (
	UserIDKey contextKey = "user_id"
	EmailKey  contextKey = "email"
	RoleKey   contextKey = "role"
)

// AuthMiddleware — проверяет JWT-токен в заголовке Authorization
// Если токен валиден — сохраняет данные пользователя в контексте.
// Если нет — возвращает 401 Unauthorized.
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			logger.Warn("AuthMiddleware: missing Authorization header")
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		const prefix = "Bearer "
		if !strings.HasPrefix(authHeader, prefix) {
			logger.Warn("AuthMiddleware: invalid Authorization header format")
			http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, prefix)
		claims, err := ValidateToken(tokenString)
		if err != nil {
			logger.Warn("AuthMiddleware: invalid token: %v", err)
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, EmailKey, claims.Email)
		ctx = context.WithValue(ctx, RoleKey, claims.Role)

		next(w, r.WithContext(ctx))
	}
}

// GetUserID — возвращает user_id из контекста
func GetUserID(ctx context.Context) string {
	if v := ctx.Value(UserIDKey); v != nil {
		return v.(string)
	}
	return ""
}