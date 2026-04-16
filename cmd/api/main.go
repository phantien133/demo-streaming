package main

import (
	"log"

	"demo-streaming/internal/config"
	"demo-streaming/internal/database"
	"demo-streaming/internal/server"
)

func main() {
	config.LoadDotEnv()
	port := config.Port()
	db, closeDB, err := database.NewGormDB(config.DatabaseURL())
	if err != nil {
		log.Fatalf("failed to connect database with gorm: %v", err)
	}
	defer func() {
		if err := closeDB(); err != nil {
			log.Printf("failed to close database connection: %v", err)
		}
	}()

	router := server.NewRouter(db)
	addr := ":" + port

	log.Printf("api server listening on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("failed to run api server: %v", err)
	}
}
