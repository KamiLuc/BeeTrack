package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	log.Println("Starting BeeTrack API on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
