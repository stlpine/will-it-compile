package storage

import (
	"fmt"
	"log"

	"github.com/stlpine/will-it-compile/internal/config"
	"github.com/stlpine/will-it-compile/internal/storage/memory"
	"github.com/stlpine/will-it-compile/internal/storage/redis"
)

// NewJobStore creates a job store based on configuration.
// If Redis is enabled, it returns a Redis-backed store.
// Otherwise, it returns an in-memory store.
func NewJobStore(cfg *config.Config) (JobStore, error) {
	if cfg.Redis.Enabled {
		log.Printf("Initializing Redis job store at %s", cfg.Redis.Addr)
		store, err := redis.NewStore(cfg.Redis)
		if err != nil {
			return nil, fmt.Errorf("failed to create Redis store: %w", err)
		}
		log.Printf("Redis job store initialized successfully (TTL: %s)", cfg.Redis.JobTTL)
		return store, nil
	}

	log.Println("Using in-memory job store (not suitable for production)")
	return memory.NewStore(), nil
}
