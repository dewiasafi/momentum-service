package router

import (
	"net/http"

	"project.com/momentum-service/internal/handler"
	"project.com/momentum-service/internal/helper"
	"project.com/momentum-service/internal/middleware"
	"project.com/momentum-service/internal/repository"
)

type RouterConfig struct {
	ActivityHandler *handler.ActivityHandler
	AuthHandler     *handler.AuthHandler
	JournalHandler  *handler.JournalHandler
	AuthRepo        repository.AuthRepository
}

func SetupRoutes(cfg RouterConfig) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		helper.JSON(w, http.StatusOK, "Momentum API is running", nil)
	})

	ActivityRoutes(mux, cfg.ActivityHandler)
	AuthRoutes(mux, cfg.AuthHandler, cfg.AuthRepo)
	JournalRoutes(mux, cfg.JournalHandler, cfg.AuthRepo)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		helper.Error(w, http.StatusNotFound, "Endpoint tidak ditemukan")
	})

	return middleware.CORS(
		middleware.CSRFMiddleware(
			middleware.StrictPath(mux),
		),
	)
}
