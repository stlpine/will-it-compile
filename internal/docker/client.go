package docker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
	"github.com/docker/docker/pkg/stdcopy"
)

const (
	// Resource limits.
	MaxMemory     = 128 * 1024 * 1024 // 128MB
	MaxMemorySwap = 128 * 1024 * 1024 // No swap
	MaxCPUQuota   = 50000             // 0.5 CPU
	MaxPidsLimit  = 100               // Max processes
	MaxOutputSize = 1 * 1024 * 1024   // 1MB output

	// Timeouts.
	MaxCompilationTime = 30 * time.Second
)

// Client wraps the Docker client with secure container operations.
type Client struct {
	cli *client.Client
}

// NewClient creates a new Docker client.
func NewClient() (*Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	return &Client{cli: cli}, nil
}

// Close closes the Docker client.
func (c *Client) Close() error {
	return c.cli.Close()
}

// ImageExists checks if a Docker image exists locally.
func (c *Client) ImageExists(ctx context.Context, imageTag string) (bool, error) {
	_, err := c.cli.ImageInspect(ctx, imageTag)
	if err != nil {
		if errdefs.IsNotFound(err) { //nolint:staticcheck // SA1019: errdefs.IsNotFound is correct for Docker client
			return false, nil
		}
		return false, fmt.Errorf("failed to inspect image: %w", err)
	}
	return true, nil
}

// CompilationConfig holds configuration for a compilation container.
type CompilationConfig struct {
	ImageTag        string
	SourceCode      string
	WorkDir         string
	Env             []string
	SecurityOptPath string // Path to seccomp profile
}

// CompilationOutput holds the output from a compilation.
type CompilationOutput struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Duration time.Duration
	TimedOut bool
}

// RunCompilation creates and runs a secure container for compilation.
func (c *Client) RunCompilation(ctx context.Context, config CompilationConfig) (*CompilationOutput, error) {
	startTime := time.Now()

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(ctx, MaxCompilationTime)
	defer cancel()

	// Create container with security constraints
	containerID, err := c.createSecureContainer(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	// Ensure cleanup - use context without cancel to allow cleanup even if parent context is cancelled
	defer func() {
		cleanupCtx := context.WithoutCancel(ctx)
		_ = c.cli.ContainerRemove(cleanupCtx, containerID, container.RemoveOptions{ //nolint:errcheck // best effort cleanup
			Force:         true,
			RemoveVolumes: true,
		})
	}()

	// Start the container
	if err := c.cli.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	// Wait for container to finish or timeout
	statusCh, errCh := c.cli.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)

	var exitCode int64
	timedOut := false

	select {
	case err := <-errCh:
		if err != nil {
			return nil, fmt.Errorf("error waiting for container: %w", err)
		}
	case status := <-statusCh:
		exitCode = status.StatusCode
	case <-ctx.Done():
		// Timeout occurred - kill the container
		timedOut = true
		killCtx := context.WithoutCancel(ctx)
		_ = c.cli.ContainerKill(killCtx, containerID, "SIGKILL") //nolint:errcheck // best effort kill
	}

	// Collect output - use context without cancel to ensure we can collect output even after timeout
	outputCtx := context.WithoutCancel(ctx)
	stdout, stderr, err := c.collectOutput(outputCtx, containerID)
	if err != nil {
		return nil, fmt.Errorf("failed to collect output: %w", err)
	}

	duration := time.Since(startTime)

	return &CompilationOutput{
		Stdout:   stdout,
		Stderr:   stderr,
		ExitCode: int(exitCode),
		Duration: duration,
		TimedOut: timedOut,
	}, nil
}

