package api

import (
	"context"
	"time"

	"github.com/stlpine/will-it-compile/pkg/models"
)

// processJob processes a compilation job asynchronously
// This runs in a goroutine and updates the job status throughout the process.
func (s *Server) processJob(job models.CompilationJob) {
	// Update status to processing
	job.Status = models.StatusProcessing
	now := time.Now()
	job.StartedAt = &now
	s.jobs.Store(job)

	// Compile the code
	result := s.compiler.Compile(context.Background(), job)

	// Update job status based on result
	completed := time.Now()
	job.CompletedAt = &completed

	if result.Error != "" {
		job.Status = models.StatusFailed
	} else {
		job.Status = models.StatusCompleted
	}

	s.jobs.Store(job)

	// Store the compilation result
	s.jobs.StoreResult(job.ID, result)
}
