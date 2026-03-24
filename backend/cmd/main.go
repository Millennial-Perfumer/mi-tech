// @title Mi-Tech API
// @version 1.0
// @description Backend API for Mi-Tech GST Invoice Manager.
// @BasePath /api

package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"mi-tech/internal/server"
	"mi-tech/internal/telemetry"
)

func main() {
	// 1. Initialize structured JSON logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// 2. Initialize OpenTelemetry tracing
	shutdown, err := telemetry.InitProvider("mi-tech-backend")
	if err != nil {
		slog.Error("Failed to initialize OpenTelemetry", "error", err)
	} else {
		defer func() {
			if err := shutdown(context.Background()); err != nil {
				slog.Error("Failed to shutdown OpenTelemetry", "error", err)
			}
		}()
	}

	srv, err := server.New()
	if err != nil {
		log.Fatalf("Failed to initialize server: %v", err)
	}

	if err := srv.Run(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
