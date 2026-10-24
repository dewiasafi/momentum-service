package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"project.com/momentum-service/internal/helper"
	"project.com/momentum-service/internal/middleware"
	"project.com/momentum-service/internal/model"
	"project.com/momentum-service/internal/repository"
)

type JournalHandler struct {
	Repo *repository.JournalRepository
}

func NewJournalHandler(repo *repository.JournalRepository) *JournalHandler {
	return &JournalHandler{Repo: repo}
}

type journalRequest struct {
	Title       string   `json:"title"`
	Content     string   `json:"content"`
	CategoryIDs []string `json:"category_ids"`
}

func journalUserID(r *http.Request) (string, bool) {
	userID, ok := r.Context().Value(middleware.ContextUserID).(string)
	return userID, ok && userID != ""
}

func journalID(r *http.Request) string {
	return strings.TrimSpace(r.PathValue("id"))
}

func decodeJournalRequest(r *http.Request) (journalRequest, error) {
	var req journalRequest
	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&req); err != nil {
		return req, err
	}

	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		return req, errors.New("judul jurnal wajib diisi")
	}
	if len([]rune(req.Title)) > 200 {
		return req, errors.New("judul maksimal 200 karakter")
	}

	return req, nil
}

func (h *JournalHandler) CreateJournal(w http.ResponseWriter, r *http.Request) {
	userID, ok := journalUserID(r)
	if !ok {
		helper.Error(w, http.StatusUnauthorized, "Sesi tidak valid")
		return
	}

	req, err := decodeJournalRequest(r)
	if err != nil {
		helper.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	journal := &model.Journal{
		Title:   req.Title,
		Content: req.Content,
	}

	if err := h.Repo.Create(r.Context(), userID, journal); err != nil {
		helper.Error(w, http.StatusInternalServerError, "Gagal membuat jurnal")
		return
	}

	if req.CategoryIDs != nil {
		if err := h.Repo.Update(
			r.Context(), userID, journal.ID, journal, req.CategoryIDs,
		); err != nil {
			helper.Error(w, http.StatusInternalServerError, "Jurnal dibuat, tetapi kategori gagal disimpan")
			return
		}
	}

	created, err := h.Repo.GetByID(r.Context(), userID, journal.ID)
	if err != nil {
		helper.Error(w, http.StatusInternalServerError, "Jurnal dibuat, tetapi gagal mengambil datanya")
		return
	}

	helper.JSON(w, http.StatusCreated, "Jurnal berhasil dibuat", created)
}

func (h *JournalHandler) ListJournals(w http.ResponseWriter, r *http.Request) {
	userID, ok := journalUserID(r)
	if !ok {
		helper.Error(w, http.StatusUnauthorized, "Sesi tidak valid")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	journals, err := h.Repo.List(r.Context(), userID, limit, offset)
	if err != nil {
		helper.Error(w, http.StatusInternalServerError, "Gagal mengambil daftar jurnal")
		return
	}

	helper.JSON(w, http.StatusOK, "Daftar jurnal berhasil diambil", journals)
}

func (h *JournalHandler) GetJournal(w http.ResponseWriter, r *http.Request) {
	userID, ok := journalUserID(r)
	if !ok {
		helper.Error(w, http.StatusUnauthorized, "Sesi tidak valid")
		return
	}

	journal, err := h.Repo.GetByID(r.Context(), userID, journalID(r))
	if errors.Is(err, repository.ErrJournalNotFound) {
		helper.Error(w, http.StatusNotFound, "Jurnal tidak ditemukan")
		return
	}
	if err != nil {
		helper.Error(w, http.StatusInternalServerError, "Gagal mengambil jurnal")
		return
	}

	helper.JSON(w, http.StatusOK, "Jurnal berhasil diambil", journal)
}

func (h *JournalHandler) UpdateJournal(w http.ResponseWriter, r *http.Request) {
	userID, ok := journalUserID(r)
	if !ok {
		helper.Error(w, http.StatusUnauthorized, "Sesi tidak valid")
		return
	}

	req, err := decodeJournalRequest(r)
	if err != nil {
		helper.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	journal := &model.Journal{
		Title:   req.Title,
		Content: req.Content,
	}

	err = h.Repo.Update(
		r.Context(), userID, journalID(r), journal, req.CategoryIDs,
	)
	if errors.Is(err, repository.ErrJournalNotFound) {
		helper.Error(w, http.StatusNotFound, "Jurnal tidak ditemukan")
		return
	}
	if err != nil {
		helper.Error(w, http.StatusInternalServerError, "Gagal memperbarui jurnal")
		return
	}

	updated, err := h.Repo.GetByID(r.Context(), userID, journalID(r))
	if err != nil {
		helper.Error(w, http.StatusInternalServerError, "Jurnal diperbarui, tetapi gagal mengambil datanya")
		return
	}

	helper.JSON(w, http.StatusOK, "Jurnal berhasil diperbarui", updated)
}

func (h *JournalHandler) DeleteJournal(w http.ResponseWriter, r *http.Request) {
	userID, ok := journalUserID(r)
	if !ok {
		helper.Error(w, http.StatusUnauthorized, "Sesi tidak valid")
		return
	}

	err := h.Repo.Delete(r.Context(), userID, journalID(r))
	if errors.Is(err, repository.ErrJournalNotFound) {
		helper.Error(w, http.StatusNotFound, "Jurnal tidak ditemukan")
		return
	}
	if err != nil {
		helper.Error(w, http.StatusInternalServerError, "Gagal menghapus jurnal")
		return
	}

	helper.JSON(w, http.StatusOK, "Jurnal berhasil dihapus", nil)
}

func (h *JournalHandler) ListVersions(w http.ResponseWriter, r *http.Request) {
	userID, ok := journalUserID(r)
	if !ok {
		helper.Error(w, http.StatusUnauthorized, "Sesi tidak valid")
		return
	}

	versions, err := h.Repo.ListVersions(
		r.Context(), userID, journalID(r),
	)
	if errors.Is(err, repository.ErrJournalNotFound) {
		helper.Error(w, http.StatusNotFound, "Jurnal tidak ditemukan")
		return
	}
	if err != nil {
		helper.Error(w, http.StatusInternalServerError, "Gagal mengambil riwayat versi")
		return
	}

	helper.JSON(w, http.StatusOK, "Riwayat versi berhasil diambil", versions)
}

func (h *JournalHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.Repo.ListCategories(r.Context())
	if err != nil {
		helper.Error(w, http.StatusInternalServerError, "Gagal mengambil kategori")
		return
	}

	helper.JSON(w, http.StatusOK, "Daftar kategori berhasil diambil", categories)
}
