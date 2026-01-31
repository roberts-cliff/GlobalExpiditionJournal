package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"globe-expedition-journal/internal/api"
	"globe-expedition-journal/internal/config"
	"globe-expedition-journal/internal/database"
	"globe-expedition-journal/internal/models"
	"globe-expedition-journal/internal/seed"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Printf("Warning: %v", err)
	}

	// Connect to database
	_, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Run migrations
	if err := database.Migrate(models.AllModels()...); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Seed initial data
	if err := seed.Countries(database.GetDB()); err != nil {
		log.Printf("Warning: failed to seed countries: %v", err)
	}

	// Create router with configuration
	routerCfg := api.RouterConfig{
		SessionSecret: cfg.SessionSecret,
		SessionMaxAge: cfg.SessionMaxAge,
		DemoMode:      cfg.DemoMode,
	}
	router := api.NewRouterWithConfig(database.GetDB(), routerCfg)

	// Create server
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Globe Expedition Journal starting on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
