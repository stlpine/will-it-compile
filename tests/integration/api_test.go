package integration

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stlpine/will-it-compile/internal/api"
	"github.com/stlpine/will-it-compile/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthEndpoint(t *testing.T) {
	server, err := api.NewServer()
	require.NoError(t, err, "Failed to create server")
	defer server.Close()

	e := api.NewEchoServer(server, false)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code, "Expected status 200")

	var response map[string]string
	err = json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err, "Failed to decode response")

	assert.Equal(t, "healthy", response["status"], "Expected status 'healthy'")
}

func TestGetEnvironments(t *testing.T) {
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

	assert.NotEmpty(t, environments, "Expected at least one environment")

	// Check that C++ environment exists
	found := false
	for _, env := range environments {
		if env.Language == "cpp" {
			found = true
			break
		}
	}

	assert.True(t, found, "Expected C++ environment to be present")
}

func TestCompileValidCode(t *testing.T) {
	// Skip if Docker is not available
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	server, err := api.NewServer()
	require.NoError(t, err, "Failed to create server")
	defer server.Close()

	e := api.NewEchoServer(server, false)

	// Create valid C++ code
	sourceCode := `#include <iostream>
int main() {
    std::cout << "Hello, World!" << std::endl;
    return 0;
}`

	encodedCode := base64.StdEncoding.EncodeToString([]byte(sourceCode))

	request := models.CompilationRequest{
		Code:     encodedCode,
		Language: models.LanguageCpp,
		Compiler: models.CompilerGCC9,
		Standard: models.StandardCpp11,
	}

	body, err := json.Marshal(request)
	require.NoError(t, err, "Failed to marshal request")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/compile", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusAccepted, rec.Code, "Expected status 202")

	var jobResponse models.JobResponse
	err = json.NewDecoder(rec.Body).Decode(&jobResponse)
	require.NoError(t, err, "Failed to decode response")

	assert.NotEmpty(t, jobResponse.JobID, "Expected job ID to be present")

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

	assert.True(t, result.Success, "Expected compilation to succeed, got error: %s", result.Error)
	assert.True(t, result.Compiled, "Expected code to compile successfully, stderr: %s", result.Stderr)
	assert.Equal(t, 0, result.ExitCode, "Expected exit code 0")
}

func TestCompileInvalidCode(t *testing.T) {
	// Skip if Docker is not available
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	server, err := api.NewServer()
	require.NoError(t, err, "Failed to create server")
	defer server.Close()

	e := api.NewEchoServer(server, false)

	// Create invalid C++ code (missing semicolon)
	sourceCode := `#include <iostream>
int main() {
    std::cout << "Missing semicolon"
    return 0;
}`

	encodedCode := base64.StdEncoding.EncodeToString([]byte(sourceCode))

	request := models.CompilationRequest{
		Code:     encodedCode,
		Language: models.LanguageCpp,
		Compiler: models.CompilerGCC9,
		Standard: models.StandardCpp11,
	}

	body, err := json.Marshal(request)
	require.NoError(t, err, "Failed to marshal request")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/compile", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusAccepted, rec.Code, "Expected status 202")

	var jobResponse models.JobResponse
	err = json.NewDecoder(rec.Body).Decode(&jobResponse)
	require.NoError(t, err, "Failed to decode response")

	// Wait for compilation
	time.Sleep(5 * time.Second)

	// Get result
	req = httptest.NewRequest(http.MethodGet, "/api/v1/compile/"+jobResponse.JobID, nil)
	rec = httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	var result models.CompilationResult
	err = json.NewDecoder(rec.Body).Decode(&result)
	require.NoError(t, err, "Failed to decode result")

	assert.False(t, result.Compiled, "Expected code to fail compilation")
	assert.NotEqual(t, 0, result.ExitCode, "Expected non-zero exit code")
	assert.NotEmpty(t, result.Stderr, "Expected error message in stderr")
}
