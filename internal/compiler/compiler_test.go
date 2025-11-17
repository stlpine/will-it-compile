package compiler

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"github.com/stlpine/will-it-compile/pkg/models"
	"github.com/stlpine/will-it-compile/pkg/runtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCompile_Success tests successful compilation.
func TestCompile_Success(t *testing.T) {
	mockRuntime := &runtime.MockRuntime{
		CompileFunc: func(ctx context.Context, config runtime.CompilationConfig) (*runtime.CompilationOutput, error) {
			return &runtime.CompilationOutput{
				Stdout:   "Compilation successful",
				Stderr:   "",
				ExitCode: 0,
				Duration: 2 * time.Second,
				TimedOut: false,
			}, nil
		},
	}

	compiler := NewCompilerWithRuntime(mockRuntime)

	sourceCode := `#include <iostream>
int main() { return 0; }`
	encodedCode := base64.StdEncoding.EncodeToString([]byte(sourceCode))

	job := models.CompilationJob{
		ID: "test-job-1",
		Request: models.CompilationRequest{
			Code:     encodedCode,
			Language: models.LanguageCpp,
			Compiler: models.CompilerGCC9,
			Standard: models.StandardCpp20,
		},
	}

	result := compiler.Compile(context.Background(), job)

	assert.True(t, result.Success, "Expected compilation to succeed")
	assert.True(t, result.Compiled, "Expected code to compile")
	assert.Equal(t, 0, result.ExitCode, "Expected exit code 0")
	assert.Equal(t, "Compilation successful", result.Stdout)
	assert.Empty(t, result.Error)
	assert.Equal(t, "test-job-1", result.JobID)
}

// TestCompile_CompilationError tests compilation failure.
func TestCompile_CompilationError(t *testing.T) {
	mockRuntime := &runtime.MockRuntime{
		CompileFunc: func(ctx context.Context, config runtime.CompilationConfig) (*runtime.CompilationOutput, error) {
			return &runtime.CompilationOutput{
				Stdout:   "",
				Stderr:   "error: expected ';' before 'return'",
				ExitCode: 1,
				Duration: 1 * time.Second,
				TimedOut: false,
			}, nil
		},
	}

	compiler := NewCompilerWithRuntime(mockRuntime)

	sourceCode := `int main() { return 0 }` // missing semicolon
	encodedCode := base64.StdEncoding.EncodeToString([]byte(sourceCode))

	job := models.CompilationJob{
		ID: "test-job-2",
		Request: models.CompilationRequest{
			Code:     encodedCode,
			Language: models.LanguageCpp,
			Compiler: models.CompilerGCC9,
		},
	}

	result := compiler.Compile(context.Background(), job)

	assert.True(t, result.Success, "Job should succeed even if compilation fails")
	assert.False(t, result.Compiled, "Expected code not to compile")
	assert.Equal(t, 1, result.ExitCode, "Expected non-zero exit code")
	assert.Contains(t, result.Stderr, "error", "Expected error in stderr")
}

// TestCompile_Timeout tests compilation timeout.
func TestCompile_Timeout(t *testing.T) {
	mockRuntime := &runtime.MockRuntime{
		CompileFunc: func(ctx context.Context, config runtime.CompilationConfig) (*runtime.CompilationOutput, error) {
			return &runtime.CompilationOutput{
				Stdout:   "",
				Stderr:   "",
				ExitCode: 137, // SIGKILL
				Duration: 30 * time.Second,
				TimedOut: true,
			}, nil
		},
	}

	compiler := NewCompilerWithRuntime(mockRuntime)

	sourceCode := `int main() { while(1); }` // infinite loop
	encodedCode := base64.StdEncoding.EncodeToString([]byte(sourceCode))

	job := models.CompilationJob{
		ID: "test-job-3",
		Request: models.CompilationRequest{
			Code:     encodedCode,
			Language: models.LanguageCpp,
			Compiler: models.CompilerGCC9,
		},
	}

	result := compiler.Compile(context.Background(), job)

	assert.True(t, result.Success, "Job should succeed")
	assert.False(t, result.Compiled, "Expected code not to compile due to timeout")
	assert.Equal(t, "compilation timeout", result.Error)
}

