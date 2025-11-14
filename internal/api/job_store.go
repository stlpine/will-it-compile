package api

import (
	"sync"

	"github.com/stlpine/will-it-compile/pkg/models"
)

// jobStore provides in-memory storage for jobs (MVP only)
// ⚠️ WARNING: This is not suitable for production with multiple instances.
// Replace with Redis or database in Phase 3.
type jobStore struct {
	mu      sync.RWMutex
	jobs    map[string]models.CompilationJob
	results map[string]models.CompilationResult
}

// newJobStore creates a new in-memory job store.
func newJobStore() *jobStore {
	return &jobStore{
		jobs:    make(map[string]models.CompilationJob),
		results: make(map[string]models.CompilationResult),
	}
}

// Store saves or updates a job.
func (s *jobStore) Store(job models.CompilationJob) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[job.ID] = job
}

// Get retrieves a job by ID.
func (s *jobStore) Get(jobID string) (models.CompilationJob, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	job, exists := s.jobs[jobID]
	return job, exists
}

// StoreResult saves a compilation result.
func (s *jobStore) StoreResult(jobID string, result models.CompilationResult) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.results[jobID] = result
}

// GetResult retrieves a compilation result by job ID.
func (s *jobStore) GetResult(jobID string) (models.CompilationResult, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result, exists := s.results[jobID]
	return result, exists
}
