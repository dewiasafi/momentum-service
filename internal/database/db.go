package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
)

func ConnectDB() (*sql.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL tidak ditemukan di environment")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("gagal membuka koneksi database: %w", err)
	}

	// Verifikasi koneksi
	if err := db.PingContext(context.Background()); err != nil {
		return nil, fmt.Errorf("gagal melakukan ping ke database: %w", err)
	}

	log.Println("Berhasil terhubung ke database PostgreSQL (Neon) menggunakan database/sql")
	return db, nil
}
