package middleware

import (
	"net/http"
	"os"
)

func CSRFMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			next.ServeHTTP(w, r)
			return
		}

		expectedOrigin := os.Getenv("FRONTEND_ORIGIN")
		requestOrigin := r.Header.Get("Origin")

		if expectedOrigin == "" {
			http.Error(
				w,
				"Konfigurasi FRONTEND_ORIGIN belum tersedia",
				http.StatusInternalServerError,
			)
			return
		}

		if requestOrigin == "" || requestOrigin != expectedOrigin {
			http.Error(
				w,
				"Request ditolak: origin tidak diizinkan",
				http.StatusForbidden,
			)
			return
		}

		next.ServeHTTP(w, r)
	})
}
