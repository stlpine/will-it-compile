package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/stlpine/will-it-compile/internal/api"
	"github.com/stlpine/will-it-compile/internal/config"
	"github.com/stlpine/will-it-compile/internal/storage"
)

func main() {
	// Load configuration from environment variables
	cfg := loadConfig()

	// Log configuration
	log.Printf("Starting will-it-compile API server")
	log.Printf("Environment: %s", cfg.Server.Environment)
	log.Printf("Port: %d", cfg.Server.Port)
	log.Printf("Redis enabled: %t", cfg.Redis.Enabled)
	if cfg.Redis.Enabled {
		log.Printf("Redis address: %s", cfg.Redis.Addr)
		log.Printf("Job TTL: %s", cfg.Redis.JobTTL)
	}

	// Create job storage
	jobStore, err := storage.NewJobStore(cfg)
	if err != nil {
		log.Fatalf("Failed to create job storage: %v", err)
	}
	defer func() {
		if err := jobStore.Close(); err != nil {
			log.Printf("Error closing job storage: %v", err)
		}
	}()

	// Create server configuration
	serverConfig := api.ServerConfig{
		MaxWorkers: cfg.Workers.MaxWorkers,
		QueueSize:  cfg.Workers.QueueSize,
	}

	// Create API server with storage
	server, err := api.NewServerWithStorage(serverConfig, jobStore)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}
	defer func() {
		if err := server.Close(); err != nil {
			log.Printf("Error closing server: %v", err)
		}
	}()

	// Create Echo instance with rate limiting enabled
	e := api.NewEchoServer(server, true)

	// Start server in goroutine
	go func() {
		addr := ":" + strconv.Itoa(cfg.Server.Port)
		log.Printf("Server listening on %s", addr)
		if err := e.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped")
}

// loadConfig loads configuration from environment variables.
func loadConfig() *config.Config {
	cfg := config.DefaultConfig()

	// Server configuration
	if port := os.Getenv("PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.Server.Port = p
		}
	}

	if env := os.Getenv("ENVIRONMENT"); env != "" {
		cfg.Server.Environment = env
	}

	// Redis configuration
	if enabled := os.Getenv("REDIS_ENABLED"); enabled == "true" {
		cfg.Redis.Enabled = true
	}

	if addr := os.Getenv("REDIS_ADDR"); addr != "" {
		cfg.Redis.Addr = addr
	}

	if password := os.Getenv("REDIS_PASSWORD"); password != "" {
		cfg.Redis.Password = password
	}

	if db := os.Getenv("REDIS_DB"); db != "" {
		if d, err := strconv.Atoi(db); err == nil {
			cfg.Redis.DB = d
		}
	}

	if poolSize := os.Getenv("REDIS_POOL_SIZE"); poolSize != "" {
		if p, err := strconv.Atoi(poolSize); err == nil {
			cfg.Redis.PoolSize = p
		}
	}

	if ttl := os.Getenv("REDIS_JOB_TTL_HOURS"); ttl != "" {
		if hours, err := strconv.Atoi(ttl); err == nil {
			cfg.Redis.JobTTL = time.Duration(hours) * time.Hour
		}
	}

	// Worker configuration
	if maxWorkers := os.Getenv("MAX_WORKERS"); maxWorkers != "" {
		if w, err := strconv.Atoi(maxWorkers); err == nil {
			cfg.Workers.MaxWorkers = w
		}
	}

	if queueSize := os.Getenv("QUEUE_SIZE"); queueSize != "" {
		if q, err := strconv.Atoi(queueSize); err == nil {
			cfg.Workers.QueueSize = q
		}
	}

	return cfg
}
