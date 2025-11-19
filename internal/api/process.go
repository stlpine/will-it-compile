package api

import (
	"context"
	"log"
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

	if err := s.jobs.Store(job); err != nil {
		log.Printf("Failed to update job %s to processing status: %v", job.ID, err)
		// Continue processing despite storage error
	}

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

	if err := s.jobs.Store(job); err != nil {
		log.Printf("Failed to update job %s to final status: %v", job.ID, err)
	}

	// Store the compilation result
	if err := s.jobs.StoreResult(job.ID, result); err != nil {
		log.Printf("Failed to store result for job %s: %v", job.ID, err)
	}
}
