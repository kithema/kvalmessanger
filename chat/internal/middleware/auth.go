package middleware

import (
	"context"
	"net/http"
	"strings"

	"2say/internal/jwt"
)

type contextKey string

const UserIDKey contextKey = "userID"

// AuthMiddleware - middleware для chi
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
			return
		}

		claims, err := jwt.ValidateToken(parts[1])
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Auth - для совместимости с http.HandlerFunc
func Auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		AuthMiddleware(next).ServeHTTP(w, r)
	}
}

func GetUserID(r *http.Request) int {
	userID, ok := r.Context().Value(UserIDKey).(int)
	if !ok {
		return 0
	}
	return userID
}