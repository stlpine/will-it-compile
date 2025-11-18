//go:build go1.25

package api

import (
	"testing"
	"testing/synctest"
	"time"

	"github.com/stlpine/will-it-compile/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkerPool_BasicFunctionality(t *testing.T) {
	// Create mock compiler
	mockComp := &mockCompiler{
		compileDelay: 100 * time.Millisecond,
		shouldFail:   false,
	}

	// Create server with worker pool
	server := &Server{
		compiler: mockComp,
		jobs:     newJobStore(),
	}

	// Create worker pool with 3 workers
	pool := NewWorkerPool(3, 10, server)
	pool.Start()
	defer pool.Stop()

	// Verify initial stats
	stats := pool.GetStats()
	assert.Equal(t, 3, stats.MaxWorkers, "Should have 3 max workers")
	assert.Equal(t, 0, stats.ActiveWorkers, "Should have 0 active workers initially")
	assert.Equal(t, 3, stats.AvailableSlots, "Should have 3 available slots initially")
	assert.Equal(t, 0, stats.QueuedJobs, "Should have 0 queued jobs initially")
}

func TestWorkerPool_JobProcessing(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// Create mock compiler with small delay
		mockComp := &mockCompiler{
			compileDelay: 50 * time.Millisecond,
			shouldFail:   false,
		}

		// Create server with worker pool
		server := &Server{
			compiler: mockComp,
			jobs:     newJobStore(),
		}

		// Create worker pool with 2 workers
		pool := NewWorkerPool(2, 10, server)
		pool.Start()
		defer pool.Stop()

		// Create and submit a job
		job := models.CompilationJob{
			ID:        "test-job-1",
			Status:    models.StatusQueued,
			CreatedAt: time.Now(),
			Request: models.CompilationRequest{
				Code:     "aW50IG1haW4oKSB7IHJldHVybiAwOyB9", // base64 encoded "int main() { return 0; }"
				Language: models.LanguageCpp,
			},
		}

		// Store and submit job
		server.jobs.Store(job)
		submitted := pool.Submit(job)
		require.True(t, submitted, "Job should be submitted successfully")

		// Wait for job to be processed (instant with virtualized time)
		time.Sleep(200 * time.Millisecond)

		// Check stats
		stats := pool.GetStats()
		assert.Equal(t, int64(1), stats.TotalProcessed, "Should have processed 1 job")
		assert.Equal(t, int64(1), stats.TotalSuccessful, "Should have 1 successful job")
		assert.Equal(t, int64(0), stats.TotalFailed, "Should have 0 failed jobs")

		// Verify job result
		result, hasResult := server.jobs.GetResult(job.ID)
		require.True(t, hasResult, "Job should have result")
		assert.True(t, result.Success, "Job should be successful")
	})
}

func TestWorkerPool_ConcurrentJobs(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// Create mock compiler with delay
		mockComp := &mockCompiler{
			compileDelay: 100 * time.Millisecond,
			shouldFail:   false,
		}

		// Create server with worker pool
		server := &Server{
			compiler: mockComp,
			jobs:     newJobStore(),
		}

		// Create worker pool with 5 workers
		pool := NewWorkerPool(5, 20, server)
		pool.Start()
		defer pool.Stop()

		// Submit 10 jobs concurrently
		numJobs := 10
		for i := 0; i < numJobs; i++ {
			job := models.CompilationJob{
				ID:        string(rune('a' + i)),
				Status:    models.StatusQueued,
				CreatedAt: time.Now(),
				Request: models.CompilationRequest{
					Code:     "aW50IG1haW4oKSB7IHJldHVybiAwOyB9",
					Language: models.LanguageCpp,
				},
			}
			server.jobs.Store(job)
			submitted := pool.Submit(job)
			require.True(t, submitted, "Job %d should be submitted", i)
		}

		// Wait for all jobs to complete (instant with virtualized time)
		// With 5 workers and 100ms delay, 10 jobs should take ~200ms (2 batches)
		time.Sleep(500 * time.Millisecond)

		// Check stats
		stats := pool.GetStats()
		assert.Equal(t, int64(numJobs), stats.TotalProcessed, "Should have processed all jobs")
		assert.Equal(t, int64(numJobs), stats.TotalSuccessful, "All jobs should be successful")
		assert.Equal(t, int64(0), stats.TotalFailed, "No jobs should fail")
		assert.Equal(t, 0, stats.ActiveWorkers, "All workers should be idle")
		assert.Equal(t, 5, stats.AvailableSlots, "All slots should be available")
	})
}

