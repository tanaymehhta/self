package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/tanaymehhta/self/backend/internal/api"
	"github.com/tanaymehhta/self/backend/internal/auth"
	"github.com/tanaymehhta/self/backend/internal/database"
	"github.com/tanaymehhta/self/backend/pkg/config"
	"github.com/tanaymehhta/self/backend/pkg/logger"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize logger
	log := logger.New(cfg)
	log.Info("Starting Self Backend API", "version", "1.0.0", "env", cfg.Env)

	// Connect to Supabase database
	db, err := database.NewSupabaseConnection(cfg, log)
	if err != nil {
		log.LogError(err, "Failed to connect to database")
		os.Exit(1)
	}
	defer db.Close()

	// Initialize JWT manager
	jwtManager := auth.NewJWTManager(cfg)

	// Create server
	server := api.NewServer(db, cfg, log, jwtManager)

	// Start server in a goroutine
	go func() {
		if err := server.Listen(cfg.Port); err != nil {
			log.LogError(err, "Server failed to start")
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// Graceful shutdown
	if err := server.Shutdown(); err != nil {
		log.LogError(err, "Server forced to shutdown")
	} else {
		log.Info("Server shutdown complete")
	}
}