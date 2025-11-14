package runtime

import (
	"context"
	"time"
)

// MockRuntime is a mock implementation of CompilationRuntime for testing.
type MockRuntime struct {
	CompileFunc     func(ctx context.Context, config CompilationConfig) (*CompilationOutput, error)
	ImageExistsFunc func(ctx context.Context, imageTag string) (bool, error)
	CloseFunc       func() error
}

// Compile calls the mock function.
func (m *MockRuntime) Compile(ctx context.Context, config CompilationConfig) (*CompilationOutput, error) {
	if m.CompileFunc != nil {
		return m.CompileFunc(ctx, config)
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

// ImageExists calls the mock function.
func (m *MockRuntime) ImageExists(ctx context.Context, imageTag string) (bool, error) {
	if m.ImageExistsFunc != nil {
		return m.ImageExistsFunc(ctx, imageTag)
	}
	// Default behavior: image exists
	return true, nil
}

// Close calls the mock function.
func (m *MockRuntime) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

// Ensure MockRuntime implements CompilationRuntime.
var _ CompilationRuntime = (*MockRuntime)(nil)
