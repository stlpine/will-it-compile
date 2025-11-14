# Will-It-Compile

A secure, cloud-ready service that checks whether code compiles in specified environments (e.g., Linux x86_64 with C++20). Compilation is performed in sandboxed Docker containers with strict security controls.

## Features

- **Multi-Environment Support**: Test code against different compilers, standards, and architectures
- **Secure Sandboxing**: Isolated Docker containers with resource limits, seccomp profiles, and capability restrictions
- **RESTful API**: Simple HTTP API for submitting code and retrieving results
- **Interactive TUI**: Rich terminal interface with live updates, job history, and file loading
- **CLI Tool**: Command-line tool for local development and CI/CD integration
- **Web Interface**: Modern React-based frontend for browser access (planned)
- **Rate Limiting**: Built-in protection against abuse
- **Cloud Ready**: Designed for deployment on container orchestration platforms

## Quick Start

### Prerequisites

- Go 1.24 or later
- Docker
- Make (optional, but recommended)

### Setup

1. Clone the repository:
```bash
git clone https://github.com/stlpine/will-it-compile.git
cd will-it-compile
```

2. Run the setup script:
```bash
chmod +x scripts/setup.sh
./scripts/setup.sh
```

Or manually:
```bash
# Install dependencies
go mod download

# Build Docker images
cd images/cpp && chmod +x build.sh compile.sh && ./build.sh && cd ../..

# Build the API server
go build -o bin/will-it-compile-api cmd/api/main.go
```

### Running the Server

```bash
# Using the binary
./bin/will-it-compile-api

# Or using make
make run

# Or using go run
go run cmd/api/main.go
```

The server will start on `http://localhost:8080` by default.

### Environment Variables

- `PORT`: Server port (default: 8080)

## CLI Tool

will-it-compile also provides a command-line interface for local development and scripting.

### Building the CLI

```bash
# Build CLI only
make build-cli

# Install to $GOPATH/bin
make install
```

### Quick Start

```bash
# Compile a C++ file
will-it-compile compile mycode.cpp

# Compile with specific standard
will-it-compile compile mycode.cpp --std=c++20

# List supported environments
will-it-compile environments

# Show version
will-it-compile version --help
```

### CLI Features

- **Local Compilation**: Uses Docker for isolated compilation
- **Auto-Detection**: Automatically detects language from file extension
- **Flexible Options**: Support for different standards and compilers
- **Scripting-Friendly**: Exit codes and quiet mode for CI/CD integration
- **Shell Completion**: Auto-generated completion for bash, zsh, fish, PowerShell

For detailed CLI documentation, see [docs/guides/CLI_GUIDE.md](docs/guides/CLI_GUIDE.md).

## TUI (Terminal User Interface)

will-it-compile includes an interactive TUI client that provides a rich terminal-based interface for submitting code and viewing results.

### Building the TUI

```bash
# Build TUI only
make build-tui

# Build everything (API, CLI, TUI)
make build
```

### Running the TUI

The TUI connects to a running API server (start it separately with `make run` in another terminal):

```bash
# Using make (connects to localhost:8080 by default)
make run-tui

# Or run directly
./bin/will-it-compile-tui

# Connect to a different API server
API_URL=http://localhost:3000 ./bin/will-it-compile-tui
```

### TUI Features

- **Interactive Code Editor**: Write or paste code with a multi-line text editor
- **File Loading**: Load code from local files (.cpp, .c, .go, .rs)
- **Live Compilation**: Submit code and watch compilation progress in real-time
- **Job History**: Browse previous compilation jobs and view detailed results
- **Live Monitoring**: Automatic polling for job status updates
- **Syntax-Aware**: Displays stdout, stderr, exit codes, and execution time
- **Keyboard Navigation**: Efficient keyboard shortcuts for all operations

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `Enter` | Submit code for compilation (in editor) |
| `f` | Open file picker to load code |
| `l` | Cycle through available languages |
| `Tab` | Toggle between editor and job history |
| `↑/↓` | Navigate in history or file picker |
| `?` | Show help screen |
| `Esc` | Return to editor |
| `q` / `Ctrl+C` | Quit |

For detailed TUI documentation, see [docs/guides/TUI_GUIDE.md](docs/guides/TUI_GUIDE.md).

### TUI Workflow

1. **Write Code**: Type or paste code in the editor, or press `f` to load from a file
2. **Select Language**: Press `l` to cycle through supported languages (C++, C, Go, Rust)
3. **Compile**: Press `Enter` to submit code to the API server
4. **View Results**: Automatically switches to job detail view showing compilation results
5. **Browse History**: Press `Tab` to view all previous jobs

### Environment Variables

- `API_URL`: API server URL (default: `http://localhost:8080`)

## Web Interface

will-it-compile includes a modern React-based web interface for browser access.

