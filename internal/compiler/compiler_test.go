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
			Compiler: models.CompilerGCC13,
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
			Compiler: models.CompilerGCC13,
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
			Compiler: models.CompilerGCC13,
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
			Compiler: models.CompilerGCC13,
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
				Compiler: models.CompilerGCC13,
			},
			expectError: false,
		},
		{
			name: "missing_code",
			request: models.CompilationRequest{
				Language: models.LanguageCpp,
				Compiler: models.CompilerGCC13,
			},
			expectError: true,
			errorMsg:    "source code is required",
		},
		{
			name: "supported_go_language",
			request: models.CompilationRequest{
				Code:     base64.StdEncoding.EncodeToString([]byte("package main")),
				Language: models.LanguageGo,
				Compiler: models.CompilerGo123,
			},
			expectError: false,
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
				Compiler: models.CompilerGCC13,
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
				Compiler: models.CompilerGCC13,
			},
			expectError:      false,
			expectedImageTag: "gcc:13",
			expectedStandard: string(models.StandardCpp20),
		},
		{
			name: "cpp_with_custom_standard",
			request: models.CompilationRequest{
				Language: models.LanguageCpp,
				Compiler: models.CompilerGCC13,
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
				Compiler: models.CompilerGCC13,
			},
			expectError:      false,
			expectedImageTag: "gcc:13",
		},
		{
			name: "unsupported_compiler",
			request: models.CompilationRequest{
				Language: models.LanguageCpp,
				Compiler: "clang-15", // Not in hardcoded environments
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
			Compiler: models.CompilerGCC13,
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
			Compiler: models.CompilerGCC13,
			Standard: models.StandardCpp17,
		},
	}

	result := compiler.Compile(context.Background(), job)

	assert.True(t, result.Success)
	assert.Equal(t, "gcc:13", capturedConfig.ImageTag)
	assert.Equal(t, sourceCode, capturedConfig.SourceCode)
	assert.Equal(t, "source.cpp", capturedConfig.SourceFilename)
	assert.Equal(t, "/workspace", capturedConfig.WorkDir)
	assert.Contains(t, capturedConfig.Env, "STANDARD=c++17")
	assert.Contains(t, capturedConfig.Env, "SOURCE_FILE=/workspace/source.cpp")
	assert.Equal(t, job.ID, capturedConfig.JobID)
	assert.Equal(t, 30*time.Second, capturedConfig.Timeout)
}

// TestCompile_CLanguage tests C language compilation.
func TestCompile_CLanguage(t *testing.T) {
	var capturedConfig runtime.CompilationConfig

	mockRuntime := &runtime.MockRuntime{
		CompileFunc: func(ctx context.Context, config runtime.CompilationConfig) (*runtime.CompilationOutput, error) {
			capturedConfig = config
			return &runtime.CompilationOutput{
				Stdout:   "C compilation successful",
				ExitCode: 0,
				Duration: time.Second,
			}, nil
		},
	}

	compiler := NewCompilerWithRuntime(mockRuntime)

	sourceCode := `#include <stdio.h>
int main() { printf("Hello, C!\\n"); return 0; }`
	encodedCode := base64.StdEncoding.EncodeToString([]byte(sourceCode))

	job := models.CompilationJob{
		ID: "test-c-job",
		Request: models.CompilationRequest{
			Code:     encodedCode,
			Language: models.LanguageC,
			Compiler: models.CompilerGCC13,
			Standard: models.StandardC17,
		},
	}

	result := compiler.Compile(context.Background(), job)

	assert.True(t, result.Success, "Expected C compilation to succeed")
	assert.True(t, result.Compiled, "Expected C code to compile")
	assert.Equal(t, 0, result.ExitCode)

	// Verify C-specific configuration
	assert.Equal(t, "source.c", capturedConfig.SourceFilename, "Expected C source filename")
	assert.Contains(t, capturedConfig.Env, "STANDARD=c17", "Expected C17 standard")
	assert.Contains(t, capturedConfig.Env, "SOURCE_FILE=/workspace/source.c", "Expected C source file path")
	assert.Contains(t, capturedConfig.CompileCommand, "source.c", "Expected compile command to use source.c")
	assert.Contains(t, capturedConfig.CompileCommand, "c17", "Expected compile command to use c17 standard")
}

