package middleware

import (
	"Rest-user-agregator/internal/authentication"
	"Rest-user-agregator/pkg/logger"
	"net/http"
	"os"
)

// CorsMiddleware — настраиваемый CORS
func CorsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("CorsMiddleware: ENTRY, user_id=%v, email=%v, role=%v",
			r.Context().Value(authentication.UserIDKey),
			r.Context().Value(authentication.EmailKey),
			r.Context().Value(authentication.RoleKey),
		)
		port := os.Getenv("SERVER_PORT")
		if port == "" {
			port = "8087"
		}
		allowedOrigin := "http://localhost:" + port

		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		logger.Debug("CorsMiddleware: before next, user_id=%v, email=%v, role=%v",
			r.Context().Value(authentication.UserIDKey),
			r.Context().Value(authentication.EmailKey),
			r.Context().Value(authentication.RoleKey),
		)

		next(w, r)
	}
}
