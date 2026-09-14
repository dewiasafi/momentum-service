package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Activity struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	IsCompleted bool   `json:"is_completed"`
	Notes       string `json:"notes"`
}

func main() {

	http.HandleFunc("/", handler)
	http.HandleFunc("/api/v1/activities", getActivitiesHandler)

	port := ":8080"
	log.Printf("Server sedang berjalan di http://localhost%s\n", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Println("Gagal menjalankan server:", err)
	}

}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Momentum API is running")
}

func getActivitiesHandler(w http.ResponseWriter, r *http.Request) {
	activities := []Activity{
		{ID: "1", Title: "Belajar Go Fundamental", IsCompleted: true, Notes: "Install pakai go get ..."},
		{ID: "2", Title: "Setup Routing REST API", IsCompleted: false},
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(activities); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}
