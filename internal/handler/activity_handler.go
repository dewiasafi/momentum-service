package handler

import (
	"encoding/json"
	"net/http"

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
	activities, err := h.Repo.GetAll()
	if err != nil {
		helper.Error(w, http.StatusInternalServerError, "Gagal mengambil data aktivitas")
		return
	}
	helper.JSON(w, http.StatusOK, "Berhasil mengambil daftar aktivitas", activities)
}

func (h *ActivityHandler) CreateActivity(w http.ResponseWriter, r *http.Request) {
	var newActivity model.Activity
	if err := json.NewDecoder(r.Body).Decode(&newActivity); err != nil {
		helper.Error(w, http.StatusBadRequest, "Format JSON tidak valid")
		return
	}

	if newActivity.Title == "" {
		helper.Error(w, http.StatusBadRequest, "Judul aktivitas tidak boleh kosong")
		return
	}

	if err := h.Repo.Create(&newActivity); err != nil {
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

	if err := h.Repo.Update(id, &act); err != nil {
		if err.Error() == "activity not found" {
			helper.Error(w, http.StatusNotFound, "Activity Not Found")
			return
		}
		return
	}

	helper.JSON(w, http.StatusOK, "Berhasil memperbarui aktivitas", act)
}

func (h *ActivityHandler) DeleteActivity(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := h.Repo.Delete(id); err != nil {
		if err.Error() == "activity not found" {
			helper.Error(w, http.StatusNotFound, "Aktivitas tidak ditemukan")
			return
		}
		helper.Error(w, http.StatusInternalServerError, "Gagal menghapus aktivitas")
		return
	}

	// 3. Jika sukses -> 200 OK dengan format JSON Envelope yang konsisten
	helper.JSON(w, http.StatusOK, "Aktivitas berhasil dihapus", nil)
}
