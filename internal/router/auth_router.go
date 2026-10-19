package router

import (
	"net/http"

	"project.com/momentum-service/internal/handler"
	"project.com/momentum-service/internal/middleware"
	"project.com/momentum-service/internal/repository"
)

func AuthRoutes(
	mux *http.ServeMux,
	h *handler.AuthHandler,
	authRepo repository.AuthRepository,
) {

	// Public Routes
	mux.HandleFunc("POST /api/v1/auth/register", h.Register)
	mux.HandleFunc("POST /api/v1/auth/login", h.Login)
	mux.HandleFunc("GET /api/v1/auth/session", h.GetSession)
	mux.HandleFunc("POST /api/v1/auth/verify-email", h.VerifyEmail)
	mux.HandleFunc("POST /api/v1/auth/resend-verification", h.ResendVerification)

	// Protected Routes (Memerlukan JWT Token yang Valid)
	mux.Handle("POST /api/v1/auth/logout", middleware.AuthMiddleware(authRepo, http.HandlerFunc(h.Logout)))
	mux.Handle("GET /api/v1/profile/me", middleware.AuthMiddleware(authRepo, http.HandlerFunc(h.GetProfile)))
	mux.Handle("PATCH /api/v1/profile/me", middleware.AuthMiddleware(authRepo, http.HandlerFunc(h.UpdateProfile)))

}
