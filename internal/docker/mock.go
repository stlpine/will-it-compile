package docker

import (
	"context"
	"time"
)

// MockDockerClient is a mock implementation of DockerClient for testing
type MockDockerClient struct {
	RunCompilationFunc func(ctx context.Context, config CompilationConfig) (*CompilationOutput, error)
	ImageExistsFunc    func(ctx context.Context, imageTag string) (bool, error)
	CloseFunc          func() error
}

// RunCompilation calls the mock function
func (m *MockDockerClient) RunCompilation(ctx context.Context, config CompilationConfig) (*CompilationOutput, error) {
	if m.RunCompilationFunc != nil {
		return m.RunCompilationFunc(ctx, config)
	}
	// Default behavior: success
	return &CompilationOutput{
		Stdout:   "Compilation successful",
		Stderr:   "",
		ExitCode: 0,
		Duration: time.Second,
		TimedOut: false,
	}, nil
}

// ImageExists calls the mock function
func (m *MockDockerClient) ImageExists(ctx context.Context, imageTag string) (bool, error) {
	if m.ImageExistsFunc != nil {
		return m.ImageExistsFunc(ctx, imageTag)
	}
	// Default behavior: image exists
	return true, nil
}

// Close calls the mock function
func (m *MockDockerClient) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

// Ensure MockDockerClient implements DockerClient
var _ DockerClient = (*MockDockerClient)(nil)
