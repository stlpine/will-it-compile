# Claude Code Context for will-it-compile

This document provides context for Claude Code when working on the will-it-compile project. It contains architectural decisions, common tasks, and important implementation details.

## Project Overview

**will-it-compile** is a secure, cloud-ready service that checks whether code compiles in specified environments. It uses Docker containers with strict security controls to safely compile untrusted code.

### Key Design Principles
1. **Security First**: Multiple layers of isolation and resource controls
2. **Simplicity**: MVP focuses on core functionality (C++ compilation only)
3. **Cloud Ready**: Designed for container orchestration platforms (see note below)
4. **Go Best Practices**: Standard project layout with cmd/internal/pkg structure

### ⚠️ Important: Kubernetes Deployment

**Current MVP Architecture**: Uses Docker client to create containers directly. This works for:
- ✅ Local development
- ✅ Single-server deployments
- ✅ VM-based deployments with Docker installed

**Does NOT work in Kubernetes** because:
- ❌ No Docker daemon in K8s pods
- ❌ Only container runtime (containerd, CRI-O) at node level
- ❌ Mounting Docker socket is a security anti-pattern

**For Kubernetes deployment**, see: [`docs/architecture/KUBERNETES_ARCHITECTURE.md`](./docs/architecture/KUBERNETES_ARCHITECTURE.md)
- Uses Kubernetes Jobs API to create ephemeral pods
- Proper RBAC and security boundaries
- Native K8s integration

## Project Structure

```
will-it-compile/
├── cmd/                  # Executable entry points
│   ├── api/              # API server (main.go)
│   ├── cli/              # Command-line tool
│   ├── tui/              # Terminal UI
│   └── worker/           # Background worker (future)
├── internal/             # Private application code
│   ├── api/              # HTTP handlers, middleware, job storage
│   ├── compiler/         # Compilation orchestration
│   ├── docker/           # Docker client wrapper with security
│   ├── environment/      # Environment management
│   ├── runtime/          # Runtime execution
│   └── security/         # Security utilities
├── pkg/                  # Public/shared code
│   ├── models/           # Shared data types
│   └── runtime/          # Runtime models
├── web/                  # React frontend (NEW - planned)
│   ├── src/              # React source code
│   │   ├── components/   # UI components
│   │   ├── pages/        # Page components
│   │   ├── services/     # API client
│   │   ├── hooks/        # Custom React hooks
│   │   └── types/        # TypeScript types
│   ├── public/           # Static assets
│   └── README.md         # Frontend documentation
├── docs/                 # Documentation (NEW - organized)
│   ├── architecture/     # System design & deployment
│   ├── development/      # Development guides
│   ├── guides/           # User guides (CLI, TUI, API)
│   └── technical/        # Technical implementation details
├── images/               # Docker images
│   └── cpp/              # C++ compiler image
├── configs/              # Configuration files (seccomp, environments)
├── scripts/              # Build and test scripts
├── tests/                # Integration tests and sample code
│   ├── integration/      # Integration tests
│   └── samples/          # Sample code files
└── deployments/          # Deployment configurations
    ├── helm/             # Kubernetes Helm charts
    └── DEPLOYMENT_GUIDE.md
```

### Monorepo Organization

This project uses a **monorepo structure**:
- **Backend** (Go): API server, CLI, TUI in `cmd/` and `internal/`
- **Frontend** (React): Web interface in `web/`
- **Docs**: All documentation organized in `docs/`
- **Infrastructure**: Docker images, configs, deployment manifests

Benefits:
- Shared version control and releases
- Easier development (single clone, unified PRs)
- Type consistency (web/src/types mirrors pkg/models)
- Coordinated deployments

## Important Architectural Decisions

### 1. Docker Client Choice
**Decision**: Use `github.com/docker/docker/client` (moby) instead of `github.com/docker/go-sdk`

**Rationale**:
- moby/docker is production-ready and stable
- Fine-grained control over security settings (critical for untrusted code)
- Industry standard with extensive documentation
- go-sdk is still pre-v1.0 and API may change

**Location**: `internal/docker/client.go`

### 2. Synchronous Processing (MVP)
**Decision**: Process compilation requests synchronously in goroutines

**Rationale**:
- Simpler for MVP
- Easier to test and debug
- Queue-based architecture planned for Phase 2

