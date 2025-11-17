# Will-It-Compile: Implementation Plan

## Project Overview

A cloud-based service that validates whether user-submitted code compiles in specified target environments (e.g., Linux x86_64, C++20, Go 1.21, etc.). The service prioritizes security through sandboxed compilation and resource constraints.

## Technology Stack Evaluation

### Primary Language: Go
**Rationale:**
- Excellent Docker/container ecosystem integration (Docker SDK, containerd)
- Strong standard library for building web services
- Built-in concurrency primitives for handling multiple compilation requests
- Static binary compilation simplifies deployment
- Good performance characteristics for I/O-bound operations

**Recommendation:** Proceed with Go. Your intuition is correct - Go's container ecosystem maturity is a significant advantage.

## Architecture Overview

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │ HTTPS
       ▼
┌─────────────────────────────────────┐
│         API Gateway/Load Balancer    │
└──────────────┬──────────────────────┘
               │
       ┌───────┴───────┐
       │               │
       ▼               ▼
┌─────────────┐ ┌─────────────┐
│ API Server  │ │ API Server  │  (Horizontally scalable)
│  (Go)       │ │  (Go)       │
└──────┬──────┘ └──────┬──────┘
       │               │
       └───────┬───────┘
               │
       ┌───────┴────────┐
       │                │
       ▼                ▼
┌─────────────┐  ┌──────────────┐
│ Message     │  │   Database   │
│ Queue       │  │  (Results/   │
│ (Optional)  │  │   Metadata)  │
└──────┬──────┘  └──────────────┘
       │
       ▼
┌─────────────────┐
│ Compilation     │
│ Workers (Go)    │
│ ┌─────────────┐ │
│ │Docker       │ │
│ │Containers   │ │
│ └─────────────┘ │
└─────────────────┘
```

## Core Components

### 1. API Server
**Responsibilities:**
- Accept compilation requests (HTTP/gRPC)
- Validate input (file size, format, environment spec)
- Authentication & rate limiting
- Queue compilation jobs
- Return results to clients

**Endpoints:**
```
POST /api/v1/compile
  Body: {
    "code": "base64_encoded_source",
    "language": "cpp",
    "standard": "c++20",
    "architecture": "x86_64",
    "os": "linux",
    "compiler": "gcc-13"
  }
  Response: {
    "job_id": "uuid",
    "status": "queued|processing|completed|failed"
  }

GET /api/v1/compile/{job_id}
  Response: {
    "status": "completed",
    "compiled": true,
    "stdout": "...",
    "stderr": "...",
    "exit_code": 0
  }

GET /api/v1/environments
  Response: [
    {"language": "cpp", "compilers": ["gcc-11", "gcc-13", "clang-15"], ...},
    {"language": "go", "versions": ["1.20", "1.21"], ...}
  ]
```

### 2. Compilation Worker
**Responsibilities:**
- Pull jobs from queue or handle direct requests
- Create isolated Docker containers
- Execute compilation within time/resource limits
- Collect and sanitize output
- Clean up containers

**Implementation:**
```go
type CompilationJob struct {
    ID           string
    SourceCode   []byte
    Language     string
    Environment  EnvironmentSpec
    CreatedAt    time.Time
}

type CompilationResult struct {
    Success    bool
    Stdout     string
    Stderr     string
    ExitCode   int
    Duration   time.Duration
    Error      error
}
```

### 3. Environment Manager
**Responsibilities:**
- Maintain pre-built Docker images for each compiler/environment
- Handle image updates and versioning
- Provide image selection logic

**Supported Environments (Initial):**
- C/C++: GCC (11, 12, 13), Clang (14, 15, 16)
- Go: 1.20, 1.21, 1.22
- Rust: 1.70, 1.75, stable
- Python: 3.10, 3.11, 3.12 (for bytecode compilation)

## Security Considerations

### 1. Sandboxing Strategy: Docker + Security Layers

**Docker Configuration:**
```go
container, err := cli.ContainerCreate(ctx, &container.Config{
    Image: imageTag,
    Cmd: []string{"/usr/bin/compile.sh"},
    WorkingDir: "/workspace",
    User: "nobody:nobody",  // Run as non-root
    NetworkDisabled: true,  // No network access
}, &container.HostConfig{
    Resources: container.Resources{
        Memory:     128 * 1024 * 1024,  // 128MB limit
        MemorySwap: 128 * 1024 * 1024,  // No swap
        CPUQuota:   50000,              // 0.5 CPU
        PidsLimit:  100,                // Process limit
    },
    SecurityOpt: []string{
        "no-new-privileges",
        "seccomp=seccomp-profile.json",
    },
    ReadonlyRootfs: true,  // Read-only filesystem
    Tmpfs: map[string]string{
        "/tmp":       "rw,noexec,nosuid,size=64m",
        "/workspace": "rw,noexec,nosuid,size=32m",
    },
    CapDrop: []string{"ALL"},  // Drop all capabilities
}, nil, nil, "")
```

**Additional Security Layers:**
- **AppArmor/SELinux:** Enable mandatory access control profiles
- **Seccomp:** Restrict system calls (whitelist: read, write, open, close, exit, etc.)
- **No-new-privileges:** Prevent privilege escalation
- **Resource limits:** CPU, memory, disk I/O, PIDs, compilation time

### 2. Input Validation

**File Size Limits:**
- Maximum source file size: 1MB (configurable)
- Maximum project size (if supporting multi-file): 10MB

**Content Validation:**
- Check for valid UTF-8/ASCII encoding
- Scan for suspicious patterns (optional, may cause false positives)
- Reject binary uploads

**Rate Limiting:**
- Per-IP: 10 requests/minute
- Per-authenticated-user: 100 requests/minute
- Queue size limits to prevent DoS

### 3. Timeout & Resource Exhaustion Protection

```go
const (
    MaxCompilationTime = 30 * time.Second
    MaxOutputSize      = 1 * 1024 * 1024  // 1MB
)

