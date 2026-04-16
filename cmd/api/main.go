package main

import (
	"log"

	"demo-streaming/internal/config"
	"demo-streaming/internal/server"
)

func main() {
	config.LoadDotEnv()
	port := config.Port()

	router := server.NewRouter()
	addr := ":" + port

	log.Printf("api server listening on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("failed to run api server: %v", err)
	}
}
