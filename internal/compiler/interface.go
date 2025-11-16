package compiler

import (
	"context"

	"github.com/stlpine/will-it-compile/pkg/models"
)

// CompilerInterface defines the interface for compilation services.
// This interface allows for mocking in tests and supports different
// compilation backends (Docker, Kubernetes, etc.).
type CompilerInterface interface {
	// Compile executes a compilation job and returns the result
	Compile(ctx context.Context, job models.CompilationJob) models.CompilationResult

	// GetSupportedEnvironments returns a list of available compilation environments
	GetSupportedEnvironments() []models.Environment

	// Close cleans up compiler resources
	Close() error
}

// Ensure *Compiler implements CompilerInterface
var _ CompilerInterface = (*Compiler)(nil)