### Status

🚧 **Planned** - The web interface is currently in the planning phase. See [web/README.md](web/README.md) for the detailed implementation plan.

### Planned Features

- **Code Editor**: Monaco Editor with syntax highlighting
- **Live Compilation**: Real-time compilation with status updates
- **Multi-Language Support**: C++, Go, Rust, Python
- **Job History**: Browse and manage previous compilations
- **Responsive Design**: Works on desktop, tablet, and mobile
- **Dark/Light Themes**: User-selectable themes

### Quick Start (Once Implemented)

```bash
# Navigate to web directory
cd web

# Install dependencies
npm install

# Start development server
npm start

# Access at http://localhost:3000
```

For detailed web frontend documentation, see [web/README.md](web/README.md).

## API Documentation

### Endpoints

#### Health Check
```
GET /health
```

**Response:**
```json
{
  "status": "healthy",
  "time": "2024-01-15T10:30:00Z"
}
```

#### Get Supported Environments
```
GET /api/v1/environments
```

**Response:**
```json
[
  {
    "language": "cpp",
    "compilers": ["gcc-13"],
    "standards": ["c++11", "c++14", "c++17", "c++20", "c++23"],
    "oses": ["linux"],
    "architectures": ["x86_64"]
  }
]
```

#### Submit Compilation Job
```
POST /api/v1/compile
Content-Type: application/json
```

**Request Body:**
```json
{
  "code": "I2luY2x1ZGUgPGlvc3RyZWFtPgppbnQgbWFpbigpIHsKICAgIHN0ZDo6Y291dCA8PCAiSGVsbG8sIFdvcmxkISIgPDwgc3RkOjplbmRsOwogICAgcmV0dXJuIDA7Cn0=",
  "language": "cpp",
  "compiler": "gcc-13",
  "standard": "c++20",
  "architecture": "x86_64",
  "os": "linux"
}
```

**Note:** The `code` field must be Base64-encoded source code.

**Response:**
```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "queued"
}
```

#### Get Compilation Result
```
GET /api/v1/compile/{job_id}
```

**Response (Completed):**
```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "success": true,
  "compiled": true,
  "stdout": "Compilation successful\n",
  "stderr": "",
  "exit_code": 0,
  "duration": 1250000000
}
```

**Response (Failed Compilation):**
```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "success": true,
  "compiled": false,
  "stdout": "",
  "stderr": "source.cpp:3:5: error: expected ';' before 'return'\n",
  "exit_code": 1,
  "duration": 890000000
}
```

## Usage Examples

### Using cURL

```bash
# Encode your source code
SOURCE_CODE=$(echo '#include <iostream>
int main() {
    std::cout << "Hello, World!" << std::endl;
    return 0;
}' | base64)

# Submit compilation job
RESPONSE=$(curl -X POST http://localhost:8080/api/v1/compile \
  -H "Content-Type: application/json" \
  -d "{
    \"code\": \"${SOURCE_CODE}\",
    \"language\": \"cpp\",
    \"compiler\": \"gcc-13\",
    \"standard\": \"c++20\"
  }")

# Extract job ID
JOB_ID=$(echo $RESPONSE | jq -r .job_id)

# Wait a moment for compilation
sleep 3

# Get result
curl http://localhost:8080/api/v1/compile/${JOB_ID} | jq .
```

### Using the Test Script

A comprehensive test script is provided:

```bash
chmod +x scripts/test-api.sh
./scripts/test-api.sh
```

### Sample Code

Sample code files are available in `tests/samples/`:
- `hello_world.cpp` - Basic Hello World program
- `syntax_error.cpp` - Code with intentional syntax error
- `cpp20_features.cpp` - C++20 features demonstration

## Development

### Project Structure

```
will-it-compile/
├── cmd/                  # Executable entry points
│   ├── api/              # API server
│   ├── cli/              # Command-line tool
│   ├── tui/              # Terminal UI
│   └── worker/           # Background worker
├── internal/             # Private Go packages
│   ├── api/              # HTTP handlers and middleware
│   ├── compiler/         # Compilation logic
│   ├── docker/           # Docker client wrapper
│   ├── environment/      # Environment management
│   ├── runtime/          # Runtime execution
│   └── security/         # Security utilities
├── pkg/                  # Public Go packages
│   ├── models/           # Shared data models
│   └── runtime/          # Runtime models
├── web/                  # React frontend (planned)
│   ├── src/              # React source code
│   ├── public/           # Static assets
│   └── README.md         # Frontend documentation
├── docs/                 # Documentation
│   ├── architecture/     # Design and architecture docs
│   ├── development/      # Development guides
│   ├── guides/           # User guides
│   └── technical/        # Technical details
├── images/               # Docker images
│   └── cpp/              # C++ compilation image
├── configs/              # Configuration files
├── scripts/              # Build and deployment scripts
├── tests/                # Test files
│   ├── integration/      # Integration tests
│   └── samples/          # Sample code files
└── deployments/          # Deployment configurations
    └── helm/             # Kubernetes Helm charts
```

