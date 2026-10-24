package router

import (
	"net/http"

	"project.com/momentum-service/internal/handler"
	"project.com/momentum-service/internal/middleware"
	"project.com/momentum-service/internal/repository"
)

func JournalRoutes(
	mux *http.ServeMux,
	h *handler.JournalHandler,
	authRepo repository.AuthRepository,
) {
	mux.Handle("POST /api/v1/journals",
		middleware.AuthMiddleware(authRepo, http.HandlerFunc(h.CreateJournal)))

	mux.Handle("GET /api/v1/journals",
		middleware.AuthMiddleware(authRepo, http.HandlerFunc(h.ListJournals)))

	mux.Handle("GET /api/v1/journals/categories",
		middleware.AuthMiddleware(authRepo, http.HandlerFunc(h.ListCategories)))

	mux.Handle("GET /api/v1/journals/{id}",
		middleware.AuthMiddleware(authRepo, http.HandlerFunc(h.GetJournal)))

	mux.Handle("PUT /api/v1/journals/{id}",
		middleware.AuthMiddleware(authRepo, http.HandlerFunc(h.UpdateJournal)))

	mux.Handle("DELETE /api/v1/journals/{id}",
		middleware.AuthMiddleware(authRepo, http.HandlerFunc(h.DeleteJournal)))

	mux.Handle("GET /api/v1/journals/{id}/versions",
		middleware.AuthMiddleware(authRepo, http.HandlerFunc(h.ListVersions)))
}
