package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {

	http.HandleFunc("/", handler)
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
