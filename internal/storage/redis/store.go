package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stlpine/will-it-compile/internal/config"
	"github.com/stlpine/will-it-compile/pkg/models"
)

// Store provides Redis-backed storage for jobs.
// Uses Redis hashes for structured job and result storage.
type Store struct {
	client *redis.Client
	ctx    context.Context
	ttl    time.Duration
}

// NewStore creates a new Redis job store.
func NewStore(cfg config.RedisConfig) (*Store, error) {
	client, err := NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return &Store{
		client: client.GetClient(),
		ctx:    context.Background(),
		ttl:    cfg.JobTTL,
	}, nil
}

// NewStoreWithClient creates a new Redis job store with an existing client.
// Useful for testing with miniredis.
func NewStoreWithClient(client *redis.Client, ttl time.Duration) *Store {
	return &Store{
		client: client,
		ctx:    context.Background(),
		ttl:    ttl,
	}
}

// Store saves or updates a job.
func (s *Store) Store(job models.CompilationJob) error {
	key := s.jobKey(job.ID)

	// Serialize timestamps
	createdAt := job.CreatedAt.Format(time.RFC3339Nano)
	startedAt := ""
	completedAt := ""

	if job.StartedAt != nil {
		startedAt = job.StartedAt.Format(time.RFC3339Nano)
	}
	if job.CompletedAt != nil {
		completedAt = job.CompletedAt.Format(time.RFC3339Nano)
	}

	// Serialize request as JSON
	requestJSON, err := json.Marshal(job.Request)
	if err != nil {
		return fmt.Errorf("failed to serialize job request: %w", err)
	}

	// Store as hash
	err = s.client.HSet(s.ctx, key, map[string]interface{}{
		"id":           job.ID,
		"request":      string(requestJSON),
		"status":       string(job.Status),
		"created_at":   createdAt,
		"started_at":   startedAt,
		"completed_at": completedAt,
	}).Err()

	if err != nil {
		return fmt.Errorf("failed to store job %s: %w", job.ID, err)
	}

	// Set TTL
	s.client.Expire(s.ctx, key, s.ttl)

	// Add to status index
	statusKey := s.statusIndexKey(job.Status)
	s.client.SAdd(s.ctx, statusKey, job.ID)
	s.client.Expire(s.ctx, statusKey, s.ttl)

	return nil
}

// Get retrieves a job by ID.
func (s *Store) Get(jobID string) (models.CompilationJob, bool) {
	key := s.jobKey(jobID)

	result, err := s.client.HGetAll(s.ctx, key).Result()
	if err != nil || len(result) == 0 {
		return models.CompilationJob{}, false
	}

	// Parse job from hash
	job := models.CompilationJob{
		ID:     result["id"],
		Status: models.JobStatus(result["status"]),
	}

	// Parse timestamps
	if createdAt, err := time.Parse(time.RFC3339Nano, result["created_at"]); err == nil {
		job.CreatedAt = createdAt
	}

	if result["started_at"] != "" {
		if startedAt, err := time.Parse(time.RFC3339Nano, result["started_at"]); err == nil {
			job.StartedAt = &startedAt
		}
	}

	if result["completed_at"] != "" {
		if completedAt, err := time.Parse(time.RFC3339Nano, result["completed_at"]); err == nil {
			job.CompletedAt = &completedAt
		}
	}

	// Parse request
	if err := json.Unmarshal([]byte(result["request"]), &job.Request); err != nil {
		return models.CompilationJob{}, false
	}

	return job, true
}

// StoreResult saves a compilation result.
func (s *Store) StoreResult(jobID string, result models.CompilationResult) error {
	key := s.resultKey(jobID)

	// Store as hash
	err := s.client.HSet(s.ctx, key, map[string]interface{}{
		"success":   result.Success,
		"compiled":  result.Compiled,
		"stdout":    result.Stdout,
		"stderr":    result.Stderr,
		"exit_code": result.ExitCode,
		"duration":  result.Duration.Nanoseconds(),
	}).Err()

	if err != nil {
		return fmt.Errorf("failed to store result for job %s: %w", jobID, err)
	}

	// Set TTL
	s.client.Expire(s.ctx, key, s.ttl)

	return nil
}

// GetResult retrieves a compilation result by job ID.
func (s *Store) GetResult(jobID string) (models.CompilationResult, bool) {
	key := s.resultKey(jobID)

	result, err := s.client.HGetAll(s.ctx, key).Result()
	if err != nil || len(result) == 0 {
		return models.CompilationResult{}, false
	}

	// Parse result from hash
	compilationResult := models.CompilationResult{
		Stdout: result["stdout"],
		Stderr: result["stderr"],
	}

	// Parse boolean fields
	if success, err := strconv.ParseBool(result["success"]); err == nil {
		compilationResult.Success = success
	}

	if compiled, err := strconv.ParseBool(result["compiled"]); err == nil {
		compilationResult.Compiled = compiled
	}

	// Parse integer fields
	if exitCode, err := strconv.Atoi(result["exit_code"]); err == nil {
		compilationResult.ExitCode = exitCode
	}

	// Parse duration
	if durationNs, err := strconv.ParseInt(result["duration"], 10, 64); err == nil {
		compilationResult.Duration = time.Duration(durationNs)
	}

	return compilationResult, true
}

// Close releases Redis connection.
func (s *Store) Close() error {
	return s.client.Close()
}

// Helper functions for Redis key generation

func (s *Store) jobKey(jobID string) string {
	return fmt.Sprintf("job:%s", jobID)
}

func (s *Store) resultKey(jobID string) string {
	return fmt.Sprintf("result:%s", jobID)
}

func (s *Store) statusIndexKey(status models.JobStatus) string {
	return fmt.Sprintf("job:index:status:%s", status)
}
