package middleware

import (
	"context"
	"net/http"
	"strings"

	"project.com/momentum-service/internal/helper"
)

type contextKey string

const UserIDKey contextKey = "user_id"

const (
	ContextUserID       contextKey = "userID"
	ContextSessionToken contextKey = "sessionTokenHash"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error": "Unauthorized: Token tidak ditemukan"}`, http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, `{"error": "Unauthorized: Format token salah (Gunakan Bearer <token>)"}`, http.StatusUnauthorized)
			return
		}

		tokenStr := parts[1]
		claims, err := helper.ValidateToken(tokenStr)
		if err != nil {
			http.Error(w, `{"error": "Unauthorized: Token tidak valid atau kedaluwarsa"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), ContextUserID, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))

	})
}
