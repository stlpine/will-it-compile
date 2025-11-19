package memory

import (
	"sync"

	"github.com/stlpine/will-it-compile/pkg/models"
)

// Store provides in-memory storage for jobs.
// ⚠️ WARNING: This is not suitable for production with multiple instances.
// Use Redis storage for production deployments.
type Store struct {
	mu      sync.RWMutex
	jobs    map[string]models.CompilationJob
	results map[string]models.CompilationResult
}

// NewStore creates a new in-memory job store.
func NewStore() *Store {
	return &Store{
		jobs:    make(map[string]models.CompilationJob),
		results: make(map[string]models.CompilationResult),
	}
}

// Store saves or updates a job.
func (s *Store) Store(job models.CompilationJob) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[job.ID] = job
	return nil
}

// Get retrieves a job by ID.
func (s *Store) Get(jobID string) (models.CompilationJob, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	job, exists := s.jobs[jobID]
	return job, exists
}

// StoreResult saves a compilation result.
func (s *Store) StoreResult(jobID string, result models.CompilationResult) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.results[jobID] = result
	return nil
}

// GetResult retrieves a compilation result by job ID.
func (s *Store) GetResult(jobID string) (models.CompilationResult, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result, exists := s.results[jobID]
	return result, exists
}

// Close releases any resources (no-op for memory store).
func (s *Store) Close() error {
	return nil
}
