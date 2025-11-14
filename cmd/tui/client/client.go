package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/stlpine/will-it-compile/pkg/models"
)

// Sentinel errors for TUI client.
var (
	ErrAPIError          = errors.New("API error")
	ErrJobNotFound       = errors.New("job not found")
	ErrHealthCheckFailed = errors.New("health check failed")
)

// Client is an HTTP client for the will-it-compile API.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new API client.
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// SubmitCompilation submits a compilation job.
func (c *Client) SubmitCompilation(ctx context.Context, req models.CompilationRequest) (*models.CompilationJob, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v1/compile", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck // standard practice for HTTP client

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body) //nolint:errcheck // best effort error message
		return nil, fmt.Errorf("%w (status %d): %s", ErrAPIError, resp.StatusCode, string(body))
	}

	var job models.CompilationJob
	if err := json.NewDecoder(resp.Body).Decode(&job); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &job, nil
}

// JobStatus represents the status of a job.
type JobStatus struct {
	JobID  string                    `json:"job_id,omitempty"`
	Status models.JobStatus          `json:"status,omitempty"`
	Result *models.CompilationResult `json:"result,omitempty"`
}

// GetJob retrieves a compilation job status/result by ID
// The API returns either JobResponse (if pending) or CompilationResult (if complete).
func (c *Client) GetJob(ctx context.Context, jobID string) (*JobStatus, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/v1/compile/"+jobID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck // standard practice for HTTP client

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrJobNotFound
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body) //nolint:errcheck // best effort error message
		return nil, fmt.Errorf("%w (status %d): %s", ErrAPIError, resp.StatusCode, string(body))
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Try to decode as CompilationResult first
	var result models.CompilationResult
	if err := json.Unmarshal(body, &result); err == nil && result.JobID != "" {
		return &JobStatus{
			JobID:  result.JobID,
			Status: models.StatusCompleted,
			Result: &result,
		}, nil
	}

	// Otherwise decode as JobResponse
	var jobResp models.JobResponse
	if err := json.Unmarshal(body, &jobResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &JobStatus{
		JobID:  jobResp.JobID,
		Status: jobResp.Status,
		Result: nil,
	}, nil
}

// GetEnvironments retrieves the list of supported environments.
func (c *Client) GetEnvironments(ctx context.Context) ([]models.EnvironmentSpec, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/v1/environments", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck // standard practice for HTTP client

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body) //nolint:errcheck // best effort error message
		return nil, fmt.Errorf("%w (status %d): %s", ErrAPIError, resp.StatusCode, string(body))
	}

	var response struct {
		Environments []models.EnvironmentSpec `json:"environments"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Environments, nil
}

// HealthCheck performs a health check on the API.
func (c *Client) HealthCheck(ctx context.Context) error {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/health", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck // standard practice for HTTP client

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w (status %d)", ErrHealthCheckFailed, resp.StatusCode)
	}

	return nil
}
