package integration

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stlpine/will-it-compile/internal/api"
	"github.com/stlpine/will-it-compile/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCompilationScenarios uses table-driven tests to cover multiple scenarios
func TestCompilationScenarios(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testCases := []struct {
		name           string
		sourceCode     string
		language       models.Language
		compiler       models.Compiler
		standard       models.Standard
		expectCompiled bool
		expectSuccess  bool
		expectedExit   int
		checkStderr    bool // true if we expect stderr to have content
	}{
		{
			name: "valid_hello_world",
			sourceCode: `#include <iostream>
int main() {
    std::cout << "Hello, World!" << std::endl;
    return 0;
}`,
			language:       models.LanguageCpp,
			compiler:       models.CompilerGCC13,
			standard:       models.StandardCpp20,
			expectCompiled: true,
			expectSuccess:  true,
			expectedExit:   0,
			checkStderr:    false,
		},
		{
			name: "missing_semicolon",
			sourceCode: `#include <iostream>
int main() {
    std::cout << "Missing semicolon"
    return 0;
}`,
			language:       models.LanguageCpp,
			compiler:       models.CompilerGCC13,
			standard:       models.StandardCpp20,
			expectCompiled: false,
			expectSuccess:  true, // Success means the job completed, not that code compiled
			expectedExit:   1,
			checkStderr:    true,
		},
		{
			name: "missing_return",
			sourceCode: `#include <iostream>
int main() {
    std::cout << "Missing return" << std::endl;
}`,
			language:       models.LanguageCpp,
			compiler:       models.CompilerGCC13,
			standard:       models.StandardCpp20,
			expectCompiled: true, // GCC allows missing return in main()
			expectSuccess:  true,
			expectedExit:   0,
			checkStderr:    false,
		},
		{
			name: "undefined_function",
			sourceCode: `#include <iostream>
int main() {
    undefinedFunction();
    return 0;
}`,
			language:       models.LanguageCpp,
			compiler:       models.CompilerGCC13,
			standard:       models.StandardCpp20,
			expectCompiled: false,
			expectSuccess:  true,
			expectedExit:   1,
			checkStderr:    true,
		},
		{
			name: "type_mismatch",
			sourceCode: `#include <iostream>
int main() {
    int x = "string";
    return 0;
}`,
			language:       models.LanguageCpp,
			compiler:       models.CompilerGCC13,
			standard:       models.StandardCpp20,
			expectCompiled: false,
			expectSuccess:  true,
			expectedExit:   1,
			checkStderr:    true,
		},
		{
			name: "valid_template_code",
			sourceCode: `#include <iostream>
#include <vector>
template<typename T>
void print(const T& value) {
    std::cout << value << std::endl;
}
int main() {
    print(42);
    print("hello");
    return 0;
}`,
			language:       models.LanguageCpp,
			compiler:       models.CompilerGCC13,
			standard:       models.StandardCpp20,
			expectCompiled: true,
			expectSuccess:  true,
			expectedExit:   0,
			checkStderr:    false,
		},
		{
			name: "cpp11_features",
			sourceCode: `#include <iostream>
#include <vector>
int main() {
    auto x = 42;
    std::vector<int> v = {1, 2, 3, 4, 5};
    for (auto i : v) {
        std::cout << i << " ";
    }
    return 0;
}`,
			language:       models.LanguageCpp,
			compiler:       models.CompilerGCC13,
			standard:       models.StandardCpp20,
			expectCompiled: true,
			expectSuccess:  true,
			expectedExit:   0,
			checkStderr:    false,
		},
	}

	for _, tc := range testCases {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			// Each test can run in parallel since they use independent servers
			t.Parallel()

			server, err := api.NewServer()
			require.NoError(t, err, "Failed to create server")
			defer server.Close()

			e := api.NewEchoServer(server, false)

			// Encode source code
			encodedCode := base64.StdEncoding.EncodeToString([]byte(tc.sourceCode))

			// Create compilation request
			request := models.CompilationRequest{
				Code:     encodedCode,
				Language: tc.language,
				Compiler: tc.compiler,
				Standard: tc.standard,
			}

			body, err := json.Marshal(request)
			require.NoError(t, err, "Failed to marshal request")

			// Submit compilation
			req := httptest.NewRequest(http.MethodPost, "/api/v1/compile", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusAccepted, rec.Code, "Expected status 202")

			var jobResponse models.JobResponse
			err = json.NewDecoder(rec.Body).Decode(&jobResponse)
			require.NoError(t, err, "Failed to decode response")
			require.NotEmpty(t, jobResponse.JobID, "Expected job ID")

			// Wait for compilation
			time.Sleep(5 * time.Second)

			// Get result
			req = httptest.NewRequest(http.MethodGet, "/api/v1/compile/"+jobResponse.JobID, nil)
			rec = httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code, "Expected status 200")

			var result models.CompilationResult
			err = json.NewDecoder(rec.Body).Decode(&result)
			require.NoError(t, err, "Failed to decode result")

			// Assert expectations
			assert.Equal(t, tc.expectSuccess, result.Success, "Success flag mismatch")
			assert.Equal(t, tc.expectCompiled, result.Compiled, "Compiled flag mismatch")
			assert.Equal(t, tc.expectedExit, result.ExitCode, "Exit code mismatch")

			if tc.checkStderr {
				assert.NotEmpty(t, result.Stderr, "Expected stderr to contain error messages")
			}
		})
	}
}