// TestCompile_GoLanguage tests Go language compilation.
func TestCompile_GoLanguage(t *testing.T) {
	var capturedConfig runtime.CompilationConfig

	mockRuntime := &runtime.MockRuntime{
		CompileFunc: func(ctx context.Context, config runtime.CompilationConfig) (*runtime.CompilationOutput, error) {
			capturedConfig = config
			return &runtime.CompilationOutput{
				Stdout:   "Go compilation successful",
				ExitCode: 0,
				Duration: time.Second,
			}, nil
		},
	}

	compiler := NewCompilerWithRuntime(mockRuntime)

	sourceCode := `package main
import "fmt"
func main() { fmt.Println("Hello, Go!") }`
	encodedCode := base64.StdEncoding.EncodeToString([]byte(sourceCode))

	job := models.CompilationJob{
		ID: "test-go-job",
		Request: models.CompilationRequest{
			Code:     encodedCode,
			Language: models.LanguageGo,
			Compiler: models.CompilerGo123,
		},
	}

	result := compiler.Compile(context.Background(), job)

	assert.True(t, result.Success, "Expected Go compilation to succeed")
	assert.True(t, result.Compiled, "Expected Go code to compile")
	assert.Equal(t, 0, result.ExitCode)

	// Verify Go-specific configuration
	assert.Equal(t, "main.go", capturedConfig.SourceFilename, "Expected Go source filename")
	assert.Contains(t, capturedConfig.Env, "SOURCE_FILE=/workspace/main.go", "Expected Go source file path")
	assert.Contains(t, capturedConfig.CompileCommand, "go build", "Expected Go build command")
	assert.Contains(t, capturedConfig.CompileCommand, "main.go", "Expected compile command to use main.go")
}

// TestCompile_RustLanguage tests Rust language compilation.
func TestCompile_RustLanguage(t *testing.T) {
	var capturedConfig runtime.CompilationConfig

	mockRuntime := &runtime.MockRuntime{
		CompileFunc: func(ctx context.Context, config runtime.CompilationConfig) (*runtime.CompilationOutput, error) {
			capturedConfig = config
			return &runtime.CompilationOutput{
				Stdout:   "Rust compilation successful",
				ExitCode: 0,
				Duration: time.Second,
			}, nil
		},
	}

	compiler := NewCompilerWithRuntime(mockRuntime)

	sourceCode := `fn main() { println!("Hello, Rust!"); }`
	encodedCode := base64.StdEncoding.EncodeToString([]byte(sourceCode))

	job := models.CompilationJob{
		ID: "test-rust-job",
		Request: models.CompilationRequest{
			Code:     encodedCode,
			Language: models.LanguageRust,
			Compiler: models.CompilerRustc180,
		},
	}

	result := compiler.Compile(context.Background(), job)

	assert.True(t, result.Success, "Expected Rust compilation to succeed")
	assert.True(t, result.Compiled, "Expected Rust code to compile")
	assert.Equal(t, 0, result.ExitCode)

	// Verify Rust-specific configuration
	assert.Equal(t, "main.rs", capturedConfig.SourceFilename, "Expected Rust source filename")
	assert.Contains(t, capturedConfig.Env, "SOURCE_FILE=/workspace/main.rs", "Expected Rust source file path")
	assert.Contains(t, capturedConfig.CompileCommand, "rustc", "Expected rustc command")
	assert.Contains(t, capturedConfig.CompileCommand, "main.rs", "Expected compile command to use main.rs")
}

// TestGetSourceFilename tests the source filename mapping for all languages.
func TestGetSourceFilename(t *testing.T) {
	compiler := NewCompilerWithRuntime(&runtime.MockRuntime{})

	testCases := []struct {
		language         models.Language
		expectedFilename string
	}{
		{models.LanguageC, "source.c"},
		{models.LanguageCpp, "source.cpp"},
		{models.LanguageCPP, "source.cpp"}, // Alternative syntax
		{models.LanguageGo, "main.go"},
		{models.LanguageRust, "main.rs"},
	}

	for _, tc := range testCases {
		t.Run(string(tc.language), func(t *testing.T) {
			filename := compiler.getSourceFilename(tc.language)
			assert.Equal(t, tc.expectedFilename, filename,
				"Expected %s to map to %s", tc.language, tc.expectedFilename)
		})
	}
}

// TestBuildCompileCommand tests compile command generation for all languages.
func TestBuildCompileCommand(t *testing.T) {
	compiler := NewCompilerWithRuntime(&runtime.MockRuntime{})

	testCases := []struct {
		name            string
		envSpec         models.EnvironmentSpec
		sourceFilename  string
		expectedCommand string
		shouldContain   []string
	}{
		{
			name: "cpp_with_standard",
			envSpec: models.EnvironmentSpec{
				Language: models.LanguageCpp,
				Standard: models.StandardCpp20,
			},
			sourceFilename:  "source.cpp",
			expectedCommand: "g++ -std=c++20 /workspace/source.cpp -o /workspace/output",
			shouldContain:   []string{"g++", "-std=c++20", "source.cpp"},
		},
		{
			name: "c_with_standard",
			envSpec: models.EnvironmentSpec{
				Language: models.LanguageC,
				Standard: models.StandardC11,
			},
			sourceFilename:  "source.c",
			expectedCommand: "g++ -std=c11 /workspace/source.c -o /workspace/output",
			shouldContain:   []string{"g++", "-std=c11", "source.c"},
		},
		{
			name: "go_language",
			envSpec: models.EnvironmentSpec{
				Language: models.LanguageGo,
			},
			sourceFilename:  "main.go",
			expectedCommand: "go build -o /workspace/output /workspace/main.go",
			shouldContain:   []string{"go build", "main.go"},
		},
		{
			name: "rust_language",
			envSpec: models.EnvironmentSpec{
				Language: models.LanguageRust,
			},
			sourceFilename:  "main.rs",
			expectedCommand: "rustc /workspace/main.rs -o /workspace/output",
			shouldContain:   []string{"rustc", "main.rs"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			command := compiler.buildCompileCommand(tc.envSpec, tc.sourceFilename)
			assert.Equal(t, tc.expectedCommand, command, "Compile command mismatch")

			for _, substr := range tc.shouldContain {
				assert.Contains(t, command, substr,
					"Expected compile command to contain '%s'", substr)
			}
		})
	}
}

