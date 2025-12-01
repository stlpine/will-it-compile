package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stlpine/will-it-compile/internal/compiler"
	"github.com/stlpine/will-it-compile/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandleCompile_NoWorkersAvailable tests that compile requests are rejected
// when all workers are busy.
func TestHandleCompile_NoWorkersAvailable(t *testing.T) {
	// Create a server with 0 workers (simulates all workers busy)
	config := ServerConfig{
		MaxWorkers: 0, // No workers available
		QueueSize:  10,
	}

	// Use mock compiler to avoid Docker dependency
	server := &Server{
		compiler: &httpMockCompiler{},
		jobs:     newHTTPMockJobStore(),
	}
	server.workerPool = NewWorkerPool(config.MaxWorkers, config.QueueSize, server)
	server.workerPool.Start()
	defer server.workerPool.Stop()

	// Create request
	reqBody := models.CompilationRequest{
		Code:     "aW50IG1haW4oKSB7IHJldHVybiAwOyB9", // base64 encoded "int main() { return 0; }"
		Language: models.LanguageCpp,
	}
	bodyBytes, err := json.Marshal(reqBody)
	require.NoError(t, err)

	// Create HTTP request
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/compile", bytes.NewReader(bodyBytes))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute handler
	err = server.HandleCompile(c)

	// Verify error response
	require.Error(t, err)
	httpErr, ok := err.(*echo.HTTPError)
	require.True(t, ok, "Expected echo.HTTPError")
	assert.Equal(t, http.StatusTooManyRequests, httpErr.Code)
	assert.Contains(t, httpErr.Message, "no workers available")
}

// TestHandleCompile_WorkersAvailable tests that compile requests are accepted
// when workers are available.
func TestHandleCompile_WorkersAvailable(t *testing.T) {
	// Create a server with workers available
	config := ServerConfig{
		MaxWorkers: 2,
		QueueSize:  10,
	}

	server := &Server{
		compiler: &httpMockCompiler{},
		jobs:     newHTTPMockJobStore(),
	}
	server.workerPool = NewWorkerPool(config.MaxWorkers, config.QueueSize, server)
	server.workerPool.Start()
	defer server.workerPool.Stop()

	// Create request
	reqBody := models.CompilationRequest{
		Code:     "aW50IG1haW4oKSB7IHJldHVybiAwOyB9", // base64 encoded
		Language: models.LanguageCpp,
	}
	bodyBytes, err := json.Marshal(reqBody)
	require.NoError(t, err)

	// Create HTTP request
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/compile", bytes.NewReader(bodyBytes))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute handler
	err = server.HandleCompile(c)

	// Verify success
	require.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, rec.Code)

	// Verify response body
	var resp models.JobResponse
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.JobID)
	assert.Equal(t, models.StatusQueued, resp.Status)
}

// Mock implementations for testing (named differently to avoid conflicts with async_job_test.go)

type httpMockCompiler struct{}

func (m *httpMockCompiler) Compile(_ context.Context, _ models.CompilationJob) models.CompilationResult {
	return models.CompilationResult{
		Success:  true,
		Compiled: true,
		ExitCode: 0,
		Stdout:   "compiled successfully",
	}
}

func (m *httpMockCompiler) GetSupportedEnvironments() []models.Environment {
	return []models.Environment{
		{
			Language:  "cpp",
			Compilers: []string{"gcc-13"},
			Standards: []string{"c++17"},
			OSes:      []string{"linux"},
			Arches:    []string{"amd64"},
		},
	}
}

func (m *httpMockCompiler) Close() error {
	return nil
}

// Ensure httpMockCompiler implements the compiler interface
var _ compiler.CompilerInterface = (*httpMockCompiler)(nil)

type httpMockJobStore struct {
	jobs    map[string]models.CompilationJob
	results map[string]models.CompilationResult
}

func newHTTPMockJobStore() *httpMockJobStore {
	return &httpMockJobStore{
		jobs:    make(map[string]models.CompilationJob),
		results: make(map[string]models.CompilationResult),
	}
}

func (s *httpMockJobStore) Store(job models.CompilationJob) error {
	s.jobs[job.ID] = job
	return nil
}

func (s *httpMockJobStore) Get(jobID string) (models.CompilationJob, bool) {
	job, exists := s.jobs[jobID]
	return job, exists
}

func (s *httpMockJobStore) StoreResult(jobID string, result models.CompilationResult) error {
	s.results[jobID] = result
	return nil
}

func (s *httpMockJobStore) GetResult(jobID string) (models.CompilationResult, bool) {
	result, exists := s.results[jobID]
	return result, exists
}

func (s *httpMockJobStore) Close() error {
	return nil
}
