package middleware

import (
	"net/http"
	"os"
)

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		allowedOrigin := os.Getenv("FRONTEND_ORIGIN")
		if allowedOrigin == "" {
			http.Error(
				w,
				"FRONTEND_ORIGIN belum disetting",
				http.StatusInternalServerError,
			)
		}

		origin := r.Header.Get("Origin")

		if origin != "" && origin == allowedOrigin {
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Vary", "Origin")
			w.Header().Set(
				"Access-Control-Allow-Methods",
				"GET, POST, PUT, PATCH, DELETE, OPTIONS",
			)
			w.Header().Set(
				"Access-Control-Allow-Headers",
				"Content-Type, X-CSRF-Token",
			)
		}

		if r.Method == http.MethodOptions {
			if origin != allowedOrigin {
				http.Error(w, "Origin tidak diizinkan", http.StatusForbidden)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
