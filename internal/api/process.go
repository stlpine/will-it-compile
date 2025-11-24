package api

import (
	"context"
	"log"
	"strings"
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
	// StatusCompleted = code compiled successfully (exit code 0)
	// StatusFailed = code failed to compile (syntax/linker errors)
	// StatusTimeout = compilation timed out
	// StatusError = infrastructure/system error
	completed := time.Now()
	job.CompletedAt = &completed

	job.Status = determineJobStatus(result)

	if err := s.jobs.Store(job); err != nil {
		log.Printf("Failed to update job %s to final status: %v", job.ID, err)
	}

	// Store the compilation result
	if err := s.jobs.StoreResult(job.ID, result); err != nil {
		log.Printf("Failed to store result for job %s: %v", job.ID, err)
	}
}

// determineJobStatus determines the appropriate job status based on compilation result.
// Status meanings:
//   - StatusCompleted: code compiled successfully (exit code 0)
//   - StatusFailed: code failed to compile (syntax/linker errors) - user's fault
//   - StatusTimeout: compilation timed out - could be user's code (infinite template) or system
//   - StatusError: infrastructure/system error - our fault
func determineJobStatus(result models.CompilationResult) models.JobStatus {
	// Check for timeout first (specific error message from compiler)
	if result.Error == "compilation timeout" {
		return models.StatusTimeout
	}

	// Check for infrastructure errors (runtime failures, validation errors)
	if result.Error != "" {
		// These are system/infrastructure errors, not user code errors
		// Examples: "compilation failed: ...", "invalid base64 encoding"
		if strings.HasPrefix(result.Error, "compilation failed:") ||
			strings.Contains(result.Error, "base64") ||
			strings.Contains(result.Error, "docker") ||
			strings.Contains(result.Error, "container") {
			return models.StatusError
		}
		// Any other error is treated as infrastructure error
		return models.StatusError
	}

	// No error - check if code actually compiled
	if result.Compiled {
		return models.StatusCompleted
	}

	// Code didn't compile (syntax errors, linker errors, etc.) - user's code issue
	return models.StatusFailed
}