**Location**: `internal/api/handlers.go:131` (processJob function)

### 3. In-Memory Job Storage
**Decision**: Store jobs and results in memory using sync.RWMutex

**Rationale**:
- Sufficient for MVP and testing
- No external dependencies
- Production should use Redis or database

**Location**: `internal/api/handlers.go:161` (jobStore)

**⚠️ Important**: This is not suitable for production with multiple instances. Replace with Redis or database in Phase 3.

### 4. Security Layers
**Implementation**: Multiple overlapping security controls

**Layers**:
1. Docker container isolation
2. No network access
3. Resource limits (CPU: 0.5, Memory: 128MB, PIDs: 100)
4. All capabilities dropped
5. Seccomp profile (syscall whitelist)
6. Non-root user execution
7. 30-second timeout
8. Output size limits (1MB)
9. Tmpfs for /tmp (prevents permanent writes)

**Location**: `internal/docker/client.go:93` (createSecureContainer)

## Key Components

### HTTP Framework
**Decision**: Use [Echo framework](https://echo.labstack.com/) (v4) for HTTP routing and middleware

**Rationale**:
- High performance with minimal overhead
- Built-in middleware (CORS, recovery, logging)
- Clean, intuitive API
- Excellent request/response handling
- Strong community support and extensive documentation

### API Server (`cmd/api/main.go`)
- Entry point for the API service
- Creates Echo instance with configured routes
- Handles graceful shutdown with context timeout
- Default port: 8080 (configurable via PORT env var)

### Server Setup (`internal/api/server.go`)
- `NewEchoServer()`: Factory function for creating configured Echo instances
- Centralizes route registration and middleware setup
- Supports optional rate limiting (enabled in production, disabled in tests)

### Handlers (`internal/api/handlers.go`)
- All handlers use Echo's `echo.Context` for request/response
- `HandleCompile`: POST /api/v1/compile - Submit compilation job
- `HandleGetJob`: GET /api/v1/compile/:job_id - Get job result (uses path parameter)
- `HandleGetEnvironments`: GET /api/v1/environments - List supported environments
- `HandleHealth`: GET /health - Health check

### Middleware (`internal/api/middleware.go`)
- **Logging**: Echo's built-in `middleware.Logger()` (JSON format with timestamps, latency, etc.)
- **Recovery**: Echo's built-in `middleware.Recover()` (panic recovery)
- **CORS**: Echo's built-in `middleware.CORSWithConfig()` (cross-origin requests)
- **Rate Limiting**: Custom `RateLimitMiddleware` (10 req/min per IP, token bucket algorithm)

### Compiler (`internal/compiler/compiler.go`)
- Orchestrates compilation process
- Validates requests (max 1MB source code)
- Selects appropriate Docker image
- Returns structured results

### Docker Client (`internal/docker/client.go`)
- Wraps Docker SDK with security defaults
- Creates isolated containers
- Copies source code via tar archive
- Collects and sanitizes output
- Enforces timeouts and resource limits

### Web Frontend (`web/` - Planned)

**Status**: 🚧 Planned - Structure created, implementation pending

**Technology Stack**:
- **React 18** with TypeScript for type safety
- **Monaco Editor** for code editing (VS Code's editor)
- **Tailwind CSS** for styling
- **Vite** for fast development and builds
- **React Query** or Zustand for state management

**Directory Structure**:
```
web/
├── src/
│   ├── components/         # Reusable UI components
│   │   ├── CodeEditor/     # Monaco-based code editor
│   │   ├── CompilerOutput/ # Result display
│   │   └── EnvironmentSelector/ # Language picker
│   ├── pages/              # Route components
│   │   ├── Home.tsx        # Main compilation page
│   │   └── JobHistory.tsx  # Past jobs
│   ├── services/
│   │   └── api.ts          # Backend API client
│   ├── hooks/
│   │   └── useCompilation.ts  # Compilation logic
│   ├── types/
│   │   └── api.ts          # TypeScript types (sync with Go)
│   └── App.tsx
└── README.md               # Frontend documentation
```

**API Integration**:
- Communicates with API server at `http://localhost:8080`
- Uses same API endpoints as CLI/TUI
- Types in `web/src/types/api.ts` must match `pkg/models/models.go`

**Development Workflow**:
1. Start backend: `make run` (port 8080)
2. Start frontend: `cd web && npm start` (port 3000)
3. Frontend proxies API requests to backend

**Important**: When modifying API models in `pkg/models/models.go`, update TypeScript types in `web/src/types/api.ts` to maintain type consistency.

## Common Tasks

### Building the Project
```bash
# Build API server
make build

# Build Docker images
make docker-build

# Build everything
make all
```

### Running the Server
```bash
# Using make
make run

# Direct execution
./bin/will-it-compile-api

# With custom port
PORT=3000 ./bin/will-it-compile-api
```

### Testing
```bash
# Run unit tests
make test

# Run integration tests (requires Docker)
go test -v ./tests/integration/

# Test with coverage
make test-coverage

# Test API with script
chmod +x scripts/test-api.sh
./scripts/test-api.sh
```

### Docker Operations
```bash
# Build C++ compiler image
cd images/cpp && ./build.sh

# Test Docker image manually
docker run --rm \
  -v /path/to/source.cpp:/workspace/source.cpp:ro \
  will-it-compile/cpp-gcc:13-alpine

# Clean up Docker images
make docker-clean
```

## Code Modification Guidelines

### Adding a New Language/Compiler

1. **Create Docker image** in `images/{language}/`:
   ```
   images/rust/
   ├── Dockerfile
   ├── compile.sh
   └── build.sh
   ```

2. **Update environment spec** in `internal/compiler/compiler.go`:
   ```go
   environments := map[string]models.EnvironmentSpec{
       "rust-stable": {
           Language: "rust",
           Compiler: "rustc",
           Version:  "1.75",
           ImageTag: "will-it-compile/rust:1.75-alpine",
       },
   }
   ```

3. **Update validation** in `validateRequest()` function

4. **Update tests** in `tests/integration/`

5. **Update documentation** in README.md and configs/environments.yaml

6. **Update web frontend** (if implemented):
   - Add language to `web/src/types/api.ts`
   - Update environment selector in `web/src/components/EnvironmentSelector/`

### Adding New API Endpoints

1. **Add handler** in `internal/api/handlers.go`
2. **Register route** in `cmd/api/main.go`
3. **Add tests** in `tests/integration/api_test.go`
4. **Update README** with endpoint documentation

### Modifying Security Settings

**⚠️ Important**: Always test security changes thoroughly

**Resource limits** - `internal/docker/client.go:17-23`:
```go
const (
    MaxMemory     = 128 * 1024 * 1024
    MaxCPUQuota   = 50000
    MaxPidsLimit  = 100
    // ...
)
```

**Seccomp profile** - `configs/seccomp-profile.json`:
- Default action: SCMP_ACT_ERRNO (deny by default)
- Whitelist required syscalls only

**Container config** - `internal/docker/client.go:93-125`:
- Modify `createSecureContainer()` function
- Test thoroughly with malicious code samples

## Security Testing Checklist

When modifying security features, test with:

- [ ] Fork bomb attempt
- [ ] Memory exhaustion
- [ ] CPU exhaustion
- [ ] Infinite loop
- [ ] File system writes
- [ ] Network access attempts
- [ ] Process spawning
- [ ] Large output generation

Sample malicious code in: `tests/samples/` (create these for testing)

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP server port |

Future additions:
- `REDIS_URL` - For job queue (Phase 3)
- `LOG_LEVEL` - Logging verbosity
- `MAX_CONCURRENT_JOBS` - Worker pool size

## Dependencies

### Direct Dependencies
- `github.com/docker/docker@v28.5.2` - Docker client
- `github.com/google/uuid@v1.6.0` - UUID generation
- `github.com/labstack/echo/v4@v4.13.4` - HTTP web framework
- `github.com/stretchr/testify@v1.11.1` - Testing toolkit (test dependency)

### Why These Versions?
- Docker v28.5.2: Latest stable with security patches
- UUID v1.6.0: Latest stable
- Echo v4.13.4: Latest v4 release, high-performance HTTP framework
- Testify v1.11.1: Latest stable, industry-standard testing framework

### Updating Dependencies
```bash
# Update all dependencies
go get -u ./...
go mod tidy

# Update specific dependency
go get github.com/docker/docker@latest
go mod tidy

# Verify after update
make test
make docker-build
```

## Known Limitations (MVP)

1. **Single file compilation only**: No multi-file project support
2. **In-memory storage**: Jobs lost on restart
3. **No authentication**: Anonymous access with rate limiting only
4. **No job queue**: Synchronous processing
5. **One language**: C++ only (GCC 13)
6. **No caching**: Each compilation runs fresh
7. **No metrics**: Basic logging only
8. **No web interface**: CLI and TUI only (web frontend planned)

See `docs/architecture/IMPLEMENTATION_PLAN.md` for Phase 2+ features.

## Troubleshooting

### Verify Setup
Before running the server, verify your setup:
```bash
./scripts/verify-setup.sh
```
This checks:
- Docker daemon is running
- Required Docker images are built
- Go is installed
- Project builds successfully

### "Failed to create docker client"
- Ensure Docker daemon is running
- Check Docker socket permissions: `/var/run/docker.sock`
- Try: `docker ps` to verify access

### "Missing required Docker images"
**New in startup**: The server now verifies all required Docker images exist at startup and will refuse to start if any are missing.

To fix:
- Build all images: `make docker-build`
- Or manually: `cd images/cpp && ./build.sh`
- Verify: `./scripts/verify-setup.sh`

**Why not build on-demand?**
Building Docker images on-demand would be a security risk:
- Supply chain attacks (compromised base images)
- Resource exhaustion (DoS attacks)
- Unpredictable performance
- Unverified dependencies

All images must be pre-built and verified during deployment.

### "Rate limit exceeded"
- Default: 10 requests/minute per IP
- Adjust in `cmd/api/main.go:27`
- Or wait 1 minute

### "Compilation timeout"
- Default: 30 seconds
- Adjust `MaxCompilationTime` in `internal/docker/client.go:23`
- Consider if code has infinite loops

### Tests failing
- Ensure Docker is running: `docker ps`
- Build images: `make docker-build`
- Check Go version: `go version` (requires 1.24+)

## Performance Considerations

### Container Cold Start
- First compilation ~2-3 seconds (container creation)
- Subsequent compilations ~1-2 seconds
- Consider pre-warming containers in production

### Resource Usage
- Each compilation: ~128MB RAM, 0.5 CPU
- Plan for N concurrent compilations × 128MB
- Recommend: 4GB RAM for ~20 concurrent jobs

### Scaling Strategy
1. **Vertical**: Increase host resources
2. **Horizontal**: Deploy multiple API instances (requires Redis for job storage)
3. **Kubernetes**: Use HPA based on queue depth

## Future Enhancements (Hints for Next Sessions)

### Phase 2 (Multi-Language Support)
- Add Go, Rust, Python compilers
- Create `internal/environment/manager.go` for image selection
- Update `configs/environments.yaml` with all supported environments

### Phase 3 (Production Hardening)
- Replace `jobStore` with Redis client
- Add message queue (Redis/RabbitMQ)
- Implement worker pool pattern
- Add Prometheus metrics
- Structured logging with levels

### Phase 4 (Advanced Features)
- Multi-file project support (zip/tar upload)
- Dependency management (package managers)
- Compilation caching (checksum-based)
- GitHub webhook integration

## Files to Review When...

### Adding Security Features
- `internal/docker/client.go` - Container security
- `configs/seccomp-profile.json` - Syscall whitelist
- `internal/security/` - Security utilities
- `docs/architecture/KUBERNETES_ARCHITECTURE.md` - K8s security model

### Debugging Compilation Issues
- `internal/compiler/compiler.go` - Compilation logic
- `images/cpp/compile.sh` - Compilation script
- `internal/docker/client.go` - Container execution
- `internal/runtime/` - Runtime execution

### Adding API Features
- `internal/api/handlers.go` - HTTP handlers
- `internal/api/middleware.go` - Middleware
- `internal/api/server.go` - Server setup and routing
- `cmd/api/main.go` - Entry point
- `pkg/models/models.go` - Data structures
- `web/src/types/api.ts` - Frontend types (must match backend)

### Working on Frontend
- `web/README.md` - Frontend documentation
- `web/src/services/api.ts` - API client
- `web/src/types/api.ts` - TypeScript types (sync with Go models)
- `pkg/models/models.go` - Backend models (keep in sync)
- `docs/guides/API_GUIDE.md` - API documentation (create if needed)

### Documentation
- `docs/README.md` - Documentation index
- `docs/architecture/` - System design docs
- `docs/guides/` - User-facing guides
- `docs/development/` - Development guides
- `README.md` - Main project overview

### Deployment
- `Makefile` - Build commands
- `scripts/setup.sh` - Setup automation
- `deployments/` - Deployment configs and Helm charts
- `docs/architecture/IMPLEMENTATION_PLAN.md` - Architecture details
- `docs/architecture/KUBERNETES_ARCHITECTURE.md` - K8s deployment
- `docs/architecture/DEPLOYMENT_ENVIRONMENTS.md` - Multi-environment strategy

## Code Style Guidelines

### Error Handling
```go
// Wrap errors with context
if err != nil {
    return fmt.Errorf("failed to create container: %w", err)
}
```

### Logging
```go
// Use standard log package (consider structured logging in Phase 3)
log.Printf("Processing job %s for language %s", jobID, language)
```

### Context Usage
```go
// Always pass context, respect timeouts
ctx, cancel := context.WithTimeout(context.Background(), MaxCompilationTime)
defer cancel()
```

### Security Constants
```go
// Define security limits as constants, not magic numbers
const MaxSourceSize = 1 * 1024 * 1024
```

## Testing Philosophy

### Unit Tests
- Test business logic in isolation
- Mock Docker client for compiler tests
- Fast, no external dependencies

### Integration Tests
- Test full flow with real Docker
- Located in `tests/integration/`
- Run with: `go test -v ./tests/integration/`
- Skip in CI with: `go test -short` (checks `testing.Short()`)

### Manual Testing
- Use `scripts/test-api.sh` for end-to-end testing
- Test with sample code in `tests/samples/`
- Always test security boundaries

### Testing Framework: Testify

**Decision**: Use `github.com/stretchr/testify` for all tests

**Benefits**:
- Readable assertions: `assert.Equal(t, expected, actual)`
- `require` package for critical assertions that should stop test execution
- Suite support for setup/teardown with `testify/suite`
- Mock support for creating test doubles
- Table-driven test patterns

**Key Testify Packages**:
1. **assert**: Assertions that continue test execution on failure
2. **require**: Assertions that stop test execution on failure (use for setup/prerequisites)
3. **suite**: Test suite with setup/teardown hooks
4. **mock**: Mock object generation (we use manual mocks for simplicity)

**Test Organization**:

```
tests/
├── integration/
│   ├── api_test.go           # Original tests (converted to testify)
│   ├── api_suite_test.go     # Suite-based tests with setup/teardown
│   └── table_driven_test.go  # Table-driven tests for multiple scenarios
internal/
├── compiler/
│   └── compiler_test.go      # Unit tests with mocks
└── docker/
    ├── interface.go           # DockerClient interface for mocking
    └── mock_test.go           # Mock implementation for tests
```

**Usage Examples**:

1. **Basic Assertions**:
```go
// assert continues on failure
assert.Equal(t, expected, actual, "optional message")
assert.NotEmpty(t, value)
assert.True(t, condition)

// require stops on failure (use for prerequisites)
require.NoError(t, err, "setup failed")
require.NotNil(t, object)
```

2. **Test Suite with Setup/Teardown**:
```go
type APISuite struct {
    suite.Suite
    server *api.Server
}

func (s *APISuite) SetupTest() {
    // Runs before each test
    server, err := api.NewServer()
    require.NoError(s.T(), err)
    s.server = server
}

func (s *APISuite) TearDownTest() {
    // Runs after each test
    s.server.Close()
}

func (s *APISuite) TestHealth() {
    // Test implementation using s.server
}

func TestAPISuite(t *testing.T) {
    suite.Run(t, new(APISuite))
}
```

3. **Table-Driven Tests**:
```go
testCases := []struct {
    name           string
    input          string
    expectedOutput string
    expectError    bool
}{
    {
        name:           "valid_code",
        input:          "int main() { return 0; }",
        expectedOutput: "success",
        expectError:    false,
    },
    // ... more cases
}

for _, tc := range testCases {
    tc := tc // capture range variable
    t.Run(tc.name, func(t *testing.T) {
        t.Parallel() // optional: run tests in parallel
        // test implementation
        assert.Equal(t, tc.expectedOutput, result)
    })
}
```

4. **Mock Interface Pattern**:
```go
// Define interface in production code
type DockerClient interface {
    RunCompilation(ctx context.Context, config CompilationConfig) (*CompilationOutput, error)
}

// Mock implementation for tests
type MockDockerClient struct {
    RunCompilationFunc func(ctx context.Context, config CompilationConfig) (*CompilationOutput, error)
}

func (m *MockDockerClient) RunCompilation(ctx context.Context, config CompilationConfig) (*CompilationOutput, error) {
    if m.RunCompilationFunc != nil {
        return m.RunCompilationFunc(ctx, config)
    }
    // default behavior
    return &CompilationOutput{ExitCode: 0}, nil
}

// Use in test
mockDocker := &MockDockerClient{
    RunCompilationFunc: func(ctx context.Context, config CompilationConfig) (*CompilationOutput, error) {
        return &CompilationOutput{ExitCode: 1}, errors.New("test error")
    },
}
```

**Best Practices**:
- Use `require` for setup/prerequisites that must succeed
- Use `assert` for actual test assertions (allows multiple failures to be reported)
- Always provide descriptive test names and messages
- Use table-driven tests for testing multiple scenarios
- Use test suites for shared setup/teardown logic
- Mock external dependencies (Docker, HTTP clients) for unit tests
- Keep integration tests separate from unit tests

**Running Tests**:
```bash
# Run all tests
go test ./...

# Run only unit tests (fast, no Docker required)
go test ./internal/...

# Run only integration tests
go test ./tests/integration/...

# Run specific test
go test -v -run TestCompile_Success ./internal/compiler/

# Run with coverage
go test -cover ./...

# Skip integration tests (uses testing.Short())
go test -short ./...
```

**Test Files**:
- `tests/integration/api_test.go` - Basic integration tests with testify assertions
- `tests/integration/api_suite_test.go` - Suite-based integration tests
- `tests/integration/table_driven_test.go` - Comprehensive table-driven scenarios
- `internal/compiler/compiler_test.go` - Unit tests for compiler with Docker mocks
- `internal/docker/mock_test.go` - Mock Docker client implementation
- `internal/docker/interface.go` - DockerClient interface for dependency injection

## Git Workflow

### Branches
- `main` - Production-ready code
- `develop` - Integration branch
- `feature/*` - New features
- `fix/*` - Bug fixes

### Commit Messages
```
feat: add Rust compiler support
fix: handle container timeout properly
docs: update API documentation
test: add integration tests for Go compilation
security: update seccomp profile
```

## Version Management

**Current Version**: MVP (Phase 1)

**Go Version**: 1.24 (specified in go.mod, mise.toml, .tool-versions)

**Docker Images**:
- `will-it-compile/cpp-gcc:13-alpine` - C++ GCC 13

**Versioning Strategy** (for future):
- API: v1, v2, etc. (URL versioning)
- Images: SemVer tags (e.g., `cpp-gcc:13.2.0-alpine`)

## Contact & Resources

- **Implementation Plan**: See `docs/architecture/IMPLEMENTATION_PLAN.md` for detailed architecture
- **API Documentation**: See README.md for endpoint details (consider creating `docs/guides/API_GUIDE.md`)
- **Security**: Review `configs/seccomp-profile.json` and Docker security docs
- **Docker SDK**: https://pkg.go.dev/github.com/docker/docker/client
- **Documentation Index**: See `docs/README.md` for all documentation
- **Frontend Guide**: See `web/README.md` for React frontend details

## Notes for Future Claude Sessions

### When Starting a New Session
1. Read this file first for context
2. Review IMPLEMENTATION_PLAN.md for architecture
3. Check current phase (MVP = Phase 1)
4. Review recent git commits for changes

### Before Making Changes
1. Understand the security implications
2. Check if change fits MVP scope
3. Run tests before and after changes
4. Update documentation

### Common Gotchas
- In-memory job storage is not persistent
- Docker daemon must be running for tests
- Rate limiting affects testing (10 req/min)
- Container creation takes 1-2 seconds

### Quick Start for Development
```bash
# Install dependencies
go mod download

# Build Docker images
make docker-build

# Build and run
make run

# In another terminal, test
./scripts/test-api.sh
```

---

**Last Updated**: 2025-11-14
**Project Phase**: MVP (Phase 1) + Phase 2 features
**Go Version**: 1.24
**Project Structure**: Monorepo (backend + frontend + docs)
**Claude Code Version**: This project was implemented with Claude Code
