package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stlpine/will-it-compile/internal/config"
	"github.com/stlpine/will-it-compile/internal/storage"
	redisstore "github.com/stlpine/will-it-compile/internal/storage/redis"
	"github.com/stlpine/will-it-compile/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// RedisIntegrationSuite tests Redis storage with a real Redis instance
type RedisIntegrationSuite struct {
	suite.Suite
	redisAddr string
	client    *redis.Client
	store     storage.JobStore
}

// SetupSuite runs once before all tests in the suite
func (s *RedisIntegrationSuite) SetupSuite() {
	// Get Redis address from environment variable
	s.redisAddr = os.Getenv("REDIS_ADDR")
	if s.redisAddr == "" {
		s.redisAddr = "localhost:6379"
	}

	// Try to connect to Redis
	s.client = redis.NewClient(&redis.Options{
		Addr: s.redisAddr,
		DB:   1, // Use DB 1 for tests to avoid conflicts
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := s.client.Ping(ctx).Err()
	if err != nil {
		s.T().Skipf("Redis not available at %s: %v. Skipping Redis integration tests.", s.redisAddr, err)
	}

	s.T().Logf("Connected to Redis at %s", s.redisAddr)
}

// SetupTest runs before each test
func (s *RedisIntegrationSuite) SetupTest() {
	// Create a fresh store for each test
	cfg := config.RedisConfig{
		Addr:         s.redisAddr,
		Password:     os.Getenv("REDIS_PASSWORD"),
		DB:           1, // Use test database
		PoolSize:     10,
		MaxRetries:   3,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		JobTTL:       1 * time.Hour, // Short TTL for tests
	}

	var err error
	s.store, err = redisstore.NewStore(cfg)
	require.NoError(s.T(), err)

	// Clean up test database before each test
	ctx := context.Background()
	s.client.FlushDB(ctx)
}

// TearDownTest runs after each test
func (s *RedisIntegrationSuite) TearDownTest() {
	if s.store != nil {
		s.store.Close() //nolint:errcheck // test cleanup
	}
}

// TearDownSuite runs once after all tests in the suite
func (s *RedisIntegrationSuite) TearDownSuite() {
	if s.client != nil {
		// Clean up test database
		ctx := context.Background()
		s.client.FlushDB(ctx)
		s.client.Close() //nolint:errcheck // test cleanup
	}
}

// TestRedisJobPersistence verifies that jobs persist in Redis
func (s *RedisIntegrationSuite) TestRedisJobPersistence() {
	job := models.CompilationJob{
		ID: "persist-test-1",
		Request: models.CompilationRequest{
			SourceCode: "int main() { return 0; }",
			Language:   "cpp",
			Compiler:   "gcc-13",
		},
		Status:    models.StatusQueued,
		CreatedAt: time.Now(),
	}

	// Store job
	err := s.store.Store(job)
	assert.NoError(s.T(), err)

	// Verify job exists in Redis
	ctx := context.Background()
	exists := s.client.Exists(ctx, "job:persist-test-1").Val()
	assert.Equal(s.T(), int64(1), exists, "Job should exist in Redis")

	// Retrieve job
	retrieved, found := s.store.Get("persist-test-1")
	assert.True(s.T(), found)
	assert.Equal(s.T(), job.ID, retrieved.ID)
	assert.Equal(s.T(), job.Request.SourceCode, retrieved.Request.SourceCode)
}

// TestRedisJobLifecycle tests the complete job lifecycle
func (s *RedisIntegrationSuite) TestRedisJobLifecycle() {
	// Create job
	job := models.CompilationJob{
		ID: "lifecycle-test-1",
		Request: models.CompilationRequest{
			SourceCode: "int main() { return 0; }",
			Language:   "cpp",
			Compiler:   "gcc-13",
		},
		Status:    models.StatusQueued,
		CreatedAt: time.Now(),
	}

	// Store initial job
	err := s.store.Store(job)
	require.NoError(s.T(), err)

	// Update to processing
	job.Status = models.StatusProcessing
	startedAt := time.Now()
	job.StartedAt = &startedAt
	err = s.store.Store(job)
	require.NoError(s.T(), err)

	// Verify status update
	retrieved, found := s.store.Get("lifecycle-test-1")
	assert.True(s.T(), found)
	assert.Equal(s.T(), models.StatusProcessing, retrieved.Status)
	assert.NotNil(s.T(), retrieved.StartedAt)

	// Store result
	result := models.CompilationResult{
		Success:  true,
		Compiled: true,
		Stdout:   "Compilation successful",
		Stderr:   "",
		ExitCode: 0,
		Duration: 1234 * time.Millisecond,
	}
	err = s.store.StoreResult("lifecycle-test-1", result)
	require.NoError(s.T(), err)

	// Update to completed
	job.Status = models.StatusCompleted
	completedAt := time.Now()
	job.CompletedAt = &completedAt
	err = s.store.Store(job)
	require.NoError(s.T(), err)

	// Verify final state
	finalJob, found := s.store.Get("lifecycle-test-1")
	assert.True(s.T(), found)
	assert.Equal(s.T(), models.StatusCompleted, finalJob.Status)
	assert.NotNil(s.T(), finalJob.CompletedAt)

	finalResult, found := s.store.GetResult("lifecycle-test-1")
	assert.True(s.T(), found)
	assert.True(s.T(), finalResult.Success)
	assert.Equal(s.T(), 0, finalResult.ExitCode)
}

// TestRedisMultipleJobs tests storing and retrieving multiple jobs
func (s *RedisIntegrationSuite) TestRedisMultipleJobs() {
	numJobs := 10

	// Store multiple jobs
	for i := 0; i < numJobs; i++ {
		job := models.CompilationJob{
			ID: fmt.Sprintf("multi-job-%d", i),
			Request: models.CompilationRequest{
				SourceCode: fmt.Sprintf("int main() { return %d; }", i),
				Language:   "cpp",
				Compiler:   "gcc-13",
			},
			Status:    models.StatusQueued,
			CreatedAt: time.Now(),
		}
		err := s.store.Store(job)
		assert.NoError(s.T(), err)
	}

	// Verify all jobs exist
	ctx := context.Background()
	keys := s.client.Keys(ctx, "job:multi-job-*").Val()
	assert.Len(s.T(), keys, numJobs, "All jobs should be in Redis")

	// Retrieve and verify each job
	for i := 0; i < numJobs; i++ {
		jobID := fmt.Sprintf("multi-job-%d", i)
		job, found := s.store.Get(jobID)
		assert.True(s.T(), found, "Job %s should exist", jobID)
		assert.Equal(s.T(), jobID, job.ID)
	}
}

// TestRedisTTL verifies that TTL is properly set on jobs
func (s *RedisIntegrationSuite) TestRedisTTL() {
	job := models.CompilationJob{
		ID: "ttl-test-1",
		Request: models.CompilationRequest{
			SourceCode: "int main() { return 0; }",
			Language:   "cpp",
			Compiler:   "gcc-13",
		},
		Status:    models.StatusQueued,
		CreatedAt: time.Now(),
	}

	err := s.store.Store(job)
	require.NoError(s.T(), err)

	// Check TTL
	ctx := context.Background()
	ttl := s.client.TTL(ctx, "job:ttl-test-1").Val()

	// TTL should be > 0 and <= 1 hour (our test TTL)
	assert.Greater(s.T(), ttl, time.Duration(0), "TTL should be set")
	assert.LessOrEqual(s.T(), ttl, 1*time.Hour, "TTL should not exceed configured value")
}

// TestRedisFailedCompilation tests storing failed compilation results
func (s *RedisIntegrationSuite) TestRedisFailedCompilation() {
	job := models.CompilationJob{
		ID: "failed-test-1",
		Request: models.CompilationRequest{
			SourceCode: "invalid code",
			Language:   "cpp",
			Compiler:   "gcc-13",
		},
		Status:    models.StatusFailed,
		CreatedAt: time.Now(),
	}

	err := s.store.Store(job)
	require.NoError(s.T(), err)

	// Store failure result
	result := models.CompilationResult{
		Success:  false,
		Compiled: false,
		Stdout:   "",
		Stderr:   "error: expected ';' before '}' token",
		ExitCode: 1,
		Duration: 500 * time.Millisecond,
	}

	err = s.store.StoreResult("failed-test-1", result)
	require.NoError(s.T(), err)

	// Verify both job and result
	retrievedJob, found := s.store.Get("failed-test-1")
	assert.True(s.T(), found)
	assert.Equal(s.T(), models.StatusFailed, retrievedJob.Status)

	retrievedResult, found := s.store.GetResult("failed-test-1")
	assert.True(s.T(), found)
	assert.False(s.T(), retrievedResult.Success)
	assert.Equal(s.T(), 1, retrievedResult.ExitCode)
	assert.Contains(s.T(), retrievedResult.Stderr, "error")
}

// TestRedisConcurrentAccess tests concurrent reads and writes
func (s *RedisIntegrationSuite) TestRedisConcurrentAccess() {
	const numGoroutines = 20
	const jobsPerGoroutine = 10

	done := make(chan bool, numGoroutines)
	errors := make(chan error, numGoroutines*jobsPerGoroutine)

	// Spawn multiple goroutines to write jobs concurrently
	for g := 0; g < numGoroutines; g++ {
		goroutineID := g
		go func() {
			for i := 0; i < jobsPerGoroutine; i++ {
				job := models.CompilationJob{
					ID: fmt.Sprintf("concurrent-%d-%d", goroutineID, i),
					Request: models.CompilationRequest{
						SourceCode: fmt.Sprintf("int main() { return %d; }", i),
						Language:   "cpp",
						Compiler:   "gcc-13",
					},
					Status:    models.StatusQueued,
					CreatedAt: time.Now(),
				}

				if err := s.store.Store(job); err != nil {
					errors <- err
				}
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
	close(errors)

	// Check for errors
	var errCount int
	for err := range errors {
		s.T().Logf("Error during concurrent access: %v", err)
		errCount++
	}
	assert.Equal(s.T(), 0, errCount, "No errors should occur during concurrent access")

	// Verify all jobs were stored
	ctx := context.Background()
	keys := s.client.Keys(ctx, "job:concurrent-*").Val()
	expectedJobs := numGoroutines * jobsPerGoroutine
	assert.Equal(s.T(), expectedJobs, len(keys), "All concurrent jobs should be stored")
}

// TestRedisConnectionFailureHandling tests behavior when Redis is unavailable
func (s *RedisIntegrationSuite) TestRedisConnectionFailureHandling() {
	// Create store with invalid address
	cfg := config.RedisConfig{
		Addr:         "invalid-host:6379",
		Password:     "",
		DB:           0,
		PoolSize:     1,
		MaxRetries:   1,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
		JobTTL:       1 * time.Hour,
	}

	_, err := redisstore.NewStore(cfg)
	assert.Error(s.T(), err, "Should fail to create store with invalid address")
	assert.Contains(s.T(), err.Error(), "failed to connect to Redis", "Error should indicate connection failure")
}

// TestRedisIntegrationSuite runs the test suite
func TestRedisIntegrationSuite(t *testing.T) {
	suite.Run(t, new(RedisIntegrationSuite))
}
