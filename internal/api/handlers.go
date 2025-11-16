package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stlpine/will-it-compile/internal/compiler"
	"github.com/stlpine/will-it-compile/pkg/models"
)

// Server represents the API server.
type Server struct {
	compiler compiler.CompilerInterface
	jobs     *jobStore
}

// NewServer creates a new API server instance.
func NewServer() (*Server, error) {
	comp, err := compiler.NewCompiler()
	if err != nil {
		return nil, fmt.Errorf("failed to create compiler: %w", err)
	}

	return &Server{
		compiler: comp,
		jobs:     newJobStore(),
	}, nil
}

// Close cleans up server resources.
func (s *Server) Close() error {
	return s.compiler.Close()
}

// =============================================================================
// HTTP Handlers
// =============================================================================

// HandleCompile submits a new compilation request
//
// @HTTP   POST /api/v1/compile
// @Accept application/json
// @Param  request body models.CompilationRequest true "Compilation request"
// @Return 202 {object} models.JobResponse "Job created and queued"
// @Return 400 {object} models.ErrorResponse "Invalid request body".
func (s *Server) HandleCompile(c echo.Context) error {
	// Parse request body
	var req models.CompilationRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	// Create compilation job
	job := models.CompilationJob{
		ID:        uuid.New().String(),
		Request:   req,
		Status:    models.StatusQueued,
		CreatedAt: time.Now(),
	}

	// Store job
	s.jobs.Store(job)

	// Process asynchronously (MVP: goroutine, Phase 2: message queue)
	go s.processJob(job)

	// Return job response
	response := models.JobResponse{
		JobID:  job.ID,
		Status: models.StatusQueued,
	}

	return c.JSON(http.StatusAccepted, response)
}

// HandleGetJob retrieves the status and result of a compilation job
//
// @HTTP   GET /api/v1/compile/:job_id
// @Param  job_id path string true "Job ID"
// @Return 200 {object} models.CompilationResult "Compilation result (if completed)"
// @Return 200 {object} models.JobResponse "Job status (if still processing)"
// @Return 400 {object} models.ErrorResponse "Missing job ID"
// @Return 404 {object} models.ErrorResponse "Job not found".
func (s *Server) HandleGetJob(c echo.Context) error {
	// Extract job ID from URL path parameter
	jobID := c.Param("job_id")
	if jobID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "job ID required")
	}

	// Check if job exists
	job, exists := s.jobs.Get(jobID)
	if !exists {
		return echo.NewHTTPError(http.StatusNotFound, "job not found")
	}

	// Check if result is available
	result, hasResult := s.jobs.GetResult(jobID)
	if hasResult {
		return c.JSON(http.StatusOK, result)
	}

	// Return current job status
	return c.JSON(http.StatusOK, models.JobResponse{
		JobID:  job.ID,
		Status: job.Status,
	})
}

// HandleGetEnvironments returns a list of supported compilation environments
//
// @HTTP   GET /api/v1/environments
// @Return 200 {array} models.Environment "List of supported environments".
func (s *Server) HandleGetEnvironments(c echo.Context) error {
	environments := s.compiler.GetSupportedEnvironments()
	return c.JSON(http.StatusOK, environments)
}

// HandleHealth returns the health status of the service
//
// @HTTP   GET /health
// @Return 200 {object} map[string]string "Health status".
func (s *Server) HandleHealth(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}
