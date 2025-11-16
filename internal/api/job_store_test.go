//go:build go1.25

package api

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stlpine/will-it-compile/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestJobStore_ConcurrentAccess verifies thread-safe concurrent operations.
// Run with: go test -race
func TestJobStore_ConcurrentAccess(t *testing.T) {
	store := newJobStore()
	const numWorkers = 100
	const opsPerWorker = 100

	var wg sync.WaitGroup
	wg.Add(numWorkers)

	// Launch multiple goroutines performing concurrent operations
	for i := 0; i < numWorkers; i++ {
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < opsPerWorker; j++ {
				jobID := fmt.Sprintf("worker-%d-job-%d", workerID, j)

				// Store job
				job := models.CompilationJob{
					ID:        jobID,
					Status:    models.StatusQueued,
					CreatedAt: time.Now(),
				}
				store.Store(job)

				// Read job (might not exist yet from other workers)
				_, _ = store.Get(jobID)

				// Update job status
				job.Status = models.StatusProcessing
				store.Store(job)

				// Store result
				result := models.CompilationResult{
					Success:  true,
					Compiled: true,
					ExitCode: 0,
				}
				store.StoreResult(jobID, result)

				// Read result
				_, _ = store.GetResult(jobID)
			}
		}(i)
	}

	wg.Wait()

	// Verify all jobs were stored
	for i := 0; i < numWorkers; i++ {
		for j := 0; j < opsPerWorker; j++ {
			jobID := fmt.Sprintf("worker-%d-job-%d", i, j)
			job, exists := store.Get(jobID)
			require.True(t, exists, "Job %s should exist", jobID)
			assert.Equal(t, models.StatusProcessing, job.Status)

			result, hasResult := store.GetResult(jobID)
			require.True(t, hasResult, "Result for %s should exist", jobID)
			assert.True(t, result.Success)
		}
	}
}

// TestJobStore_ReadWriteMix tests mixed read/write operations.
func TestJobStore_ReadWriteMix(t *testing.T) {
	store := newJobStore()
	const numReaders = 50
	const numWriters = 10
	const duration = 100 * time.Millisecond

	done := make(chan struct{})
	time.AfterFunc(duration, func() { close(done) })

	var wg sync.WaitGroup

	// Start readers
	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
					store.Get(fmt.Sprintf("job-%d", id%numWriters))
					store.GetResult(fmt.Sprintf("job-%d", id%numWriters))
				}
			}
		}(i)
	}

	// Start writers
	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			jobID := fmt.Sprintf("job-%d", id)
			counter := 0
			for {
				select {
				case <-done:
					return
				default:
					job := models.CompilationJob{
						ID:        jobID,
						Status:    models.StatusProcessing,
						CreatedAt: time.Now(),
					}
					store.Store(job)

					result := models.CompilationResult{
						Success:  counter%2 == 0,
						Compiled: true,
						ExitCode: counter,
					}
					store.StoreResult(jobID, result)
					counter++
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify final state is consistent
	for i := 0; i < numWriters; i++ {
		jobID := fmt.Sprintf("job-%d", i)
		job, exists := store.Get(jobID)
		assert.True(t, exists)
		assert.Equal(t, jobID, job.ID)

		_, hasResult := store.GetResult(jobID)
		assert.True(t, hasResult)
	}
}

// TestJobStore_EdgeCases tests boundary conditions.
func TestJobStore_EdgeCases(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "nonexistent job",
			test: func(t *testing.T) {
				store := newJobStore()
				_, exists := store.Get("nonexistent")
				assert.False(t, exists)
			},
		},
		{
			name: "nonexistent result",
			test: func(t *testing.T) {
				store := newJobStore()
				_, exists := store.GetResult("nonexistent")
				assert.False(t, exists)
			},
		},
		{
			name: "empty job id",
			test: func(t *testing.T) {
				store := newJobStore()
				job := models.CompilationJob{ID: "", Status: models.StatusQueued}
				store.Store(job)
				retrieved, exists := store.Get("")
				assert.True(t, exists)
				assert.Equal(t, "", retrieved.ID)
			},
		},
		{
			name: "overwrite job",
			test: func(t *testing.T) {
				store := newJobStore()
				job1 := models.CompilationJob{ID: "test", Status: models.StatusQueued}
				job2 := models.CompilationJob{ID: "test", Status: models.StatusCompleted}

				store.Store(job1)
				store.Store(job2)

				retrieved, _ := store.Get("test")
				assert.Equal(t, models.StatusCompleted, retrieved.Status)
			},
		},
		{
			name: "result before job",
			test: func(t *testing.T) {
				store := newJobStore()
				result := models.CompilationResult{Success: true}
				store.StoreResult("orphan", result)

				retrieved, exists := store.GetResult("orphan")
				assert.True(t, exists)
				assert.True(t, retrieved.Success)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.test(t)
		})
	}
}

// TestJobStore_UpdateSequence verifies job lifecycle updates work correctly.
func TestJobStore_UpdateSequence(t *testing.T) {
	store := newJobStore()

	job := models.CompilationJob{
		ID:        "lifecycle-test",
		Status:    models.StatusQueued,
		CreatedAt: time.Now(),
	}

	// Initial store
	store.Store(job)
	retrieved, exists := store.Get(job.ID)
	require.True(t, exists)
	assert.Equal(t, models.StatusQueued, retrieved.Status)
	assert.Nil(t, retrieved.StartedAt)

	// Update to processing
	started := time.Now()
	job.Status = models.StatusProcessing
	job.StartedAt = &started
	store.Store(job)

	retrieved, exists = store.Get(job.ID)
	require.True(t, exists)
	assert.Equal(t, models.StatusProcessing, retrieved.Status)
	assert.NotNil(t, retrieved.StartedAt)

	// Add result
	result := models.CompilationResult{
		Success:  true,
		Compiled: true,
		ExitCode: 0,
		Stdout:   "Success",
	}
	store.StoreResult(job.ID, result)

	// Update to completed
	completed := time.Now()
	job.Status = models.StatusCompleted
	job.CompletedAt = &completed
	store.Store(job)

	retrieved, exists = store.Get(job.ID)
	require.True(t, exists)
	assert.Equal(t, models.StatusCompleted, retrieved.Status)
	assert.NotNil(t, retrieved.CompletedAt)

	retrievedResult, hasResult := store.GetResult(job.ID)
	require.True(t, hasResult)
	assert.Equal(t, "Success", retrievedResult.Stdout)
}
