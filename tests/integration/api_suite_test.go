package integration

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stlpine/will-it-compile/internal/api"
	"github.com/stlpine/will-it-compile/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// APISuite is a test suite for API integration tests
// It provides common setup/teardown and helper methods
type APISuite struct {
	suite.Suite
	server *api.Server
	echo   *echo.Echo
}

// SetupSuite runs once before all tests in the suite
func (s *APISuite) SetupSuite() {
	// This would be a good place to ensure Docker images are built
	// or to do other one-time setup
}

// SetupTest runs before each test in the suite
func (s *APISuite) SetupTest() {
	server, err := api.NewServer()
	require.NoError(s.T(), err, "Failed to create server")
	s.server = server
	s.echo = api.NewEchoServer(server, false)
}

// TearDownTest runs after each test in the suite
func (s *APISuite) TearDownTest() {
	if s.server != nil {
		s.server.Close()
	}
}

// Helper method to submit a compilation request
func (s *APISuite) submitCompilation(request models.CompilationRequest) models.JobResponse {
	body, err := json.Marshal(request)
	require.NoError(s.T(), err, "Failed to marshal request")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/compile", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	s.echo.ServeHTTP(rec, req)

	assert.Equal(s.T(), http.StatusAccepted, rec.Code, "Expected status 202")

	var jobResponse models.JobResponse
	err = json.NewDecoder(rec.Body).Decode(&jobResponse)
	require.NoError(s.T(), err, "Failed to decode response")

	return jobResponse
}

// Helper method to get compilation result
func (s *APISuite) getCompilationResult(jobID string) models.CompilationResult {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/compile/"+jobID, nil)
	rec := httptest.NewRecorder()

	s.echo.ServeHTTP(rec, req)

	assert.Equal(s.T(), http.StatusOK, rec.Code, "Expected status 200")

	var result models.CompilationResult
	err := json.NewDecoder(rec.Body).Decode(&result)
	require.NoError(s.T(), err, "Failed to decode result")

	return result
}

// Helper method to create a compilation request
func (s *APISuite) createCppRequest(sourceCode string) models.CompilationRequest {
	encodedCode := base64.StdEncoding.EncodeToString([]byte(sourceCode))
	return models.CompilationRequest{
		Code:     encodedCode,
		Language: models.LanguageCpp,
		Compiler: models.CompilerGCC13,
		Standard: models.StandardCpp20,
	}
}

// TestAPISuite_Health tests the health endpoint
func (s *APISuite) TestHealth() {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	s.echo.ServeHTTP(rec, req)

	assert.Equal(s.T(), http.StatusOK, rec.Code, "Expected status 200")

	var response map[string]string
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(s.T(), err, "Failed to decode response")

	assert.Equal(s.T(), "healthy", response["status"], "Expected status 'healthy'")
}

// TestAPISuite_GetEnvironments tests the environments endpoint
func (s *APISuite) TestGetEnvironments() {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/environments", nil)
	rec := httptest.NewRecorder()

	s.echo.ServeHTTP(rec, req)

	assert.Equal(s.T(), http.StatusOK, rec.Code, "Expected status 200")

	var environments []models.Environment
	err := json.NewDecoder(rec.Body).Decode(&environments)
	require.NoError(s.T(), err, "Failed to decode response")

	assert.NotEmpty(s.T(), environments, "Expected at least one environment")

	// Check that C++ environment exists
	found := false
	for _, env := range environments {
		if env.Language == "cpp" {
			found = true
			break
		}
	}

	assert.True(s.T(), found, "Expected C++ environment to be present")
}

// TestAPISuite_CompileValidCode tests successful compilation
func (s *APISuite) TestCompileValidCode() {
	if testing.Short() {
		s.T().Skip("Skipping integration test in short mode")
	}

	sourceCode := `#include <iostream>
int main() {
    std::cout << "Hello, World!" << std::endl;
    return 0;
}`

	request := s.createCppRequest(sourceCode)
	jobResponse := s.submitCompilation(request)

	assert.NotEmpty(s.T(), jobResponse.JobID, "Expected job ID to be present")

	// Wait for compilation
	time.Sleep(5 * time.Second)

	result := s.getCompilationResult(jobResponse.JobID)

	assert.True(s.T(), result.Success, "Expected compilation to succeed")
	assert.True(s.T(), result.Compiled, "Expected code to compile successfully")
	assert.Equal(s.T(), 0, result.ExitCode, "Expected exit code 0")
}

// TestAPISuite_CompileInvalidCode tests failed compilation
func (s *APISuite) TestCompileInvalidCode() {
	if testing.Short() {
		s.T().Skip("Skipping integration test in short mode")
	}

	sourceCode := `#include <iostream>
int main() {
    std::cout << "Missing semicolon"
    return 0;
}`

	request := s.createCppRequest(sourceCode)
	jobResponse := s.submitCompilation(request)

	// Wait for compilation
	time.Sleep(5 * time.Second)

	result := s.getCompilationResult(jobResponse.JobID)

	assert.False(s.T(), result.Compiled, "Expected code to fail compilation")
	assert.NotEqual(s.T(), 0, result.ExitCode, "Expected non-zero exit code")
	assert.NotEmpty(s.T(), result.Stderr, "Expected error message in stderr")
}

// TestAPISuite runs the test suite
func TestAPISuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test suite in short mode")
	}
	suite.Run(t, new(APISuite))
}
