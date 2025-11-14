package docker

import "context"

// DockerClient is an interface for Docker operations
// This allows for easier mocking in tests.
type DockerClient interface {
	RunCompilation(ctx context.Context, config CompilationConfig) (*CompilationOutput, error)
	ImageExists(ctx context.Context, imageTag string) (bool, error)
	Close() error
}

// Ensure Client implements DockerClient.
var _ DockerClient = (*Client)(nil)
