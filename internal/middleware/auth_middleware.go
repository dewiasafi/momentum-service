package middleware

import (
	"context"
	"database/sql"
	"net/http"
	"os"

	"project.com/momentum-service/internal/helper"
	"project.com/momentum-service/internal/repository"
)

type contextKey string

const UserIDKey contextKey = "user_id"

const (
	ContextUserID       contextKey = "userID"
	ContextSessionToken contextKey = "sessionTokenHash"
	SessionCookieName              = "momentum_session"
)

func AuthMiddleware(authRepo repository.AuthRepository, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(SessionCookieName)

		if err != nil || cookie.Value == "" {
			helper.Error(w, http.StatusUnauthorized, "Sesi tidak ditemukan. Silakan login.")
			return
		}

		tokenHash := helper.HashToken(cookie.Value)

		session, err := authRepo.GetSessionByTokenHash(r.Context(), tokenHash)

		if err != nil {
			if err == sql.ErrNoRows {
				clearSessionCookie(w, r)
				helper.Error(w, http.StatusUnauthorized, "Sesi tidak valid atau sudah kedaluwarsa.")
				return
			}
			helper.Error(w, http.StatusInternalServerError, "Gagal memvalidasi sesi.")
			return
		}

		ctx := context.WithValue(r.Context(), ContextUserID, session.UserID)
		ctx = context.WithValue(ctx, ContextSessionToken, tokenHash)
	})
}

func clearSessionCookie(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   isProduction(),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

func isProduction() bool {
	return os.Getenv("APP_ENV") == "production"
}
