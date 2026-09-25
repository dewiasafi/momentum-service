package router

import (
	"net/http"

	"project.com/momentum-service/internal/handler"
)

func RegisterActivityRoutes(mux *http.ServeMux, h *handler.ActivityHandler) {
	mux.HandleFunc("GET /api/v1/activities", h.GetActivities)
	mux.HandleFunc("POST /api/v1/activities", h.CreateActivity)
	mux.HandleFunc("PUT /api/v1/activities/{id}", h.UpdateActivity)
	mux.HandleFunc("DELETE /api/v1/activities/{id}", h.DeleteActivity)
	mux.HandleFunc("GET /api/v1/activities/{id}", h.GetActivityByID)
}
