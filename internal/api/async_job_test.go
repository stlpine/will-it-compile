//go:build go1.25

package api

import (
	"context"
	"fmt"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stlpine/will-it-compile/internal/compiler"
	"github.com/stlpine/will-it-compile/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockCompiler implements a compiler for testing that uses time.Sleep
// to simulate compilation delay. With virtualized time, these delays happen instantly.
type mockCompiler struct {
	compileDelay time.Duration
	shouldFail   bool
}

func (m *mockCompiler) Compile(ctx context.Context, job models.CompilationJob) models.CompilationResult {
	// Simulate compilation time - happens instantly with virtualized time
	time.Sleep(m.compileDelay)

	if m.shouldFail {
		return models.CompilationResult{
			Success:  false,
			Compiled: false,
			Error:    "compilation error",
			ExitCode: 1,
			Stderr:   "error: compilation failed",
		}
	}

	return models.CompilationResult{
		Success:  true,
		Compiled: true,
		ExitCode: 0,
		Stdout:   "compilation successful",
	}
}

func (m *mockCompiler) GetSupportedEnvironments() []models.Environment {
	return []models.Environment{
		{
			Language:  "cpp",
			Compilers: []string{"gcc-13"},
			Standards: []string{"c++11", "c++14", "c++17", "c++20"},
			OSes:      []string{"linux"},
			Arches:    []string{"amd64"},
		},
	}
}

func (m *mockCompiler) Close() error {
	return nil
}

// Ensure mockCompiler implements the compiler interface at compile time
var _ compiler.CompilerInterface = (*mockCompiler)(nil)

// TestAsyncJobProcessing tests async job processing with virtualized time.
// Tests that would normally take seconds run instantly.
func TestAsyncJobProcessing(t *testing.T) {
	tests := []struct {
		name         string
		delay        time.Duration
		shouldFail   bool
		expectStatus models.JobStatus
		expectError  bool
	}{
		{
			name:         "successful compilation",
			delay:        2 * time.Second,
			shouldFail:   false,
			expectStatus: models.StatusCompleted,
			expectError:  false,
		},
		{
			name:         "failed compilation",
			delay:        500 * time.Millisecond,
			shouldFail:   true,
			expectStatus: models.StatusFailed,
			expectError:  true,
		},
		{
			name:         "long running success",
			delay:        5 * time.Second,
			shouldFail:   false,
			expectStatus: models.StatusCompleted,
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				server := &Server{
					compiler: &mockCompiler{compileDelay: tt.delay, shouldFail: tt.shouldFail},
					jobs:     newJobStore(),
				}

				job := models.CompilationJob{
					ID: "test-job",
					Request: models.CompilationRequest{
						Code:     "I2luY2x1ZGUgPGlvc3RyZWFtPg==",
						Language: models.LanguageCpp,
						Compiler: models.CompilerGCC13,
					},
					Status:    models.StatusQueued,
					CreatedAt: time.Now(),
				}

				// Process job asynchronously
				done := make(chan struct{})
				go func() {
					server.processJob(job)
					close(done)
				}()

				// Wait for completion - happens instantly!
				<-done

				// Verify job status
				processedJob, exists := server.jobs.Get(job.ID)
				require.True(t, exists)
				assert.Equal(t, tt.expectStatus, processedJob.Status)

				// Verify result
				result, hasResult := server.jobs.GetResult(job.ID)
				require.True(t, hasResult)

				if tt.expectError {
					assert.False(t, result.Success)
					assert.False(t, result.Compiled)
					assert.NotEqual(t, 0, result.ExitCode)
					assert.NotEmpty(t, result.Error)
				} else {
					assert.True(t, result.Success)
					assert.True(t, result.Compiled)
					assert.Equal(t, 0, result.ExitCode)
				}
			})
		})
	}
}

// TestConcurrentJobs verifies multiple jobs process correctly in parallel.
// Without virtualized time, this would take 5 seconds. With it: instant.
func TestConcurrentJobs(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		server := &Server{
			compiler: &mockCompiler{compileDelay: 1 * time.Second, shouldFail: false},
			jobs:     newJobStore(),
		}

		jobCount := 5
		done := make(chan string, jobCount)

		// Start multiple jobs concurrently
		for i := 0; i < jobCount; i++ {
			job := models.CompilationJob{
				ID: fmt.Sprintf("job-%d", i),
				Request: models.CompilationRequest{
					Code:     "I2luY2x1ZGUgPGlvc3RyZWFtPg==",
					Language: models.LanguageCpp,
					Compiler: models.CompilerGCC13,
				},
				Status:    models.StatusQueued,
				CreatedAt: time.Now(),
			}

			go func(j models.CompilationJob) {
				server.processJob(j)
				done <- j.ID
			}(job)
		}

		// Wait for all jobs
		completedJobs := make(map[string]bool)
		for i := 0; i < jobCount; i++ {
			jobID := <-done
			completedJobs[jobID] = true
		}

		// Verify all completed successfully
		assert.Len(t, completedJobs, jobCount)
		for i := 0; i < jobCount; i++ {
			jobID := fmt.Sprintf("job-%d", i)
			job, exists := server.jobs.Get(jobID)
			require.True(t, exists)
			assert.Equal(t, models.StatusCompleted, job.Status)

			result, hasResult := server.jobs.GetResult(jobID)
			require.True(t, hasResult)
			assert.True(t, result.Success)
		}
	})
}

// TestJobLifecycle verifies job status transitions from queued to completed.
func TestJobLifecycle(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		server := &Server{
			compiler: &mockCompiler{compileDelay: 100 * time.Millisecond, shouldFail: false},
			jobs:     newJobStore(),
		}

		job := models.CompilationJob{
			ID: "lifecycle-job",
			Request: models.CompilationRequest{
				Code:     "I2luY2x1ZGUgPGlvc3RyZWFtPg==",
				Language: models.LanguageCpp,
				Compiler: models.CompilerGCC13,
			},
			Status:    models.StatusQueued,
			CreatedAt: time.Now(),
		}

		// Store initial job
		server.jobs.Store(job)

		// Verify initial state
		initialJob, _ := server.jobs.Get(job.ID)
		assert.Equal(t, models.StatusQueued, initialJob.Status)
		assert.Nil(t, initialJob.StartedAt)
		assert.Nil(t, initialJob.CompletedAt)

		// Process job
		done := make(chan struct{})
		go func() {
			server.processJob(job)
			close(done)
		}()
		<-done

		// Verify final state
		finalJob, exists := server.jobs.Get(job.ID)
		require.True(t, exists)
		assert.Equal(t, models.StatusCompleted, finalJob.Status)
		assert.NotNil(t, finalJob.StartedAt)
		assert.NotNil(t, finalJob.CompletedAt)
		assert.True(t, finalJob.CompletedAt.After(*finalJob.StartedAt))
	})
}
