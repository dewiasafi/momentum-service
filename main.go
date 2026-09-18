package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

type Activity struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	IsCompleted bool   `json:"is_completed"`
	Notes       string `json:"notes"`
}

var db *sql.DB

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Println("Peringatan: File .env tidak ditemukan, membaca dari OS environment")
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL belum disetel di .env")
	}

	db, err = sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal("Gagal membuka koneksi database:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Gagal terhubung ke database:", err)
	}

	http.HandleFunc("/", handler)
	http.HandleFunc("GET /api/v1/activities", getActivitiesHandler)
	http.HandleFunc("POST /api/v1/activities", createActivityHandler)

	port := ":8080"
	log.Printf("Server sedang berjalan di http://localhost%s\n", port)
	errServer := http.ListenAndServe(port, corsMiddleware(http.DefaultServeMux))
	if errServer != nil {
		log.Println("Gagal menjalankan server:", errServer)
	}

}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Momentum API is running")
}

func getActivitiesHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, title, is_completed, COALESCE(notes,'') FROM activities ORDER BY id ASC")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	list := []Activity{}

	for rows.Next() {
		var a Activity
		err := rows.Scan(&a.ID, &a.Title, &a.IsCompleted, &a.Notes)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		list = append(list, a)
	}

	if err = rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(list)
}

func createActivityHandler(w http.ResponseWriter, r *http.Request) {

	var newActivity Activity

	err := json.NewDecoder(r.Body).Decode(&newActivity)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid JSON format",
		})
		return
	}

	if newActivity.Title == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Title cannot be empty",
		})
		return
	}
	query := `INSERT INTO activities (title, notes) VALUES ($1, $2) RETURNING id, is_completed`

	err = db.QueryRow(query, newActivity.Title, newActivity.Notes).Scan(&newActivity.ID, &newActivity.IsCompleted)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Gagal menyimpan ke database"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newActivity)
}
