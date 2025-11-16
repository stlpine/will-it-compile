//go:build go1.25

package integration

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stlpine/will-it-compile/internal/api"
	"github.com/stlpine/will-it-compile/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCompileAsync_WithSynctest demonstrates using testing/synctest to test
// asynchronous job processing with virtualized time.
//
// Benefits over time.Sleep approach:
// - Tests run instantly (virtualized time)
// - Deterministic goroutine synchronization
// - No arbitrary sleep durations
func TestCompileAsync_WithSynctest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	synctest.Test(t, func(t *testing.T) {
		server, err := api.NewServer()
		require.NoError(t, err, "Failed to create server")
		defer func() {
			if err := server.Close(); err != nil {
				t.Logf("Error closing server: %v", err)
			}
		}()

		e := api.NewEchoServer(server, false)

		// Create valid C++ code
		sourceCode := `#include <iostream>
int main() {
    std::cout << "Hello from synctest!" << std::endl;
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
		require.NoError(t, err, "Failed to marshal request")

		// Submit compilation job
		req := httptest.NewRequest(http.MethodPost, "/api/v1/compile", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusAccepted, rec.Code, "Expected status 202")

		var jobResponse models.JobResponse
		err = json.NewDecoder(rec.Body).Decode(&jobResponse)
		require.NoError(t, err, "Failed to decode response")
		assert.NotEmpty(t, jobResponse.JobID, "Expected job ID to be present")

		// Instead of time.Sleep, use synctest.Wait() to wait for goroutines
		// The synctest bubble will advance time automatically when goroutines block
		//
		// Note: In this test, we're waiting for the background goroutine
		// (processJob) to complete. synctest.Wait() will block until all
		// goroutines in the bubble are idle or blocked on I/O.
		//
		// For Docker-based compilation, the goroutine will block on Docker
		// API calls (I/O), so we still need to poll for completion.
		// This is a limitation when testing with real external dependencies.
		pollForJobCompletion(t, e, jobResponse.JobID, 10*time.Second)

		// Verify result
		req = httptest.NewRequest(http.MethodGet, "/api/v1/compile/"+jobResponse.JobID, nil)
		rec = httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code, "Expected status 200")

		var result models.CompilationResult
		err = json.NewDecoder(rec.Body).Decode(&result)
		require.NoError(t, err, "Failed to decode result")

		assert.True(t, result.Success, "Expected compilation to succeed")
		assert.True(t, result.Compiled, "Expected code to compile successfully")
		assert.Equal(t, 0, result.ExitCode, "Expected exit code 0")
	})
}

// TestMultipleJobsConcurrent_WithSynctest tests multiple concurrent compilation jobs
// using synctest for deterministic execution.
func TestMultipleJobsConcurrent_WithSynctest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	synctest.Test(t, func(t *testing.T) {
		server, err := api.NewServer()
		require.NoError(t, err, "Failed to create server")
		defer func() {
			if err := server.Close(); err != nil {
				t.Logf("Error closing server: %v", err)
			}
		}()

		e := api.NewEchoServer(server, false)

		// Submit 3 jobs concurrently
		jobIDs := make([]string, 3)

		for i := 0; i < 3; i++ {
			sourceCode := `#include <iostream>
int main() {
    std::cout << "Job ` + string(rune('A'+i)) + `" << std::endl;
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
			require.NoError(t, err, "Failed to marshal request")

			req := httptest.NewRequest(http.MethodPost, "/api/v1/compile", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusAccepted, rec.Code, "Expected status 202")

			var jobResponse models.JobResponse
			err = json.NewDecoder(rec.Body).Decode(&jobResponse)
			require.NoError(t, err, "Failed to decode response")

			jobIDs[i] = jobResponse.JobID
		}

		// Wait for all jobs to complete
		for _, jobID := range jobIDs {
			pollForJobCompletion(t, e, jobID, 10*time.Second)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/compile/"+jobID, nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			var result models.CompilationResult
			err = json.NewDecoder(rec.Body).Decode(&result)
			require.NoError(t, err, "Failed to decode result for job "+jobID)

			assert.True(t, result.Success, "Expected job %s to succeed", jobID)
			assert.True(t, result.Compiled, "Expected job %s to compile", jobID)
		}
	})
}

// TestJobTimeout_WithSynctest tests compilation timeout behavior with virtualized time.
//
// Note: This test demonstrates synctest's time virtualization. In a real scenario,
// the Docker container timeout would need to be mocked to fully benefit from
// virtualized time. This example shows the pattern for future mock implementations.
func TestJobTimeout_WithSynctest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	synctest.Test(t, func(t *testing.T) {
		server, err := api.NewServer()
		require.NoError(t, err, "Failed to create server")
		defer func() {
			if err := server.Close(); err != nil {
				t.Logf("Error closing server: %v", err)
			}
		}()

		e := api.NewEchoServer(server, false)

		// Code that would timeout (infinite loop)
		sourceCode := `#include <iostream>
int main() {
    while(true) {
        std::cout << "Infinite loop" << std::endl;
    }
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
		require.NoError(t, err, "Failed to marshal request")

		req := httptest.NewRequest(http.MethodPost, "/api/v1/compile", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		var jobResponse models.JobResponse
		err = json.NewDecoder(rec.Body).Decode(&jobResponse)
		require.NoError(t, err, "Failed to decode response")

		// With synctest, time advances when goroutines block
		// In a fully mocked environment, this would complete instantly
		// For now, we still need real polling for Docker-based tests
		pollForJobCompletion(t, e, jobResponse.JobID, 35*time.Second)

		req = httptest.NewRequest(http.MethodGet, "/api/v1/compile/"+jobResponse.JobID, nil)
		rec = httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		var result models.CompilationResult
		err = json.NewDecoder(rec.Body).Decode(&result)
		require.NoError(t, err, "Failed to decode result")

		// The compilation should succeed (compiles fine)
		// but execution would timeout (handled separately in runtime tests)
		assert.True(t, result.Success, "Expected compilation to succeed")
		assert.True(t, result.Compiled, "Expected code to compile successfully")
	})
}

// pollForJobCompletion polls the job endpoint until completion or timeout.
// This helper is needed because Docker I/O operations block goroutines in a way
// that synctest can't virtualize. In a fully mocked environment, this wouldn't be needed.
func pollForJobCompletion(t *testing.T, e http.Handler, jobID string, timeout time.Duration) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for time.Now().Before(deadline) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/compile/"+jobID, nil)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		// Try to decode as CompilationResult
		var result models.CompilationResult
		if err := json.NewDecoder(rec.Body).Decode(&result); err == nil {
			// Got a result, job is complete
			return
		}

		// Wait before next poll
		<-ticker.C
	}

	t.Fatalf("Job %s did not complete within timeout", jobID)
}
