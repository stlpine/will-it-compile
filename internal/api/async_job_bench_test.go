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
)

// Benchmarks demonstrate the performance benefits of virtualized time.
// These benchmarks would take minutes to run without synctest!

// BenchmarkAsyncJobProcessing_NoVirtualization shows real-world performance.
// This benchmark uses actual time.Sleep() to show the real cost.
func BenchmarkAsyncJobProcessing_NoVirtualization(b *testing.B) {
	b.Skip("Skipped by default - takes too long. Remove skip to compare.")

	server := &Server{
		compiler: &mockCompiler{compileDelay: 10 * time.Millisecond, shouldFail: false},
		jobs:     newJobStore(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		job := models.CompilationJob{
			ID: fmt.Sprintf("bench-job-%d", i),
			Request: models.CompilationRequest{
				Code:     "I2luY2x1ZGUgPGlvc3RyZWFtPg==",
				Language: models.LanguageCpp,
				Compiler: models.CompilerGCC9,
			},
			Status:    models.StatusQueued,
			CreatedAt: time.Now(),
		}

		done := make(chan struct{})
		go func() {
			server.processJob(job)
			close(done)
		}()
		<-done
	}
}

// BenchmarkAsyncJobProcessing_WithVirtualization shows synctest performance.
// With virtualized time, the 10ms delay becomes instant!
func BenchmarkAsyncJobProcessing_WithVirtualization(b *testing.B) {
	synctest.Test(&testing.T{}, func(t *testing.T) {
		server := &Server{
			compiler: &mockCompiler{compileDelay: 10 * time.Millisecond, shouldFail: false},
			jobs:     newJobStore(),
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			job := models.CompilationJob{
				ID: fmt.Sprintf("bench-job-%d", i),
				Request: models.CompilationRequest{
					Code:     "I2luY2x1ZGUgPGlvc3RyZWFtPg==",
					Language: models.LanguageCpp,
					Compiler: models.CompilerGCC9,
				},
				Status:    models.StatusQueued,
				CreatedAt: time.Now(),
			}

			done := make(chan struct{})
			go func() {
				server.processJob(job)
				close(done)
			}()
			<-done
		}
	})
}

// BenchmarkConcurrentJobs measures throughput with multiple parallel jobs.
func BenchmarkConcurrentJobs(b *testing.B) {
	synctest.Test(&testing.T{}, func(t *testing.T) {
		server := &Server{
			compiler: &mockCompiler{compileDelay: 5 * time.Millisecond, shouldFail: false},
			jobs:     newJobStore(),
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			const concurrency = 10
			done := make(chan struct{}, concurrency)

			for j := 0; j < concurrency; j++ {
				job := models.CompilationJob{
					ID: fmt.Sprintf("bench-%d-%d", i, j),
					Request: models.CompilationRequest{
						Code:     "I2luY2x1ZGUgPGlvc3RyZWFtPg==",
						Language: models.LanguageCpp,
						Compiler: models.CompilerGCC9,
					},
					Status:    models.StatusQueued,
					CreatedAt: time.Now(),
				}

				go func(j models.CompilationJob) {
					server.processJob(j)
					done <- struct{}{}
				}(job)
			}

			for j := 0; j < concurrency; j++ {
				<-done
			}
		}
	})
}

// BenchmarkJobStore measures storage performance under concurrent load.
func BenchmarkJobStore(b *testing.B) {
	store := newJobStore()

	b.Run("sequential_writes", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			job := models.CompilationJob{
				ID:        fmt.Sprintf("job-%d", i),
				Status:    models.StatusQueued,
				CreatedAt: time.Now(),
			}
			store.Store(job)
		}
	})

	b.Run("sequential_reads", func(b *testing.B) {
		// Prepopulate
		for i := 0; i < 1000; i++ {
			job := models.CompilationJob{
				ID:        fmt.Sprintf("job-%d", i),
				Status:    models.StatusQueued,
				CreatedAt: time.Now(),
			}
			store.Store(job)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			store.Get(fmt.Sprintf("job-%d", i%1000))
		}
	})

	b.Run("concurrent_mixed", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				if i%2 == 0 {
					job := models.CompilationJob{
						ID:        fmt.Sprintf("parallel-job-%d", i),
						Status:    models.StatusQueued,
						CreatedAt: time.Now(),
					}
					store.Store(job)
				} else {
					store.Get(fmt.Sprintf("parallel-job-%d", i-1))
				}
				i++
			}
		})
	})
}

// mockCompilerWithVariableDelay simulates realistic variable compilation times.
type mockCompilerWithVariableDelay struct {
	baseDelay time.Duration
}

func (m *mockCompilerWithVariableDelay) Compile(ctx context.Context, job models.CompilationJob) models.CompilationResult {
	// Simulate variable compilation time (50-150% of base)
	variance := time.Duration(len(job.Request.Code)%50) * time.Millisecond
	time.Sleep(m.baseDelay + variance)

	return models.CompilationResult{
		Success:  true,
		Compiled: true,
		ExitCode: 0,
		Stdout:   "compilation successful",
	}
}

func (m *mockCompilerWithVariableDelay) GetSupportedEnvironments() []models.Environment {
	return []models.Environment{
		{
			Language:  "cpp",
			Compilers: []string{"gcc-13"},
			Standards: []string{"c++20"},
			OSes:      []string{"linux"},
			Arches:    []string{"amd64"},
		},
	}
}

func (m *mockCompilerWithVariableDelay) Close() error {
	return nil
}

var _ compiler.CompilerInterface = (*mockCompilerWithVariableDelay)(nil)

// BenchmarkRealisticWorkload simulates a realistic mixed workload.
func BenchmarkRealisticWorkload(b *testing.B) {
	synctest.Test(&testing.T{}, func(t *testing.T) {
		server := &Server{
			compiler: &mockCompilerWithVariableDelay{baseDelay: 100 * time.Millisecond},
			jobs:     newJobStore(),
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Simulate bursts of concurrent jobs (realistic user behavior)
			const burstSize = 5
			done := make(chan struct{}, burstSize)

			for j := 0; j < burstSize; j++ {
				job := models.CompilationJob{
					ID: fmt.Sprintf("realistic-%d-%d", i, j),
					Request: models.CompilationRequest{
						Code:     fmt.Sprintf("code-with-length-%d", j*100),
						Language: models.LanguageCpp,
						Compiler: models.CompilerGCC9,
					},
					Status:    models.StatusQueued,
					CreatedAt: time.Now(),
				}

				go func(j models.CompilationJob) {
					server.processJob(j)
					done <- struct{}{}
				}(job)
			}

			for j := 0; j < burstSize; j++ {
				<-done
			}
		}
	})
}