ctx, cancel := context.WithTimeout(context.Background(), MaxCompilationTime)
defer cancel()
```

**Protection mechanisms:**
- Hard timeout on container execution
- Kill runaway compilation processes
- Limit output capture (stdout/stderr)
- Disk quota enforcement

### 4. Container Isolation

**Best Practices:**
- Use ephemeral containers (create -> run -> destroy)
- Never reuse containers between compilations
- Run containers in isolated namespaces
- Consider using gVisor or Kata Containers for additional isolation (advanced)

### 5. Output Sanitization

- Strip potential ANSI escape sequences that could affect terminal
- Limit output length
- Remove absolute paths that might leak system information
- Filter out sensitive environment variables if accidentally printed

### 6. Supply Chain Security

- Pin Docker base image versions with SHA256 digests
- Regularly scan images for vulnerabilities
- Use minimal base images (e.g., Alpine, distroless)
- Sign images and verify signatures

## Detailed Component Design

### Docker Image Structure

**Official Docker Hub Images:**
```
gcc:13                  # C/C++ (Debian-based)
golang:1.22-alpine      # Go (Alpine-based)
rust:1.75-alpine        # Rust (Alpine-based)
```

**Dockerfile Example (C++):**
```dockerfile
FROM alpine:3.19@sha256:...
RUN apk add --no-cache gcc g++ musl-dev
RUN adduser -D -u 1000 nobody
COPY compile.sh /usr/bin/compile.sh
RUN chmod +x /usr/bin/compile.sh
USER nobody
WORKDIR /workspace
```

**compile.sh:**
```bash
#!/bin/sh
set -e
timeout 25s g++ -std=c++20 -Wall -Wextra -o /tmp/output /workspace/source.cpp
```

### Compilation Flow

```go
func (w *Worker) CompileCode(job CompilationJob) CompilationResult {
    // 1. Select appropriate Docker image
    image := w.selectImage(job.Environment)

    // 2. Write source code to temporary location
    workDir, err := prepareWorkspace(job.SourceCode)
    defer cleanup(workDir)

    // 3. Create container with security constraints
    container := w.createSecureContainer(image, workDir)
    defer w.removeContainer(container.ID)

    // 4. Start container with timeout
    ctx, cancel := context.WithTimeout(context.Background(), MaxCompilationTime)
    defer cancel()

    err = w.docker.ContainerStart(ctx, container.ID, types.ContainerStartOptions{})

    // 5. Wait for completion or timeout
    statusCh, errCh := w.docker.ContainerWait(ctx, container.ID, container.WaitConditionNotRunning)

    select {
    case result := <-statusCh:
        // 6. Collect output
        stdout, stderr := w.collectOutput(container.ID)
        return CompilationResult{
            Success: result.StatusCode == 0,
            Stdout: stdout,
            Stderr: stderr,
            ExitCode: int(result.StatusCode),
        }
    case <-ctx.Done():
        w.docker.ContainerKill(context.Background(), container.ID, "SIGKILL")
        return CompilationResult{
            Success: false,
            Error: errors.New("compilation timeout"),
        }
    }
}
```

## Scalability Considerations

### Horizontal Scaling
- API servers are stateless and can scale independently
- Compilation workers can scale based on queue depth
- Use container orchestration (Kubernetes) for auto-scaling

### Queue-Based Architecture (Recommended for Production)
```
API Server → Redis/RabbitMQ → Worker Pool → Results Cache/DB
```

**Benefits:**
- Decouples request acceptance from processing
- Better handling of traffic spikes
- Retry mechanisms
- Worker failure isolation

### Resource Management
- Pool of pre-warmed Docker images to reduce cold-start latency
- Connection pooling for database/cache
- Graceful degradation under load

## Alternative Sandboxing Technologies

While Docker is recommended, consider these alternatives:

### 1. **gVisor** (Stronger Isolation)
- OCI-compatible runtime with additional syscall filtering
- Better isolation than standard containers
- Trade-off: ~10-30% performance overhead

### 2. **Firecracker** (AWS Technology)
- Lightweight microVMs
- Strong isolation boundaries
- Trade-off: More complex setup, longer cold-start

### 3. **Podman** (Rootless Containers)
- Docker alternative with rootless mode by default
- Similar API to Docker
- Better for multi-tenant environments

### 4. **Nsjail** (Direct Process Isolation)
- Lightweight namespace/seccomp wrapper
- Faster than containers
- Trade-off: More manual configuration, Linux-only

**Recommendation:** Start with Docker for MVP, evaluate gVisor for production if additional isolation is needed.

## Deployment Architecture

### Cloud Platform Options

**AWS:**
- ECS/Fargate for container orchestration
- Lambda for API (with cold-start considerations)
- SQS for job queue
- ElastiCache (Redis) for results

**GCP:**
- Cloud Run for API and workers
- Cloud Tasks for job queue
- Memorystore (Redis) for caching

**Self-hosted Kubernetes:**
- Most flexible, works on any cloud
- Full control over scaling and resources

**Recommendation:** Start with managed container service (ECS/Cloud Run) for simpler operations.

## Monitoring & Observability

**Key Metrics:**
- Compilation success/failure rate
- Average compilation time
- Queue depth
- Resource utilization (CPU, memory)
- Container creation/cleanup time
- Error rates by environment

**Tools:**
- Prometheus + Grafana for metrics
- OpenTelemetry for distributed tracing
- Structured logging (JSON) with log aggregation

## Development Phases

### Phase 1: MVP (4-6 weeks)
- [ ] Basic API server (single environment: C++ with GCC)
- [ ] Docker-based sandboxing
- [ ] Simple synchronous request/response
- [ ] Basic security controls (timeouts, resource limits)
- [ ] Local development environment

### Phase 2: Multi-Environment Support (2-3 weeks)
- [ ] Add support for multiple compilers and languages
- [ ] Environment configuration system
- [ ] Docker image management

### Phase 3: Production Hardening (3-4 weeks)
- [ ] Queue-based architecture
- [ ] Rate limiting & authentication
- [ ] Enhanced security (seccomp, AppArmor)
- [ ] Monitoring and alerting
- [ ] Load testing

### Phase 4: Advanced Features (Ongoing)
- [ ] Multi-file project support
- [ ] Custom compiler flags
- [ ] Dependency management (package managers)
- [ ] WebAssembly compilation targets
- [ ] Compilation caching

## Project Structure

```
will-it-compile/
├── cmd/
│   ├── api/           # API server entry point
│   └── worker/        # Worker service entry point
├── internal/
│   ├── api/           # HTTP handlers, middleware
│   ├── compiler/      # Compilation logic
│   ├── docker/        # Docker client wrapper
│   ├── environment/   # Environment/image management
│   ├── security/      # Security utilities
│   └── queue/         # Job queue (if async)
├── pkg/
│   └── models/        # Shared data models
├── images/            # Dockerfiles for compilation environments
│   ├── cpp/
│   ├── go/
│   └── rust/
├── configs/           # Configuration files
│   ├── seccomp-profile.json
│   └── environments.yaml
├── scripts/           # Build and deployment scripts
├── tests/
│   ├── integration/
│   └── e2e/
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## Open Questions & Decisions Needed

