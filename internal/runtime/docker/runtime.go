package docker

import (
	"context"
	"fmt"

	"github.com/stlpine/will-it-compile/internal/docker"
	"github.com/stlpine/will-it-compile/pkg/runtime"
)

// DockerRuntime implements CompilationRuntime using Docker
// This is used for local development and single-server deployments
type DockerRuntime struct {
	client docker.DockerClient
}

// NewDockerRuntime creates a new Docker-based compilation runtime
func NewDockerRuntime() (*DockerRuntime, error) {
	client, err := docker.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	return &DockerRuntime{
		client: client,
	}, nil
}

// Compile runs compilation using Docker containers
func (d *DockerRuntime) Compile(ctx context.Context, config runtime.CompilationConfig) (*runtime.CompilationOutput, error) {
	// Convert runtime.CompilationConfig to docker.CompilationConfig
	dockerConfig := docker.CompilationConfig{
		ImageTag:   config.ImageTag,
		SourceCode: config.SourceCode,
		WorkDir:    config.WorkDir,
		Env:        config.Env,
	}

	// Apply timeout if specified
	if config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, config.Timeout)
		defer cancel()
	}

	// Run compilation using Docker
	output, err := d.client.RunCompilation(ctx, dockerConfig)
	if err != nil {
		return nil, fmt.Errorf("docker compilation failed: %w", err)
	}

	// Convert docker.CompilationOutput to runtime.CompilationOutput
	return &runtime.CompilationOutput{
		Stdout:   output.Stdout,
		Stderr:   output.Stderr,
		ExitCode: output.ExitCode,
		Duration: output.Duration,
		TimedOut: output.TimedOut,
	}, nil
}

// ImageExists checks if a Docker image exists locally
func (d *DockerRuntime) ImageExists(ctx context.Context, imageTag string) (bool, error) {
	return d.client.ImageExists(ctx, imageTag)
}

// Close cleans up Docker client resources
func (d *DockerRuntime) Close() error {
	return d.client.Close()
}

// Ensure DockerRuntime implements CompilationRuntime
var _ runtime.CompilationRuntime = (*DockerRuntime)(nil)
