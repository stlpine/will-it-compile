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
// to simulate compilation delay. With synctest, these delays happen instantly.
type mockCompiler struct {
	compileDelay time.Duration
	shouldFail   bool
}

func (m *mockCompiler) Compile(ctx context.Context, job models.CompilationJob) models.CompilationResult {
	// Simulate compilation time - with synctest, this happens instantly!
	time.Sleep(m.compileDelay)

	if m.shouldFail {
		return models.CompilationResult{
			Success:  false,
			Compiled: false,
			Error:    "mock compilation error",
			ExitCode: 1,
			Stderr:   "error: mock compilation failed",
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

// TestProcessJob_WithSynctest demonstrates testing async job processing
// with virtualized time. This test would take 2 seconds without synctest,
// but runs instantly with it.
func TestProcessJob_WithSynctest(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// Create server with mock compiler that takes 2 seconds
		server := &Server{
			compiler: &mockCompiler{compileDelay: 2 * time.Second, shouldFail: false},
			jobs:     newJobStore(),
		}

		// Create a job
		job := models.CompilationJob{
			ID: "test-job-1",
			Request: models.CompilationRequest{
				Code:     "I2luY2x1ZGUgPGlvc3RyZWFtPg==",
				Language: models.LanguageCpp,
				Compiler: models.CompilerGCC13,
			},
			Status:    models.StatusQueued,
			CreatedAt: time.Now(),
		}

		// Start async processing
		done := make(chan struct{})
		go func() {
			server.processJob(job)
			close(done)
		}()

		// Wait for goroutine to complete
		// With synctest, the 2-second sleep happens instantly!
		<-done

		// Verify job was processed
		processedJob, exists := server.jobs.Get(job.ID)
		require.True(t, exists, "Job should exist")
		assert.Equal(t, models.StatusCompleted, processedJob.Status, "Job should be completed")

		// Verify result
		result, hasResult := server.jobs.GetResult(job.ID)
		require.True(t, hasResult, "Job should have result")
		assert.True(t, result.Success, "Compilation should succeed")
		assert.True(t, result.Compiled, "Code should compile")
		assert.Equal(t, 0, result.ExitCode, "Exit code should be 0")
	})
}

// TestProcessJob_Failure_WithSynctest tests failed compilation with synctest
func TestProcessJob_Failure_WithSynctest(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// Create server with failing mock compiler
		server := &Server{
			compiler: &mockCompiler{compileDelay: 500 * time.Millisecond, shouldFail: true},
			jobs:     newJobStore(),
		}

		job := models.CompilationJob{
			ID: "test-job-2",
			Request: models.CompilationRequest{
				Code:     "aW52YWxpZCBjb2Rl",
				Language: models.LanguageCpp,
				Compiler: models.CompilerGCC13,
			},
			Status:    models.StatusQueued,
			CreatedAt: time.Now(),
		}

		// Start async processing
		done := make(chan struct{})
		go func() {
			server.processJob(job)
			close(done)
		}()

		// Wait for completion - instant with synctest!
		<-done

		// Verify job failed
		processedJob, exists := server.jobs.Get(job.ID)
		require.True(t, exists, "Job should exist")
		assert.Equal(t, models.StatusFailed, processedJob.Status, "Job should be failed")

		// Verify error result
		result, hasResult := server.jobs.GetResult(job.ID)
		require.True(t, hasResult, "Job should have result")
		assert.False(t, result.Success, "Compilation should fail")
		assert.False(t, result.Compiled, "Code should not compile")
		assert.NotEqual(t, 0, result.ExitCode, "Exit code should be non-zero")
		assert.NotEmpty(t, result.Error, "Should have error message")
	})
}

// TestMultipleJobsConcurrently_WithSynctest demonstrates testing multiple
// concurrent jobs with deterministic execution order.
func TestMultipleJobsConcurrently_WithSynctest(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		server := &Server{
			compiler: &mockCompiler{compileDelay: 1 * time.Second, shouldFail: false},
			jobs:     newJobStore(),
		}

		// Process 5 jobs concurrently
		jobCount := 5
		jobIDs := make([]string, jobCount)
		done := make(chan struct{}, jobCount)

		for i := 0; i < jobCount; i++ {
			jobID := models.CompilationJob{
				ID: fmt.Sprintf("concurrent-job-%d", i),
				Request: models.CompilationRequest{
					Code:     "I2luY2x1ZGUgPGlvc3RyZWFtPg==",
					Language: models.LanguageCpp,
					Compiler: models.CompilerGCC13,
				},
				Status:    models.StatusQueued,
				CreatedAt: time.Now(),
			}
			jobIDs[i] = jobID.ID

			go func(job models.CompilationJob) {
				server.processJob(job)
				done <- struct{}{}
			}(jobID)
		}

		// Wait for all goroutines to complete
		// Without synctest: would take 5 seconds (or 1 second if truly concurrent)
		// With synctest: instant!
		for i := 0; i < jobCount; i++ {
			<-done
		}

		// Verify all jobs completed
		for _, id := range jobIDs {
			job, exists := server.jobs.Get(id)
			require.True(t, exists, "Job %s should exist", id)
			assert.Equal(t, models.StatusCompleted, job.Status, "Job %s should be completed", id)

			result, hasResult := server.jobs.GetResult(id)
			require.True(t, hasResult, "Job %s should have result", id)
			assert.True(t, result.Success, "Job %s should succeed", id)
		}
	})
}

// TestJobStatusTransitions_WithSynctest verifies status transitions during processing
func TestJobStatusTransitions_WithSynctest(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		server := &Server{
			compiler: &mockCompiler{compileDelay: 100 * time.Millisecond, shouldFail: false},
			jobs:     newJobStore(),
		}

		job := models.CompilationJob{
			ID: "status-test-job",
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

		// Verify initial status
		initialJob, _ := server.jobs.Get(job.ID)
		assert.Equal(t, models.StatusQueued, initialJob.Status, "Should start as queued")
		assert.Nil(t, initialJob.StartedAt, "Should not have started yet")
		assert.Nil(t, initialJob.CompletedAt, "Should not be completed yet")

		// Start processing
		done := make(chan struct{})
		go func() {
			server.processJob(job)
			close(done)
		}()

		// Wait for completion
		<-done

		// Verify final status
		finalJob, exists := server.jobs.Get(job.ID)
		require.True(t, exists, "Job should exist")
		assert.Equal(t, models.StatusCompleted, finalJob.Status, "Should be completed")
		assert.NotNil(t, finalJob.StartedAt, "Should have start time")
		assert.NotNil(t, finalJob.CompletedAt, "Should have completion time")
		assert.True(t, finalJob.CompletedAt.After(*finalJob.StartedAt), "Completion should be after start")
	})
}