// TestCompile_RuntimeError tests runtime errors.
func TestCompile_RuntimeError(t *testing.T) {
	mockRuntime := &runtime.MockRuntime{
		CompileFunc: func(ctx context.Context, config runtime.CompilationConfig) (*runtime.CompilationOutput, error) {
			return nil, errors.New("failed to create container") //nolint:err113 // mock error for testing
		},
	}

	compiler := NewCompilerWithRuntime(mockRuntime)

	sourceCode := `int main() { return 0; }`
	encodedCode := base64.StdEncoding.EncodeToString([]byte(sourceCode))

	job := models.CompilationJob{
		ID: "test-job-4",
		Request: models.CompilationRequest{
			Code:     encodedCode,
			Language: models.LanguageCpp,
			Compiler: models.CompilerGCC9,
		},
	}

	result := compiler.Compile(context.Background(), job)

	assert.False(t, result.Success, "Expected compilation to fail")
	assert.False(t, result.Compiled, "Expected code not to compile")
	assert.Contains(t, result.Error, "failed to create container")
}

// TestValidateRequest tests request validation using table-driven tests.
func TestValidateRequest(t *testing.T) {
	compiler := NewCompilerWithRuntime(&runtime.MockRuntime{})

	testCases := []struct {
		name        string
		request     models.CompilationRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid_request",
			request: models.CompilationRequest{
				Code:     base64.StdEncoding.EncodeToString([]byte("int main() {}")),
				Language: models.LanguageCpp,
				Compiler: models.CompilerGCC9,
			},
			expectError: false,
		},
		{
			name: "missing_code",
			request: models.CompilationRequest{
				Language: models.LanguageCpp,
				Compiler: models.CompilerGCC9,
			},
			expectError: true,
			errorMsg:    "source code is required",
		},
		{
			name: "unsupported_language",
			request: models.CompilationRequest{
				Code:     base64.StdEncoding.EncodeToString([]byte("package main")),
				Language: models.LanguageGo,
			},
			expectError: true,
			errorMsg:    "unsupported language",
		},
		{
			name: "code_too_large",
			request: models.CompilationRequest{
				Code:     base64.StdEncoding.EncodeToString(make([]byte, 2*1024*1024)),
				Language: models.LanguageCpp,
			},
			expectError: true,
			errorMsg:    "too large",
		},
		{
			name: "alternative_cpp_syntax",
			request: models.CompilationRequest{
				Code:     base64.StdEncoding.EncodeToString([]byte("int main() {}")),
				Language: models.LanguageCPP,
				Compiler: models.CompilerGCC9,
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := compiler.validateRequest(tc.request)

			if tc.expectError {
				require.Error(t, err, "Expected validation error")
				assert.Contains(t, err.Error(), tc.errorMsg)
			} else {
				assert.NoError(t, err, "Expected no validation error")
			}
		})
	}
}

