package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"

	"project.com/momentum-service/internal/handler"
	"project.com/momentum-service/internal/repository"
	"project.com/momentum-service/internal/router"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Peringatan: File .env tidak ditemukan, membaca dari OS environment")
	}

	if os.Getenv("APP_ENV") == "production" &&
		os.Getenv("JWT_SECRET_KEY") == "" {
		log.Fatal("JWT_SECRET_KEY wajib disetting di production")
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL belum disetting")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal("Gagal membuka koneksi database:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Gagal terhubung ke database:", err)
	}

	activityRepo := repository.NewActivityRepository(db)
	activityHandler := handler.NewActivityHandler(activityRepo)

	authRepo := repository.NewAuthRepository(db)
	authHandler := handler.NewAuthHandler(authRepo)

	journalRepo := repository.NewJournalRepository(db)
	journalHandler := handler.NewJournalHandler(journalRepo)

	appRouter := router.SetupRoutes(router.RouterConfig{
		ActivityHandler: activityHandler,
		AuthHandler:     authHandler,
		AuthRepo:        authRepo,
		JournalHandler:  journalHandler,
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := ":" + port
	log.Printf("Server berjalan di port %s\n", port)
	log.Fatal(http.ListenAndServe(addr, appRouter))

}