// TestMultiLanguageCompilation tests compilation with different languages in sequence.
func TestMultiLanguageCompilation(t *testing.T) {
	testCases := []struct {
		name           string
		language       models.Language
		compiler       models.Compiler
		standard       models.Standard
		sourceCode     string
		expectedFile   string
		expectedStdEnv string
	}{
		{
			name:           "cpp_compilation",
			language:       models.LanguageCpp,
			compiler:       models.CompilerGCC13,
			standard:       models.StandardCpp17,
			sourceCode:     `int main() { return 0; }`,
			expectedFile:   "source.cpp",
			expectedStdEnv: "STANDARD=c++17",
		},
		{
			name:           "c_compilation",
			language:       models.LanguageC,
			compiler:       models.CompilerGCC13,
			standard:       models.StandardC99,
			sourceCode:     `int main() { return 0; }`,
			expectedFile:   "source.c",
			expectedStdEnv: "STANDARD=c99",
		},
		{
			name:         "go_compilation",
			language:     models.LanguageGo,
			compiler:     models.CompilerGo123,
			sourceCode:   `package main; func main() {}`,
			expectedFile: "main.go",
		},
		{
			name:         "rust_compilation",
			language:     models.LanguageRust,
			compiler:     models.CompilerRustc180,
			sourceCode:   `fn main() {}`,
			expectedFile: "main.rs",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var capturedConfig runtime.CompilationConfig

			mockRuntime := &runtime.MockRuntime{
				CompileFunc: func(ctx context.Context, config runtime.CompilationConfig) (*runtime.CompilationOutput, error) {
					capturedConfig = config
					return &runtime.CompilationOutput{
						Stdout:   "Success",
						ExitCode: 0,
						Duration: time.Second,
					}, nil
				},
			}

			compiler := NewCompilerWithRuntime(mockRuntime)
			encodedCode := base64.StdEncoding.EncodeToString([]byte(tc.sourceCode))

			job := models.CompilationJob{
				ID: "test-multi-" + tc.name,
				Request: models.CompilationRequest{
					Code:     encodedCode,
					Language: tc.language,
					Compiler: tc.compiler,
					Standard: tc.standard,
				},
			}

			result := compiler.Compile(context.Background(), job)

			assert.True(t, result.Success, "Expected %s compilation to succeed", tc.language)
			assert.Equal(t, tc.expectedFile, capturedConfig.SourceFilename,
				"Expected %s to use %s", tc.language, tc.expectedFile)

			if tc.expectedStdEnv != "" {
				assert.Contains(t, capturedConfig.Env, tc.expectedStdEnv,
					"Expected %s environment variable", tc.expectedStdEnv)
			}
		})
	}
}

// TestLanguageValidation tests that all supported languages pass validation.
func TestLanguageValidation(t *testing.T) {
	compiler := NewCompilerWithRuntime(&runtime.MockRuntime{})

	// These should now all pass validation (no longer MVP-restricted)
	supportedLanguages := []models.Language{
		models.LanguageC,
		models.LanguageCpp,
		models.LanguageCPP, // Alternative syntax
		models.LanguageGo,
		models.LanguageRust,
	}

	for _, lang := range supportedLanguages {
		t.Run(string(lang), func(t *testing.T) {
			req := models.CompilationRequest{
				Code:     base64.StdEncoding.EncodeToString([]byte("test code")),
				Language: lang,
			}

			err := compiler.validateRequest(req)
			assert.NoError(t, err, "Expected %s to be supported", lang)
		})
	}
}

// TestCompileCommand_EdgeCases tests edge cases in compile command generation.
func TestCompileCommand_EdgeCases(t *testing.T) {
	compiler := NewCompilerWithRuntime(&runtime.MockRuntime{})

	testCases := []struct {
		name           string
		envSpec        models.EnvironmentSpec
		sourceFilename string
		shouldNotPanic bool
	}{
		{
			name: "empty_standard",
			envSpec: models.EnvironmentSpec{
				Language: models.LanguageCpp,
				Standard: "",
			},
			sourceFilename: "source.cpp",
			shouldNotPanic: true,
		},
		{
			name: "unknown_language_fallback",
			envSpec: models.EnvironmentSpec{
				Language: models.Language("unknown"),
				Standard: models.StandardCpp20,
			},
			sourceFilename: "source.cpp",
			shouldNotPanic: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				command := compiler.buildCompileCommand(tc.envSpec, tc.sourceFilename)
				assert.NotEmpty(t, command, "Expected non-empty compile command")
			}, "buildCompileCommand should not panic")
		})
	}
}