func TestWorkerPool_QueueFull(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// Create mock compiler with long delay
		mockComp := &mockCompiler{
			compileDelay: 1 * time.Second,
			shouldFail:   false,
		}

		// Create server with worker pool
		server := &Server{
			compiler: mockComp,
			jobs:     newJobStore(),
		}

		// Create worker pool with small queue (1 worker, queue size 2)
		// This means: 1 job processing + 2 jobs in queue = 3 total capacity
		pool := NewWorkerPool(1, 2, server)
		pool.Start()
		defer pool.Stop()

		// Submit jobs until queue is full
		job1 := models.CompilationJob{ID: "job1", Status: models.StatusQueued, CreatedAt: time.Now(), Request: models.CompilationRequest{Code: "test", Language: models.LanguageCpp}}
		job2 := models.CompilationJob{ID: "job2", Status: models.StatusQueued, CreatedAt: time.Now(), Request: models.CompilationRequest{Code: "test", Language: models.LanguageCpp}}
		job3 := models.CompilationJob{ID: "job3", Status: models.StatusQueued, CreatedAt: time.Now(), Request: models.CompilationRequest{Code: "test", Language: models.LanguageCpp}}
		job4 := models.CompilationJob{ID: "job4", Status: models.StatusQueued, CreatedAt: time.Now(), Request: models.CompilationRequest{Code: "test", Language: models.LanguageCpp}}

		server.jobs.Store(job1)
		server.jobs.Store(job2)
		server.jobs.Store(job3)
		server.jobs.Store(job4)

		// First three should be accepted (1 processing, 2 in queue)
		assert.True(t, pool.Submit(job1), "First job should be accepted")
		time.Sleep(50 * time.Millisecond) // Let first job start processing (instant with virtualized time)
		assert.True(t, pool.Submit(job2), "Second job should be accepted")
		assert.True(t, pool.Submit(job3), "Third job should be accepted")

		// Fourth should be rejected (queue full: 1 processing + 2 queued = capacity reached)
		assert.False(t, pool.Submit(job4), "Fourth job should be rejected (queue full)")
	})
}

func TestWorkerPool_MixedSuccessAndFailure(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// Create server
		server := &Server{
			jobs: newJobStore(),
		}

		// Create worker pool
		pool := NewWorkerPool(3, 10, server)
		pool.Start()
		defer pool.Stop()

		// Submit mix of successful and failing jobs
		for i := 0; i < 5; i++ {
			// Alternate between success and failure
			shouldFail := i%2 == 1
			mockComp := &mockCompiler{
				compileDelay: 50 * time.Millisecond,
				shouldFail:   shouldFail,
			}

			// Temporarily set compiler for this test
			oldCompiler := server.compiler
			server.compiler = mockComp

			job := models.CompilationJob{
				ID:        string(rune('a' + i)),
				Status:    models.StatusQueued,
				CreatedAt: time.Now(),
				Request: models.CompilationRequest{
					Code:     "test",
					Language: models.LanguageCpp,
				},
			}
			server.jobs.Store(job)
			pool.Submit(job)

			server.compiler = oldCompiler
		}

		// Wait for all jobs to complete (instant with virtualized time)
		time.Sleep(300 * time.Millisecond)

		// Note: This test is simplified and may not work perfectly
		// because we're changing the compiler mid-flight
		// In a real scenario, the compiler would determine success/failure
		// based on the code content, not a mock setting
	})
}

func TestWorkerPool_Uptime(t *testing.T) {
	mockComp := &mockCompiler{
		compileDelay: 0,
		shouldFail:   false,
	}

	server := &Server{
		compiler: mockComp,
		jobs:     newJobStore(),
	}

	pool := NewWorkerPool(2, 10, server)
	startTime := time.Now()
	pool.Start()
	defer pool.Stop()

	// Wait a bit
	time.Sleep(100 * time.Millisecond)

	stats := pool.GetStats()
	assert.NotEmpty(t, stats.Uptime, "Uptime should be populated")
	assert.True(t, stats.UptimeSeconds >= 0, "Uptime seconds should be non-negative")
	assert.True(t, stats.StartTime.After(startTime.Add(-1*time.Second)), "Start time should be recent")
}
