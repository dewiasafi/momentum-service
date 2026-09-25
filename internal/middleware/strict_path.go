package middleware

import (
	"net/http"
	"strings"

	"project.com/momentum-service/internal/helper"
)

func StrictPath(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.RequestURI, "//") {
			helper.Error(w, http.StatusNotFound, "Endpoint Not Found ")
			return
		}
		next.ServeHTTP(w, r)
	})
}
