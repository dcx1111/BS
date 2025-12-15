package main

import (
	"log"

	"image-manager/internal/config"
	"image-manager/internal/database"
	"image-manager/internal/server"
)

func main() {
	cfg := config.Load()
	db := database.New(cfg)

	srv := server.New(db, cfg)

	if err := srv.Run(); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