// TestSelectEnvironment tests environment selection.
func TestSelectEnvironment(t *testing.T) {
	compiler := NewCompilerWithRuntime(&runtime.MockRuntime{})

	testCases := []struct {
		name             string
		request          models.CompilationRequest
		expectError      bool
		expectedImageTag string
		expectedStandard string
	}{
		{
			name: "default_cpp_environment",
			request: models.CompilationRequest{
				Language: models.LanguageCpp,
				Compiler: models.CompilerGCC9,
			},
			expectError:      false,
			expectedImageTag: "gcc:13",
			expectedStandard: string(models.StandardCpp20),
		},
		{
			name: "cpp_with_custom_standard",
			request: models.CompilationRequest{
				Language: models.LanguageCpp,
				Compiler: models.CompilerGCC9,
				Standard: models.StandardCpp17,
			},
			expectError:      false,
			expectedImageTag: "gcc:13",
			expectedStandard: string(models.StandardCpp17),
		},
		{
			name: "alternative_cpp_syntax",
			request: models.CompilationRequest{
				Language: models.LanguageCPP,
				Compiler: models.CompilerGCC9,
			},
			expectError:      false,
			expectedImageTag: "gcc:13",
		},
		{
			name: "unsupported_compiler",
			request: models.CompilationRequest{
				Language: models.LanguageCpp,
				Compiler: models.CompilerClang15,
			},
			expectError: true,
		},
		{
			name: "missing_compiler_uses_default",
			request: models.CompilationRequest{
				Language: models.LanguageCpp,
			},
			expectError:      false,
			expectedImageTag: "gcc:13",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			env, err := compiler.selectEnvironment(tc.request)

			if tc.expectError {
				require.Error(t, err, "Expected environment selection error")
			} else {
				require.NoError(t, err, "Expected no environment selection error")
				assert.Equal(t, tc.expectedImageTag, env.ImageTag)
				if tc.expectedStandard != "" {
					assert.Equal(t, tc.expectedStandard, string(env.Standard))
				}
			}
		})
	}
}

// TestGetSupportedEnvironments tests the environments list.
func TestGetSupportedEnvironments(t *testing.T) {
	compiler := NewCompilerWithRuntime(&runtime.MockRuntime{})

	environments := compiler.GetSupportedEnvironments()

	require.NotEmpty(t, environments, "Expected at least one environment")

	// Check for C++ environment
	found := false
	for _, env := range environments {
		if env.Language == "cpp" {
			found = true
			assert.Contains(t, env.Compilers, "gcc-13")
			assert.Contains(t, env.Standards, "c++20")
			assert.Contains(t, env.OSes, "linux")
			assert.Contains(t, env.Arches, "x86_64")
			break
		}
	}

	assert.True(t, found, "Expected C++ environment in list")
}

// TestCompile_InvalidBase64 tests invalid base64 encoding.
func TestCompile_InvalidBase64(t *testing.T) {
	mockRuntime := &runtime.MockRuntime{}
	compiler := NewCompilerWithRuntime(mockRuntime)

	job := models.CompilationJob{
		ID: "test-job-5",
		Request: models.CompilationRequest{
			Code:     "not-valid-base64!@#$",
			Language: models.LanguageCpp,
			Compiler: models.CompilerGCC9,
		},
	}

	result := compiler.Compile(context.Background(), job)

	assert.False(t, result.Success, "Expected compilation to fail")
	assert.Contains(t, result.Error, "invalid base64")
}

// TestCompile_VerifyRuntimeConfig tests that correct config is passed to runtime.
func TestCompile_VerifyRuntimeConfig(t *testing.T) {
	var capturedConfig runtime.CompilationConfig

	mockRuntime := &runtime.MockRuntime{
		CompileFunc: func(ctx context.Context, config runtime.CompilationConfig) (*runtime.CompilationOutput, error) {
			capturedConfig = config
			return &runtime.CompilationOutput{
				Stdout:   "OK",
				ExitCode: 0,
				Duration: time.Second,
			}, nil
		},
	}

	compiler := NewCompilerWithRuntime(mockRuntime)

	sourceCode := `int main() { return 0; }`
	encodedCode := base64.StdEncoding.EncodeToString([]byte(sourceCode))

	job := models.CompilationJob{
		ID: "test-job-6",
		Request: models.CompilationRequest{
			Code:     encodedCode,
			Language: models.LanguageCpp,
			Compiler: models.CompilerGCC9,
			Standard: models.StandardCpp17,
		},
	}

	result := compiler.Compile(context.Background(), job)

	assert.True(t, result.Success)
	assert.Equal(t, "gcc:13", capturedConfig.ImageTag)
	assert.Equal(t, sourceCode, capturedConfig.SourceCode)
	assert.Equal(t, "/workspace", capturedConfig.WorkDir)
	assert.Contains(t, capturedConfig.Env, "CPP_STANDARD=c++17")
	assert.Equal(t, job.ID, capturedConfig.JobID)
	assert.Equal(t, 30*time.Second, capturedConfig.Timeout)
}
