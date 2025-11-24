package redis

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stlpine/will-it-compile/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestStore(t *testing.T) (*Store, *miniredis.Miniredis) {
	// Create miniredis server
	mr, err := miniredis.Run()
	require.NoError(t, err)

	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	// Create store
	store := NewStoreWithClient(client, 24*time.Hour)

	return store, mr
}

func TestRedisStore_StoreAndGet(t *testing.T) {
	store, mr := setupTestStore(t)
	defer mr.Close()
	defer store.Close() //nolint:errcheck // test cleanup

	// Create test job
	job := models.CompilationJob{
		ID: "test-job-1",
		Request: models.CompilationRequest{
			Code:     "int main() { return 0; }",
			Language: "cpp",
			Compiler: "gcc-13",
		},
		Status:    models.StatusQueued,
		CreatedAt: time.Now(),
	}

	// Store job
	err := store.Store(job)
	assert.NoError(t, err)

	// Retrieve job
	retrieved, found := store.Get("test-job-1")
	assert.True(t, found)
	assert.Equal(t, job.ID, retrieved.ID)
	assert.Equal(t, job.Request.Code, retrieved.Request.Code)
	assert.Equal(t, job.Status, retrieved.Status)
}

func TestRedisStore_GetNonExistent(t *testing.T) {
	store, mr := setupTestStore(t)
	defer mr.Close()
	defer store.Close() //nolint:errcheck // test cleanup

	// Try to get non-existent job
	_, found := store.Get("non-existent")
	assert.False(t, found)
}

func TestRedisStore_StoreResult(t *testing.T) {
	store, mr := setupTestStore(t)
	defer mr.Close()
	defer store.Close() //nolint:errcheck // test cleanup

	// Create test result
	result := models.CompilationResult{
		JobID:    "test-job-1",
		Success:  true,
		Compiled: true,
		Stdout:   "Compilation successful",
		Stderr:   "",
		ExitCode: 0,
		Duration: 1234 * time.Millisecond,
		Error:    "",
	}

	// Store result
	err := store.StoreResult("test-job-1", result)
	assert.NoError(t, err)

	// Retrieve result
	retrieved, found := store.GetResult("test-job-1")
	assert.True(t, found)
	assert.Equal(t, result.JobID, retrieved.JobID)
	assert.Equal(t, result.Success, retrieved.Success)
	assert.Equal(t, result.Compiled, retrieved.Compiled)
	assert.Equal(t, result.Stdout, retrieved.Stdout)
	assert.Equal(t, result.ExitCode, retrieved.ExitCode)
	assert.Equal(t, result.Duration, retrieved.Duration)
	assert.Equal(t, result.Error, retrieved.Error)
}

func TestRedisStore_UpdateJobStatus(t *testing.T) {
	store, mr := setupTestStore(t)
	defer mr.Close()
	defer store.Close() //nolint:errcheck // test cleanup

	// Create initial job
	job := models.CompilationJob{
		ID: "test-job-1",
		Request: models.CompilationRequest{
			Code:     "int main() { return 0; }",
			Language: "cpp",
			Compiler: "gcc-13",
		},
		Status:    models.StatusQueued,
		CreatedAt: time.Now(),
	}

	// Store job
	err := store.Store(job)
	require.NoError(t, err)

	// Update status to processing
	job.Status = models.StatusProcessing
	startedAt := time.Now()
	job.StartedAt = &startedAt

	err = store.Store(job)
	require.NoError(t, err)

	// Retrieve and verify
	retrieved, found := store.Get("test-job-1")
	assert.True(t, found)
	assert.Equal(t, models.StatusProcessing, retrieved.Status)
	assert.NotNil(t, retrieved.StartedAt)

	// Update to completed
	job.Status = models.StatusCompleted
	completedAt := time.Now()
	job.CompletedAt = &completedAt

	err = store.Store(job)
	require.NoError(t, err)

	// Retrieve final state
	retrieved, found = store.Get("test-job-1")
	assert.True(t, found)
	assert.Equal(t, models.StatusCompleted, retrieved.Status)
	assert.NotNil(t, retrieved.CompletedAt)
}

func TestRedisStore_TTL(t *testing.T) {
	store, mr := setupTestStore(t)
	defer mr.Close()
	defer store.Close() //nolint:errcheck // test cleanup

	// Create job
	job := models.CompilationJob{
		ID: "test-job-ttl",
		Request: models.CompilationRequest{
			Code:     "int main() { return 0; }",
			Language: "cpp",
			Compiler: "gcc-13",
		},
		Status:    models.StatusQueued,
		CreatedAt: time.Now(),
	}

	// Store job
	err := store.Store(job)
	require.NoError(t, err)

	// Check that TTL is set
	ttl := mr.TTL("job:test-job-ttl")
	assert.Greater(t, ttl, time.Duration(0))
	assert.LessOrEqual(t, ttl, 24*time.Hour)
}

func TestRedisStore_FailedCompilation(t *testing.T) {
	store, mr := setupTestStore(t)
	defer mr.Close()
	defer store.Close() //nolint:errcheck // test cleanup

	// Create job that will fail
	job := models.CompilationJob{
		ID: "test-job-failed",
		Request: models.CompilationRequest{
			Code:     "invalid code",
			Language: "cpp",
			Compiler: "gcc-13",
		},
		Status:    models.StatusFailed,
		CreatedAt: time.Now(),
	}

	err := store.Store(job)
	require.NoError(t, err)

	// Store failure result
	result := models.CompilationResult{
		Success:  false,
		Compiled: false,
		Stdout:   "",
		Stderr:   "error: expected ';' before '}' token",
		ExitCode: 1,
		Duration: 500 * time.Millisecond,
	}

	err = store.StoreResult("test-job-failed", result)
	require.NoError(t, err)

	// Retrieve and verify
	retrievedJob, found := store.Get("test-job-failed")
	assert.True(t, found)
	assert.Equal(t, models.StatusFailed, retrievedJob.Status)

	retrievedResult, found := store.GetResult("test-job-failed")
	assert.True(t, found)
	assert.False(t, retrievedResult.Success)
	assert.Equal(t, 1, retrievedResult.ExitCode)
	assert.Contains(t, retrievedResult.Stderr, "error")
}