// createSecureContainer creates a container with all security constraints.
func (c *Client) createSecureContainer(ctx context.Context, config CompilationConfig) (string, error) {
	// Security options
	securityOpt := []string{
		"no-new-privileges",
	}

	// Add seccomp profile if provided
	if config.SecurityOptPath != "" {
		securityOpt = append(securityOpt, "seccomp="+config.SecurityOptPath)
	}

	// Container configuration
	containerConfig := &container.Config{
		Image:           config.ImageTag,
		Cmd:             []string{"/usr/bin/compile.sh"},
		WorkingDir:      "/workspace",
		User:            "compiler",
		NetworkDisabled: true,
		Env:             config.Env,
	}

	// Host configuration with resource limits and security
	hostConfig := &container.HostConfig{
		Resources: container.Resources{
			Memory:     MaxMemory,
			MemorySwap: MaxMemorySwap,
			CPUQuota:   MaxCPUQuota,
			PidsLimit:  func() *int64 { v := int64(MaxPidsLimit); return &v }(),
		},
		SecurityOpt:    securityOpt,
		ReadonlyRootfs: false,           // Must be false to copy files before start
		CapDrop:        []string{"ALL"}, // Drop all capabilities
		Tmpfs: map[string]string{
			"/tmp": "rw,noexec,nosuid,size=64m",
		},
		// Ensure no mounts from host
		Mounts: []mount.Mount{},
	}

	// Create the container
	resp, err := c.cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		return "", err
	}

	// Copy source code into container
	if err := c.copySourceToContainer(ctx, resp.ID, config.SourceCode); err != nil {
		// Cleanup on error
		_ = c.cli.ContainerRemove(context.Background(), resp.ID, container.RemoveOptions{Force: true}) //nolint:errcheck // already in error path
		return "", fmt.Errorf("failed to copy source code: %w", err)
	}

	return resp.ID, nil
}

// copySourceToContainer copies source code into the container.
func (c *Client) copySourceToContainer(ctx context.Context, containerID, sourceCode string) error {
	// Create a tar archive with the source code
	tarContent, err := createSourceTar(sourceCode, "source.cpp")
	if err != nil {
		return err
	}

	// Copy to container
	return c.cli.CopyToContainer(ctx, containerID, "/workspace", tarContent, container.CopyToContainerOptions{})
}

// collectOutput retrieves stdout and stderr from the container.
func (c *Client) collectOutput(ctx context.Context, containerID string) (string, string, error) {
	logs, err := c.cli.ContainerLogs(ctx, containerID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	})
	if err != nil {
		return "", "", err
	}
	defer logs.Close() //nolint:errcheck // read-only operation

	// Use limited readers to prevent excessive output
	stdoutBuf := &limitedWriter{limit: MaxOutputSize}
	stderrBuf := &limitedWriter{limit: MaxOutputSize}

	// Docker multiplexes stdout/stderr
	if _, err := stdcopy.StdCopy(stdoutBuf, stderrBuf, logs); err != nil && !errors.Is(err, io.EOF) {
		return "", "", err
	}

	return sanitizeOutput(stdoutBuf.String()), sanitizeOutput(stderrBuf.String()), nil
}

// sanitizeOutput removes potentially dangerous content from output.
func sanitizeOutput(output string) string {
	// Remove ANSI escape sequences
	output = removeANSIEscapes(output)

	// Truncate if too long
	if len(output) > MaxOutputSize {
		output = output[:MaxOutputSize] + "\n... (output truncated)"
	}

	return output
}

// removeANSIEscapes removes ANSI escape sequences.
func removeANSIEscapes(s string) string {
	// Simple implementation - in production, use a library like github.com/acarl005/stripansi
	result := strings.Builder{}
	inEscape := false

	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
				inEscape = false
			}
			continue
		}
		result.WriteRune(r)
	}

	return result.String()
}

// limitedWriter wraps a strings.Builder with a size limit.
type limitedWriter struct {
	strings.Builder
	limit int
}

func (w *limitedWriter) Write(p []byte) (n int, err error) {
	remaining := w.limit - w.Len()
	if remaining <= 0 {
		return 0, io.EOF
	}

	if len(p) > remaining {
		p = p[:remaining]
	}

	return w.Builder.Write(p)
}
