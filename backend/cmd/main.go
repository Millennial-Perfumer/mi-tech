// @title Mi-Tech API
// @version 1.0
// @description Backend API for Mi-Tech GST Invoice Manager.
// @BasePath /api

package main

import (
	"log"

	"mi-tech/internal/server"
)

func main() {
	srv, err := server.New()
	if err != nil {
		log.Fatalf("Failed to initialize server: %v", err)
	}

	if err := srv.Run(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
