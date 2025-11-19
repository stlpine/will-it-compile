package storage

import (
	"github.com/stlpine/will-it-compile/pkg/models"
)

// JobStore defines the interface for job storage implementations.
// This abstraction allows switching between in-memory (development)
// and Redis/database (production) storage without changing business logic.
type JobStore interface {
	// Store saves or updates a job.
	Store(job models.CompilationJob) error

	// Get retrieves a job by ID.
	// Returns the job and true if found, zero value and false if not found.
	Get(jobID string) (models.CompilationJob, bool)

	// StoreResult saves a compilation result.
	StoreResult(jobID string, result models.CompilationResult) error

	// GetResult retrieves a compilation result by job ID.
	// Returns the result and true if found, zero value and false if not found.
	GetResult(jobID string) (models.CompilationResult, bool)

	// Close releases any resources held by the store.
	Close() error
}
