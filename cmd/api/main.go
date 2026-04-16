package main

import (
	"log"
	"os"

	"streaming-learn/internal/server"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router := server.NewRouter()
	addr := ":" + port

	log.Printf("api server listening on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("failed to run api server: %v", err)
	}
}
