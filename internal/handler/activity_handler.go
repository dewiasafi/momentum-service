package handler

import (
	"encoding/json"
	"errors"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"project.com/momentum-service/internal/helper"
	"project.com/momentum-service/internal/model"
	"project.com/momentum-service/internal/repository"
)

type ActivityHandler struct {
	Repo *repository.ActivityRepository
}

func NewActivityHandler(repo *repository.ActivityRepository) *ActivityHandler {
	return &ActivityHandler{Repo: repo}
}

func (h *ActivityHandler) GetActivities(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	page, _ := strconv.Atoi(q.Get("page"))
	if page <= 0 {
		page = 1
	}

	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	filter := repository.ActivityFilter{
		Page:      page,
		Limit:     limit,
		Category:  q.Get("category"),
		Priority:  q.Get("priority"),
		Search:    q.Get("search"),
		SortBy:    q.Get("sort_by"),
		SortOrder: q.Get("order"),
	}

	if completedStr := q.Get("is_completed"); completedStr != "" {
		if completedBool, err := strconv.ParseBool(completedStr); err == nil {
			filter.IsCompleted = &completedBool
		}
	}

	if startDateStr := q.Get("start_date"); startDateStr != "" {
		parsedStart, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			helper.Error(w, http.StatusBadRequest, "Format start_date tidak valid. Gunakan format YYYY-MM-DD")
			return
		}
		filter.StartDate = &parsedStart
	}

	if endDateStr := q.Get("end_date"); endDateStr != "" {
		parsedEnd, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			helper.Error(w, http.StatusBadRequest, "Format end_date tidak valid. Gunakan format YYYY-MM-DD")
			return
		}
		endOfDay := time.Date(parsedEnd.Year(), parsedEnd.Month(), parsedEnd.Day(), 23, 59, 59, 999999999, parsedEnd.Location())
		filter.EndDate = &endOfDay
	}

	activities, totalItems, err := h.Repo.GetAllFiltered(filter)
	if err != nil {
		log.Printf("[GetActivities Error]: %v\n", err)
		helper.Error(w, http.StatusInternalServerError, "Gagal mengambil data aktivitas")
		return
	}

	totalPages := 0
	if totalItems > 0 {
		totalPages = int(math.Ceil(float64(totalItems) / float64(limit)))
	}

	// 5. Envelope Response
	responseData := map[string]interface{}{
		"items":       activities,
		"page":        page,
		"limit":       limit,
		"total_items": totalItems,
		"total_pages": totalPages,
	}

	helper.JSON(w, http.StatusOK, "Berhasil mengambil daftar aktivitas", responseData)
}

func (h *ActivityHandler) GetActivityByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if id == "" {
		helper.Error(w, http.StatusBadRequest, "ID aktivitas wajib diisi")
		return
	}

	activity, err := h.Repo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			log.Printf("[GetActivityByID Error]: %v\n", err)
			helper.Error(w, http.StatusNotFound, "Aktivitas tidak ditemukan")
			return
		}
		log.Printf("[GetActivityByID Error]: %v\n", err)
		helper.Error(w, http.StatusInternalServerError, "Gagal mengambil detail aktivitas")
		return
	}

	helper.JSON(w, http.StatusOK, "Berhasil mengambil detail aktivitas", activity)
}

func (h *ActivityHandler) CreateActivity(w http.ResponseWriter, r *http.Request) {
	var newActivity model.Activity
	if err := json.NewDecoder(r.Body).Decode(&newActivity); err != nil {
		helper.Error(w, http.StatusBadRequest, "Format JSON tidak valid")
		return
	}

	if err := newActivity.Validate(); err != nil {
		helper.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.Repo.Create(&newActivity); err != nil {
		log.Printf("[CreateActivity Error]: %v\n", err)
		helper.Error(w, http.StatusInternalServerError, "Gagal menyimpan aktivitas ke database")
		return
	}

	helper.JSON(w, http.StatusCreated, "Berhasil menambahkan aktivitas baru", newActivity)
}

func (h *ActivityHandler) UpdateActivity(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var act model.Activity
	if err := json.NewDecoder(r.Body).Decode(&act); err != nil {
		helper.Error(w, http.StatusBadRequest, "Format JSON tidak valid")
		return
	}

	act.ID = id

	if err := act.Validate(); err != nil {
		helper.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.Repo.Update(id, &act); err != nil {
		log.Printf("[UpdateActivity Error]: %v\n", err)
		if errors.Is(err, repository.ErrNotFound) {
			helper.Error(w, http.StatusNotFound, "Aktivitas tidak ditemukan")
			return
		}

		helper.Error(w, http.StatusInternalServerError, "Gagal memperbarui aktivitas")
		return
	}

	helper.JSON(w, http.StatusOK, "Berhasil memperbarui aktivitas", act)
}

func (h *ActivityHandler) DeleteActivity(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := h.Repo.Delete(id); err != nil {
		log.Printf("[DeleteActivity Error]: %v\n", err)
		if errors.Is(err, repository.ErrNotFound) {
			helper.Error(w, http.StatusNotFound, "Aktivitas tidak ditemukan")
			return
		}
		helper.Error(w, http.StatusInternalServerError, "Gagal menghapus aktivitas")
		return
	}

	helper.JSON(w, http.StatusOK, "Aktivitas berhasil dihapus", nil)
}
