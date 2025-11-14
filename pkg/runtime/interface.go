package runtime

import (
	"context"
	"time"
)

// CompilationRuntime abstracts the execution environment for code compilation
// This allows the same code to work with Docker (local dev) or Kubernetes (production).
type CompilationRuntime interface {
	// Compile runs code compilation in an isolated environment
	Compile(ctx context.Context, config CompilationConfig) (*CompilationOutput, error)

	// ImageExists checks if the required compilation image is available
	ImageExists(ctx context.Context, imageTag string) (bool, error)

	// Close cleans up any resources held by the runtime
	Close() error
}

// CompilationConfig holds configuration for a compilation job.
type CompilationConfig struct {
	// JobID is a unique identifier for this compilation
	JobID string

	// ImageTag is the container image to use (e.g., "will-it-compile/cpp-gcc:13-alpine")
	ImageTag string

	// SourceCode is the actual source code to compile
	SourceCode string

	// Env is a list of environment variables in "KEY=VALUE" format
	Env []string

	// WorkDir is the working directory inside the container
	WorkDir string

	// Timeout is the maximum time allowed for compilation
	Timeout time.Duration
}

// CompilationOutput holds the result of a compilation.
type CompilationOutput struct {
	// Stdout is the standard output from the compilation
	Stdout string

	// Stderr is the standard error from the compilation
	Stderr string

	// ExitCode is the exit code of the compilation process
	ExitCode int

	// Duration is how long the compilation took
	Duration time.Duration

	// TimedOut indicates if the compilation exceeded the timeout
	TimedOut bool
}
