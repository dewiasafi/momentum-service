package router

import (
	"net/http"

	"project.com/momentum-service/internal/handler"
	"project.com/momentum-service/internal/helper"
	"project.com/momentum-service/internal/middleware"
)

type RouterConfig struct {
	ActivityHandler *handler.ActivityHandler
}

func SetupRoutes(cfg RouterConfig) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		helper.JSON(w, http.StatusOK, "Momentum API is running", nil)
	})

	ActivityRoutes(mux, cfg.ActivityHandler)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		helper.Error(w, http.StatusNotFound, "Endpoint tidak ditemukan")
	})

	return middleware.CORS(middleware.StrictPath(mux))
}
