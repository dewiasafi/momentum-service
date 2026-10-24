package model

import (
	"errors"
	"strings"
	"time"
)

type JournalCategory struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Slug   string `json:"slug"`
	Source string `json:"source"`
}

type Journal struct {
	ID         string            `json:"id"`
	UserID     string            `json:"-"`
	Title      string            `json:"title"`
	Content    string            `json:"content,omitempty"`
	Categories []JournalCategory `json:"categories"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
}

type JournalVersion struct {
	ID            string    `json:"id"`
	JournalID     string    `json:"journal_id"`
	VersionNumber int       `json:"version_number"`
	Title         string    `json:"title"`
	Content       string    `json:"content,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

type Category struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Slug     string `json:"slug"`
	IsSystem bool   `json:"is_system"`
}

func (j *Journal) Validate() error {
	j.Title = strings.TrimSpace(j.Title)

	if j.Title == "" {
		return errors.New("judul jurnal wajib diisi")
	}

	if len([]rune(j.Title)) > 200 {
		return errors.New("judul jurnal maksimal 200 karakter")
	}

	return nil
}
