package models

import "time"

// CompilationJob represents a job to be processed
type CompilationJob struct {
	ID          string             `json:"id"`
	Request     CompilationRequest `json:"request"`
	Status      JobStatus          `json:"status"`
	CreatedAt   time.Time          `json:"created_at"`
	StartedAt   *time.Time         `json:"started_at,omitempty"`
	CompletedAt *time.Time         `json:"completed_at,omitempty"`
}

// JobResponse is returned when a job is created
type JobResponse struct {
	JobID  string    `json:"job_id"`
	Status JobStatus `json:"status"`
}
