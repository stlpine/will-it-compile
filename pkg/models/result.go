package models

import "time"

// CompilationResult represents the result of a compilation
type CompilationResult struct {
	JobID    string        `json:"job_id"`
	Success  bool          `json:"success"`
	Compiled bool          `json:"compiled"` // Whether it compiled successfully
	Stdout   string        `json:"stdout"`
	Stderr   string        `json:"stderr"`
	ExitCode int           `json:"exit_code"`
	Duration time.Duration `json:"duration"`
	Error    string        `json:"error,omitempty"`
}