// TestRequestValidation uses table-driven tests for request validation
// Note: The API accepts requests (returns 202) and validates them during processing
func TestRequestValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping validation test in short mode")
	}

	testCases := []struct {
		name          string
		request       models.CompilationRequest
		errorContains string // Error will be in the result, not immediate response
	}{
		{
			name: "missing_code",
			request: models.CompilationRequest{
				Language: models.LanguageCpp,
				Compiler: models.CompilerGCC13,
				Standard: models.StandardCpp20,
			},
			errorContains: "source code is required",
		},
		{
			name: "missing_language",
			request: models.CompilationRequest{
				Code:     base64.StdEncoding.EncodeToString([]byte("int main() { return 0; }")),
				Compiler: models.CompilerGCC13,
				Standard: models.StandardCpp20,
			},
			errorContains: "invalid language",
		},
		{
			name: "unsupported_language",
			request: models.CompilationRequest{
				Code:     base64.StdEncoding.EncodeToString([]byte("package main")),
				Language: models.LanguageGo,
				Compiler: models.CompilerGo,
			},
			errorContains: "unsupported",
		},
		{
			name: "code_too_large",
			request: models.CompilationRequest{
				Code:     base64.StdEncoding.EncodeToString(make([]byte, 2*1024*1024)), // 2MB
				Language: models.LanguageCpp,
				Compiler: models.CompilerGCC13,
				Standard: models.StandardCpp20,
			},
			errorContains: "too large",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server, err := api.NewServer()
			require.NoError(t, err, "Failed to create server")
			defer server.Close()

			e := api.NewEchoServer(server, false)

			body, err := json.Marshal(tc.request)
			require.NoError(t, err, "Failed to marshal request")

			// Submit request
			req := httptest.NewRequest(http.MethodPost, "/api/v1/compile", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			// API accepts all requests initially
			assert.Equal(t, http.StatusAccepted, rec.Code, "Expected status 202")

			var jobResponse models.JobResponse
			err = json.NewDecoder(rec.Body).Decode(&jobResponse)
			require.NoError(t, err, "Failed to decode response")
			require.NotEmpty(t, jobResponse.JobID, "Expected job ID")

			// Wait for processing
			time.Sleep(2 * time.Second)

			// Check result - validation errors appear here
			req = httptest.NewRequest(http.MethodGet, "/api/v1/compile/"+jobResponse.JobID, nil)
			rec = httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			var result models.CompilationResult
			err = json.NewDecoder(rec.Body).Decode(&result)
			require.NoError(t, err, "Failed to decode result")

			// Validation failures result in Success=false with error message
			assert.False(t, result.Success, "Expected validation to fail")
			assert.Contains(t, result.Error, tc.errorContains, "Error message mismatch")
		})
	}
}

// TestEnvironmentSpecs uses table-driven tests to verify environment specifications
func TestEnvironmentSpecs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testCases := []struct {
		name             string
		expectedLanguage string
		expectedCompiler string
		minVersion       string
	}{
		{
			name:             "cpp_environment",
			expectedLanguage: "cpp",
			expectedCompiler: "gcc",
			minVersion:       "13",
		},
		// Future test cases for other languages can be added here
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			server, err := api.NewServer()
			require.NoError(t, err, "Failed to create server")
			defer server.Close()

			e := api.NewEchoServer(server, false)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/environments", nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code, "Expected status 200")

			var environments []models.Environment
			err = json.NewDecoder(rec.Body).Decode(&environments)
			require.NoError(t, err, "Failed to decode response")

			// Find the environment
			found := false
			for _, env := range environments {
				if env.Language == tc.expectedLanguage {
					found = true
					// Check if expected compiler is in the list
					compilerFound := false
					for _, compiler := range env.Compilers {
						if strings.Contains(compiler, tc.expectedCompiler) {
							compilerFound = true
							break
						}
					}
					assert.True(t, compilerFound, "Expected compiler %s not found", tc.expectedCompiler)
					break
				}
			}

			assert.True(t, found, "Expected environment %s not found", tc.expectedLanguage)
		})
	}
}
