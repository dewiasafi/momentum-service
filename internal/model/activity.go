package model

import (
	"errors"
	"strings"
	"time"
)

type Activity struct {
	ID               string     `json:"id"`
	Title            string     `json:"title"`
	Notes            string     `json:"notes,omitempty"`
	IsCompleted      bool       `json:"is_completed"`
	Category         string     `json:"category"`
	Priority         string     `json:"priority"`
	EstimatedMinutes int        `json:"estimated_minutes"`
	CreatedAt        time.Time  `json:"created_at"`
	CompletedAt      *time.Time `json:"completed_at"`
}

func (a *Activity) Validate() error {
	if strings.TrimSpace(a.Title) == "" {
		return errors.New("judul aktivitas tidak boleh kosong")
	}
	if a.IsCompleted && strings.TrimSpace(a.Notes) == "" {
		return errors.New("catatan (notes) wajib diisi saat menyelesaikan aktivitas")
	}

	if a.Category == "" {
		a.Category = "General"
	}

	if a.Priority == "" {
		a.Priority = "medium"
	}
	return nil
}
