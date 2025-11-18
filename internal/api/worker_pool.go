package api

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/stlpine/will-it-compile/pkg/models"
)

// WorkerPool manages a pool of workers for processing compilation jobs.
type WorkerPool struct {
	// Configuration
	maxWorkers int

	// Job queue
	jobQueue chan models.CompilationJob

	// Worker tracking
	activeWorkers   atomic.Int32
	availableSlots  atomic.Int32
	totalProcessed  atomic.Int64
	totalSuccessful atomic.Int64
	totalFailed     atomic.Int64

	// Server reference for job processing
	server *Server

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Start time for uptime calculation
	startTime time.Time
}

// WorkerStats represents the current state of the worker pool.
type WorkerStats struct {
	MaxWorkers      int       `json:"max_workers"`
	ActiveWorkers   int       `json:"active_workers"`
	AvailableSlots  int       `json:"available_slots"`
	QueuedJobs      int       `json:"queued_jobs"`
	TotalProcessed  int64     `json:"total_processed"`
	TotalSuccessful int64     `json:"total_successful"`
	TotalFailed     int64     `json:"total_failed"`
	Uptime          string    `json:"uptime"`
	UptimeSeconds   int64     `json:"uptime_seconds"`
	StartTime       time.Time `json:"start_time"`
}

// NewWorkerPool creates a new worker pool with the specified number of workers.
func NewWorkerPool(maxWorkers int, queueSize int, server *Server) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())

	pool := &WorkerPool{
		maxWorkers: maxWorkers,
		jobQueue:   make(chan models.CompilationJob, queueSize),
		server:     server,
		ctx:        ctx,
		cancel:     cancel,
		startTime:  time.Now(),
	}

	// Initially all workers are available
	pool.availableSlots.Store(int32(maxWorkers))

	return pool
}

// Start starts all workers in the pool.
func (wp *WorkerPool) Start() {
	log.Printf("Starting worker pool with %d workers (queue size: %d)", wp.maxWorkers, cap(wp.jobQueue))

	for i := 0; i < wp.maxWorkers; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}
}

// Stop gracefully stops the worker pool.
func (wp *WorkerPool) Stop() {
	log.Println("Stopping worker pool...")
	wp.cancel()
	close(wp.jobQueue)
	wp.wg.Wait()
	log.Println("Worker pool stopped")
}

// Submit submits a job to the worker pool.
// Returns true if the job was queued, false if the queue is full.
func (wp *WorkerPool) Submit(job models.CompilationJob) bool {
	select {
	case wp.jobQueue <- job:
		return true
	default:
		// Queue is full
		return false
	}
}

// GetStats returns the current worker pool statistics.
func (wp *WorkerPool) GetStats() WorkerStats {
	uptime := time.Since(wp.startTime)
	return WorkerStats{
		MaxWorkers:      wp.maxWorkers,
		ActiveWorkers:   int(wp.activeWorkers.Load()),
		AvailableSlots:  int(wp.availableSlots.Load()),
		QueuedJobs:      len(wp.jobQueue),
		TotalProcessed:  wp.totalProcessed.Load(),
		TotalSuccessful: wp.totalSuccessful.Load(),
		TotalFailed:     wp.totalFailed.Load(),
		Uptime:          formatUptime(uptime),
		UptimeSeconds:   int64(uptime.Seconds()),
		StartTime:       wp.startTime,
	}
}

// worker is the main worker loop that processes jobs from the queue.
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()

	log.Printf("Worker %d started", id)

	for {
		select {
		case <-wp.ctx.Done():
			log.Printf("Worker %d stopping", id)
			return

		case job, ok := <-wp.jobQueue:
			if !ok {
				// Channel closed, exit
				log.Printf("Worker %d: job queue closed", id)
				return
			}

			// Mark worker as active
			wp.activeWorkers.Add(1)
			wp.availableSlots.Add(-1)

			log.Printf("Worker %d: processing job %s", id, job.ID)

			// Process the job
			wp.server.processJob(job)

			// Update stats
			wp.totalProcessed.Add(1)

			// Check final job status
			finalJob, exists := wp.server.jobs.Get(job.ID)
			if exists {
				if finalJob.Status == models.StatusCompleted {
					wp.totalSuccessful.Add(1)
				} else if finalJob.Status == models.StatusFailed || finalJob.Status == models.StatusTimeout {
					wp.totalFailed.Add(1)
				}
			}

			// Mark worker as available
			wp.activeWorkers.Add(-1)
			wp.availableSlots.Add(1)

			log.Printf("Worker %d: finished job %s", id, job.ID)
		}
	}
}

// formatUptime formats a duration into a human-readable uptime string.
func formatUptime(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)
	} else if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}