For more details, see [docs/architecture/PROJECT_LAYOUT.md](docs/architecture/PROJECT_LAYOUT.md).

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run integration tests (requires Docker)
go test -v ./tests/integration/
```

### Building

```bash
# Build the API server
make build

# Build Docker images
make docker-build

# Build everything
make all
```

### Code Formatting

```bash
# Format code
make fmt

# Run linter (requires golangci-lint)
make lint

# Run linter with auto-fix
make lint-fix
```

## Security

This service implements multiple layers of security:

### Container Isolation
- **Read-only root filesystem**: Prevents malicious code from modifying the container
- **No network access**: Containers cannot make external network connections
- **Resource limits**: CPU, memory, and process count restrictions
- **Capability dropping**: All Linux capabilities are dropped
- **Non-root execution**: Code runs as unprivileged user

### Seccomp Profile
A custom seccomp profile restricts system calls to a minimal whitelist required for compilation.

### Input Validation
- Maximum source code size: 1MB
- Base64 encoding validation
- Language and compiler validation

### Rate Limiting
- 10 requests per minute per IP address (configurable)
- Protection against DoS attacks

### Output Sanitization
- ANSI escape sequence removal
- Output size limits (1MB)
- Timeout on compilation (30 seconds)

## Deployment

### Docker Deployment

The API server can be containerized for deployment:

```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o api cmd/api/main.go

FROM alpine:3.19
RUN apk add --no-cache docker-cli
COPY --from=builder /app/api /usr/local/bin/
CMD ["api"]
```

**Note:** The Docker socket must be mounted to allow the API server to create compilation containers.

### Kubernetes Deployment

For production deployment on Kubernetes:

1. Build and push Docker images
2. Create ConfigMaps for configuration
3. Deploy with appropriate RBAC permissions
4. Use DaemonSets or node affinity for Docker socket access
5. Implement horizontal pod autoscaling

### Cloud Platforms

- **AWS**: Deploy on ECS/Fargate with proper IAM roles
- **GCP**: Use Cloud Run with appropriate permissions
- **Azure**: Deploy on Azure Container Instances

## Configuration

Configuration files are located in `configs/`:

- `environments.yaml`: Supported compilation environments
- `seccomp-profile.json`: Seccomp security profile

## Monitoring

Key metrics to monitor:
- Compilation success/failure rate
- Average compilation time
- Queue depth
- Container creation/cleanup time
- Error rates by environment
- Resource utilization

Recommended tools:
- Prometheus for metrics collection
- Grafana for visualization
- OpenTelemetry for distributed tracing

## Limitations

Current MVP limitations:
- Single-file compilation only
- Synchronous processing (async with message queue planned)
- In-memory job storage (production should use Redis/database)
- Limited language support (C++ only in MVP)

## Roadmap

### Phase 2
- [ ] Support for Go, Rust, Python
- [ ] Multiple compilers per language
- [ ] Custom compiler flags

### Phase 3
- [ ] Queue-based architecture with Redis
- [ ] Authentication and API keys
- [ ] Enhanced monitoring and metrics
- [ ] Multi-file project support

### Phase 4
- [ ] Dependency management
- [ ] Compilation caching
- [ ] WebAssembly targets
- [ ] GitHub integration

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

MIT License - see LICENSE file for details

## Security Issues

If you discover a security vulnerability, please email security@example.com instead of using the issue tracker.

## Documentation

Comprehensive documentation is available in the [docs/](docs/) directory:

- **[Architecture](docs/architecture/)** - System design and deployment strategies
- **[Development](docs/development/)** - Development guides and implementation details
- **[User Guides](docs/guides/)** - CLI, TUI, and API usage guides
- **[Technical Details](docs/technical/)** - In-depth implementation documentation

Quick links:
- [Implementation Plan](docs/architecture/IMPLEMENTATION_PLAN.md) - Detailed architecture and roadmap
- [CLI Guide](docs/guides/CLI_GUIDE.md) - Command-line tool documentation
- [TUI Guide](docs/guides/TUI_GUIDE.md) - Terminal UI documentation
- [Kubernetes Architecture](docs/architecture/KUBERNETES_ARCHITECTURE.md) - K8s deployment guide

## Support

For questions and support:
- GitHub Issues: https://github.com/stlpine/will-it-compile/issues
- Documentation: See [docs/](docs/) for comprehensive guides

## Acknowledgments

- Docker for containerization platform
- Go community for excellent libraries
- Alpine Linux for minimal base images
