package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"project.com/momentum-service/internal/helper"
	"project.com/momentum-service/internal/middleware"
	"project.com/momentum-service/internal/model"
	"project.com/momentum-service/internal/repository"
)

type AuthHandler struct {
	authRepo repository.AuthRepository
}

func NewAuthHandler(authRepo repository.AuthRepository) *AuthHandler {
	return &AuthHandler{authRepo: authRepo}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.Error(w, http.StatusBadRequest, "Format data tidak valid.")
		return
	}

	if !helper.IsValidEmail(req.Email) {
		helper.Error(w, http.StatusBadRequest, "Format email tidak valid.")
		return
	}

	if !helper.IsStrongPassword(req.Password) {
		helper.Error(w, http.StatusBadRequest, "Password harus lebih dari 8 karakter dan mengandung kombinasi huruf, angka, serta simbol.")
		return
	}

	existingUser, _ := h.authRepo.GetUserByEmail(r.Context(), req.Email)
	if existingUser != nil {
		helper.Error(w, http.StatusConflict, "Email sudah terdaftar.")
		return
	}

	hashedPassword, err := helper.HashPassword(req.Password)
	if err != nil {
		helper.Error(w, http.StatusInternalServerError, "Gagal memproses password.")
		return
	}

	user := &model.User{
		Email:    req.Email,
		Password: hashedPassword,
		Name:     req.Name,
	}

	if err := h.authRepo.CreateUser(r.Context(), user); err != nil {
		helper.Error(w, http.StatusInternalServerError, "Gagal membuat akun.")
		return
	}

	// Buat token acak 32 bytes (plain token dikirim ke link/email user)
	rawToken, err := helper.GenerateSessionToken()
	if err != nil {
		helper.Error(w, http.StatusInternalServerError, "Gagal memproses token verifikasi.")
		return
	}

	// Simpan hash SHA-256 token dengan masa berlaku 30 menit
	tokenHash := helper.HashToken(rawToken)
	vToken := &model.VerificationToken{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(30 * time.Minute),
	}

	if err := h.authRepo.CreateVerificationToken(r.Context(), vToken); err != nil {
		log.Printf("[Register Error] Gagal menyimpan token verifikasi: %v\n", err)
	}

	// Log tautan verifikasi ke konsol (simulasi pengiriman email)
	verificationURL := "http://localhost:5173/verify-email?token=" + rawToken
	log.Printf("\n========================================\n[EMAIL VERIFIKASI] Kirim ke: %s\nTautan: %s\n(Kedaluwarsa dalam 30 menit)\n========================================\n", user.Email, verificationURL)

	helper.JSON(w, http.StatusCreated, "Registrasi berhasil. Silakan cek email Anda untuk memverifikasi akun sebelum login.", map[string]interface{}{
		"user": model.UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			CreatedAt: user.CreatedAt,
		},
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.Error(w, http.StatusBadRequest, "Format data tidak valid.")
		return
	}

	user, err := h.authRepo.GetUserByEmail(r.Context(), req.Email)
	if err != nil || !helper.CheckPasswordHash(req.Password, user.Password) {
		helper.Error(w, http.StatusUnauthorized, "Email atau password salah.")
		return
	}

	// Gate: Hanya akun yang sudah terverifikasi yang boleh login
	if user.EmailVerifiedAt == nil {
		helper.Error(w, http.StatusForbidden, "Email belum diverifikasi. Silakan periksa email Anda atau minta kirim ulang tautan verifikasi.")
		return
	}

	token, err := helper.GenerateToken(user.ID)
	if err != nil {
		helper.Error(w, http.StatusInternalServerError, "Gagal men-generate token.")
		return
	}

	helper.JSON(w, http.StatusOK, "Login berhasil", map[string]interface{}{
		"token": token,
		"user": model.UserResponse{
			ID:              user.ID,
			Email:           user.Email,
			Name:            user.Name,
			EmailVerifiedAt: user.EmailVerifiedAt,
			CreatedAt:       user.CreatedAt,
		},
	})
}

