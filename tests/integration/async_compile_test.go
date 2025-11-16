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

// TestAsyncCompilation tests asynchronous compilation with real Docker.
// Note: Uses real time because Docker I/O operations are not durably blocking
// and don't work with synctest's virtualized time.
func TestAsyncCompilation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	server, err := api.NewServer()
	require.NoError(t, err)
	defer func() {
		if err := server.Close(); err != nil {
			t.Logf("Error closing server: %v", err)
		}
	}()

	e := api.NewEchoServer(server, false)

	sourceCode := `#include <iostream>
int main() {
    std::cout << "Hello, World!" << std::endl;
    return 0;
}`

	encodedCode := base64.StdEncoding.EncodeToString([]byte(sourceCode))
	request := models.CompilationRequest{
		Code:     encodedCode,
		Language: models.LanguageCpp,
		Compiler: models.CompilerGCC13,
		Standard: models.StandardCpp20,
	}

	body, err := json.Marshal(request)
	require.NoError(t, err)

	// Submit job
	req := httptest.NewRequest(http.MethodPost, "/api/v1/compile", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusAccepted, rec.Code)

	var jobResponse models.JobResponse
	err = json.NewDecoder(rec.Body).Decode(&jobResponse)
	require.NoError(t, err)
	assert.NotEmpty(t, jobResponse.JobID)

	// Poll for completion (Docker I/O is not durably blocking)
	pollForCompletion(t, e, jobResponse.JobID, 10*time.Second)

	// Verify result
	req = httptest.NewRequest(http.MethodGet, "/api/v1/compile/"+jobResponse.JobID, nil)
	rec = httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var result models.CompilationResult
	err = json.NewDecoder(rec.Body).Decode(&result)
	require.NoError(t, err)

	assert.True(t, result.Success)
	assert.True(t, result.Compiled)
	assert.Equal(t, 0, result.ExitCode)
}

// pollForCompletion polls until job completes or timeout.
// Required because Docker I/O operations are not durably blocking.
func pollForCompletion(t *testing.T, handler http.Handler, jobID string, timeout time.Duration) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for time.Now().Before(deadline) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/compile/"+jobID, nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		var result models.CompilationResult
		if err := json.NewDecoder(rec.Body).Decode(&result); err == nil {
			// Check if this is actually a completed result (not just a JobResponse)
			// A completed result will have Duration > 0 or JobID will be set with actual data
			if result.JobID != "" && result.Duration > 0 {
				return
			}
		}

		<-ticker.C
	}

	t.Fatalf("Job %s did not complete within %v", jobID, timeout)
}
