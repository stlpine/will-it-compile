package config

import (
	"time"
)

// Config holds the application configuration.
type Config struct {
	// Server configuration
	Server ServerConfig

	// Redis configuration
	Redis RedisConfig

	// Worker pool configuration
	Workers WorkerPoolConfig

	// Compilation configuration
	Compilation CompilationConfig
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Port        int
	Environment string // "development" or "production"
}

// RedisConfig holds Redis connection settings.
type RedisConfig struct {
	// Enabled determines whether to use Redis or in-memory storage
	Enabled bool

	// Addr is the Redis server address (host:port)
	Addr string

	// Password for Redis authentication (optional)
	Password string

	// DB is the Redis database number (0-15)
	DB int

	// PoolSize is the maximum number of socket connections
	PoolSize int

	// MaxRetries is the maximum number of retries for failed operations
	MaxRetries int

	// ReadTimeout for socket reads
	ReadTimeout time.Duration

	// WriteTimeout for socket writes
	WriteTimeout time.Duration

	// JobTTL is the time-to-live for job data
	JobTTL time.Duration
}

// WorkerPoolConfig holds worker pool settings.
type WorkerPoolConfig struct {
	// MaxWorkers is the number of concurrent workers
	MaxWorkers int

	// QueueSize is the size of the job queue buffer
	QueueSize int
}

// CompilationConfig holds compilation-specific settings.
type CompilationConfig struct {
	// MaxSourceSize is the maximum allowed source code size in bytes
	MaxSourceSize int

	// Timeout is the maximum compilation time
	Timeout time.Duration
}

// DefaultConfig returns a configuration with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:        8080,
			Environment: "development",
		},
		Redis: RedisConfig{
			Enabled:      false,
			Addr:         "localhost:6379",
			Password:     "",
			DB:           0,
			PoolSize:     20,
			MaxRetries:   3,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
			JobTTL:       24 * time.Hour,
		},
		Workers: WorkerPoolConfig{
			MaxWorkers: 5,
			QueueSize:  100,
		},
		Compilation: CompilationConfig{
			MaxSourceSize: 1 * 1024 * 1024, // 1MB
			Timeout:       30 * time.Second,
		},
	}
}