func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req model.VerifyEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Token) == "" {
		helper.Error(w, http.StatusBadRequest, "Token verifikasi wajib diisi.")
		return
	}

	tokenHash := helper.HashToken(req.Token)
	vToken, err := h.authRepo.GetVerificationToken(r.Context(), tokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			helper.Error(w, http.StatusBadRequest, "Token verifikasi tidak valid.")
			return
		}
		helper.Error(w, http.StatusInternalServerError, "Gagal memproses verifikasi email.")
		return
	}

	if vToken.UsedAt != nil {
		helper.Error(w, http.StatusBadRequest, "Token verifikasi sudah pernah digunakan.")
		return
	}

	if time.Now().After(vToken.ExpiresAt) {
		helper.Error(w, http.StatusBadRequest, "Token verifikasi sudah kedaluwarsa. Silakan minta kirim ulang tautan verifikasi.")
		return
	}

	if err := h.authRepo.MarkEmailVerified(r.Context(), vToken.UserID, vToken.ID); err != nil {
		helper.Error(w, http.StatusInternalServerError, "Gagal memverifikasi akun.")
		return
	}

	helper.JSON(w, http.StatusOK, "Email berhasil diverifikasi. Silakan login ke akun Anda.", nil)
}

func (h *AuthHandler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	var req model.ResendVerificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || !helper.IsValidEmail(req.Email) {
		helper.Error(w, http.StatusBadRequest, "Format email tidak valid.")
		return
	}

	user, err := h.authRepo.GetUserByEmail(r.Context(), req.Email)
	if err != nil || user == nil {
		helper.JSON(w, http.StatusOK, "Jika email terdaftar dan belum diverifikasi, tautan baru telah dikirim.", nil)
		return
	}

	if user.EmailVerifiedAt != nil {
		helper.Error(w, http.StatusBadRequest, "Email ini sudah terverifikasi. Silakan langsung login.")
		return
	}

	rawToken, err := helper.GenerateSessionToken()
	if err != nil {
		helper.Error(w, http.StatusInternalServerError, "Gagal memproses token verifikasi.")
		return
	}

	tokenHash := helper.HashToken(rawToken)
	vToken := &model.VerificationToken{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(30 * time.Minute),
	}

	if err := h.authRepo.CreateVerificationToken(r.Context(), vToken); err != nil {
		helper.Error(w, http.StatusInternalServerError, "Gagal menyimpan token verifikasi.")
		return
	}

	verificationURL := "http://localhost:5173/verify-email?token=" + rawToken
	log.Printf("\n========================================\n[RESEND EMAIL VERIFIKASI] Kirim ke: %s\nTautan: %s\n(Kedaluwarsa dalam 30 menit)\n========================================\n", user.Email, verificationURL)

	helper.JSON(w, http.StatusOK, "Tautan verifikasi baru berhasil dikirim ke email Anda.", nil)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {

	helper.JSON(w, http.StatusOK, "Logout berhasil.", nil)
}

func (h *AuthHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		helper.JSON(w, http.StatusOK, "Status sesi", map[string]bool{"active": false})
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		helper.JSON(w, http.StatusOK, "Status sesi", map[string]bool{"active": false})
		return
	}

	tokenStr := parts[1]

	_, err := helper.ValidateToken(tokenStr)
	if err != nil {
		helper.JSON(w, http.StatusOK, "Status sesi", map[string]bool{"active": false})
		return
	}

	helper.JSON(w, http.StatusOK, "Status sesi", map[string]bool{"active": true})
}

func (h *AuthHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.ContextUserID).(string)
	if !ok {
		helper.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, err := h.authRepo.GetUserByID(r.Context(), userID)
	if err != nil {
		helper.Error(w, http.StatusNotFound, "Pengguna tidak ditemukan.")
		return
	}

	helper.JSON(w, http.StatusOK, "Profil berhasil dimuat", model.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
	})
}

func (h *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.ContextUserID).(string)
	if !ok {
		helper.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req model.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.Error(w, http.StatusBadRequest, "Format data tidak valid.")
		return
	}

	user, err := h.authRepo.GetUserByID(r.Context(), userID)
	if err != nil {
		helper.Error(w, http.StatusNotFound, "Pengguna tidak ditemukan.")
		return
	}

	if req.Name != nil {
		user.Name = *req.Name
	}

	if err := h.authRepo.UpdateUser(r.Context(), user); err != nil {
		helper.Error(w, http.StatusInternalServerError, "Gagal memperbarui profil.")
		return
	}

	helper.JSON(w, http.StatusOK, "Profil berhasil diperbarui", model.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
	})
}
