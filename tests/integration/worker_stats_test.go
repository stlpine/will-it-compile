package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stlpine/will-it-compile/internal/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWorkerStatsEndpoint verifies that the /api/v1/workers/stats endpoint exists and returns valid stats.
func TestWorkerStatsEndpoint(t *testing.T) {
	// Create server with default config
	server, err := api.NewServer()
	require.NoError(t, err, "failed to create server")
	defer func() {
		if err := server.Close(); err != nil {
			t.Logf("Error closing server: %v", err)
		}
	}()

	// Create Echo instance without rate limiting for tests
	e := api.NewEchoServer(server, false)

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/api/v1/workers/stats", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	// Execute request
	e.ServeHTTP(rec, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rec.Code, "expected 200 OK, got %d: %s", rec.Code, rec.Body.String())

	// Parse response
	var stats api.WorkerStats
	err = json.Unmarshal(rec.Body.Bytes(), &stats)
	require.NoError(t, err, "failed to parse response JSON")

	// Verify stats structure
	assert.Equal(t, 5, stats.MaxWorkers, "expected 5 max workers")
	assert.GreaterOrEqual(t, stats.ActiveWorkers, 0, "active workers should be non-negative")
	assert.GreaterOrEqual(t, stats.AvailableSlots, 0, "available slots should be non-negative")
	assert.GreaterOrEqual(t, stats.QueuedJobs, 0, "queued jobs should be non-negative")
	assert.GreaterOrEqual(t, stats.TotalProcessed, int64(0), "total processed should be non-negative")
	assert.GreaterOrEqual(t, stats.TotalSuccessful, int64(0), "total successful should be non-negative")
	assert.GreaterOrEqual(t, stats.TotalFailed, int64(0), "total failed should be non-negative")
	assert.NotEmpty(t, stats.Uptime, "uptime should not be empty")
	assert.GreaterOrEqual(t, stats.UptimeSeconds, int64(0), "uptime seconds should be non-negative")
	assert.False(t, stats.StartTime.IsZero(), "start time should be set")

	t.Logf("Worker stats: %+v", stats)
}

// TestWorkerStatsEndpointNotFound verifies we don't accidentally break the route.
func TestWorkerStatsEndpointNotFound(t *testing.T) {
	// Create server
	server, err := api.NewServer()
	require.NoError(t, err, "failed to create server")
	defer func() {
		if err := server.Close(); err != nil {
			t.Logf("Error closing server: %v", err)
		}
	}()

	e := api.NewEchoServer(server, false)

	// Test wrong path should return 404
	testCases := []string{
		"/api/v1/worker/stats",      // Missing 's' in 'workers'
		"/api/v1/workers/statistics", // Wrong endpoint name
		"/api/v2/workers/stats",      // Wrong API version
		"/workers/stats",             // Missing API prefix
	}

	for _, path := range testCases {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusNotFound, rec.Code,
				"expected 404 for wrong path %s, got %d", path, rec.Code)
		})
	}
}
