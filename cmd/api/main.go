package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"

	"project.com/momentum-service/internal/database"
	"project.com/momentum-service/internal/handler"
	"project.com/momentum-service/internal/repository"
	"project.com/momentum-service/internal/router"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Peringatan: File .env tidak ditemukan, membaca dari OS environment")
	}

	dsn := os.Getenv("DATABASE_URL")

	if dsn == "" {
		log.Fatal("DATABASE_URL belum disetel di .env")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal("Gagal membuka koneksi database:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Gagal terhubung ke database:", err)
	}

	if err := database.RunAutoMigration(db); err != nil {
		log.Fatal("Gagal menjalankan migrasi:", err)
	}

	activityRepo := repository.NewActivityRepository(db)

	activityHandler := handler.NewActivityHandler(activityRepo)

	appRouter := router.SetupRoutes(router.RouterConfig{
		ActivityHandler: activityHandler,
	})
	port := ":8080"
	log.Printf("Server sedang berjalan di http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, appRouter))
}