1. **Synchronous vs Asynchronous API:**
   - Sync: Simpler, but ties up connections
   - Async: More complex, better for scale
   - **Recommendation:** Start sync, move to async in Phase 3

2. **Authentication:**
   - API keys
   - OAuth2
   - Anonymous with strict rate limits
   - **Decision needed:** What's your user model?

3. **Multi-file Projects:**
   - Single file only (simpler)
   - Archive upload (zip/tar)
   - Git repository URL
   - **Recommendation:** Start single-file, add archives later

4. **Result Storage:**
   - In-memory (Redis) with TTL
   - Database (PostgreSQL)
   - **Recommendation:** Redis for recent results + DB for audit logs

5. **Custom Dependencies:**
   - No external dependencies (safest)
   - Allow standard library only
   - Support package managers (complex, security risk)
   - **Recommendation:** Standard library only for MVP

## Security Incident Response

**Container Escape Detection:**
- Monitor for unexpected network activity
- Alert on container lifecycle anomalies
- Log all container events

**Response Plan:**
1. Kill container immediately
2. Quarantine worker node
3. Analyze logs and source code
4. Review and update security policies

## Cost Estimation (AWS Example)

**Small Scale (1000 compilations/day):**
- ECS Fargate: ~$50/month
- ALB: ~$20/month
- ElastiCache: ~$15/month
- **Total: ~$85/month**

**Medium Scale (100k compilations/day):**
- ECS Fargate: ~$500/month
- ALB: ~$20/month
- ElastiCache: ~$50/month
- SQS: ~$5/month
- **Total: ~$575/month**

## Next Steps

1. Review and approve this plan
2. Set up project repository and structure
3. Implement Phase 1 MVP
4. Create Docker images for first environment (C++)
5. Build API server with basic compilation endpoint
6. Add comprehensive tests
7. Deploy to staging environment

## Conclusion

Go + Docker is an excellent choice for this project. The combination provides strong security primitives, good performance, and straightforward cloud deployment. The key to success will be rigorous security testing and proper resource isolation. Consider engaging a security consultant for a review before production launch.
