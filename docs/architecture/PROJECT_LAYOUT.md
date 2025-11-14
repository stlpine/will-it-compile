# Project Layout

This project follows the [Standard Go Project Layout](https://github.com/golang-standards/project-layout), which is widely adopted in the Go community.

## Directory Structure

```
will-it-compile/
├── cmd/                    # Main applications for this project
│   └── api/               # API server entry point (main.go)
├── internal/              # Private application code
│   ├── api/              # HTTP handlers, middleware, server
│   ├── compiler/         # Compilation orchestration
│   ├── docker/           # Docker client wrapper
│   └── runtime/          # Runtime implementations
│       ├── docker/       # Docker runtime adapter
│       └── kubernetes/   # Kubernetes runtime implementation
├── pkg/                   # Public library code (can be imported by external projects)
│   ├── models/           # Data models (shared types)
│   └── runtime/          # Runtime interface definition
├── deployments/          # Infrastructure and deployment configurations
│   ├── helm/            # Helm charts
│   └── DEPLOYMENT_GUIDE.md
├── images/               # Container images for compilation environments
│   └── cpp/             # C++ compiler image
├── configs/              # Configuration files
│   └── seccomp-profile.json
├── scripts/              # Build, install, analysis, and other scripts
├── tests/                # Additional external test apps and test data
│   ├── integration/     # Integration tests
│   └── samples/         # Sample code for testing
├── go.mod               # Go module definition
├── go.sum               # Go module checksums
├── Makefile             # Build automation
└── README.md            # Project documentation
```

## Directory Explanations

### `/cmd`

Main applications for this project.

- The directory name for each application should match the name of the executable
- Don't put a lot of code in this directory
- Import and invoke code from `/internal` and `/pkg` directories
- Example: `cmd/api/main.go` creates the API server

**Why this structure?**
- Clear separation of concerns
- Easy to add multiple executables (e.g., `cmd/worker/`, `cmd/cli/`)
- Entry points are obvious

### `/internal`

Private application and library code.

- Code you don't want others importing in their applications or libraries
- Go compiler enforces this: external packages cannot import anything from `/internal`
- Can have your own internal package structure (e.g., `/internal/api`, `/internal/compiler`)

**Why use `/internal`?**
- Prevents external packages from depending on your implementation details
- Freedom to refactor without breaking external users
- Clear signal: "This is our private code"

**What goes here?**
- Application-specific business logic
- Internal services and utilities
- Adapters and implementations

### `/pkg`

Library code that's safe to use by external applications.

- Other projects can import these packages
- Think carefully before putting code here
- Only put well-designed, stable APIs

**Why use `/pkg`?**
- Clear signal: "This is our public API"
- Other projects can import these packages
- Forces you to think about API design

**What goes here?**
- Reusable interfaces
- Well-defined data models
- Stable utilities

**Current example:**
- `pkg/runtime/` - Runtime interface (can be implemented externally)
- `pkg/models/` - Data models (can be used by clients)

### `/deployments` (or `/deploy`)

Infrastructure and deployment configurations.

**Standard names:**
- `/deployments` - preferred by golang-standards
- `/deploy` - shorter alternative
- Both are widely used in Go projects

**What goes here?**
- Kubernetes manifests
- Helm charts
- Docker Compose files
- Terraform configurations
- CI/CD pipeline definitions
- Deployment scripts

**Why separate from source?**
- Clear separation of code vs. infrastructure
- Different teams may manage them
- Different change frequency
- Infrastructure as code should be versioned separately

### `/images`

Container images needed by the application.

**Not in standard layout, but useful for:**
- Building custom runtime environments
- Compilation/execution containers
- Sidecar containers

**Why here?**
- These aren't the main application images
- They're dependencies/tools used by the application
- Each has its own Dockerfile and build process

### `/configs`

Configuration file templates or default configs.

**What goes here?**
- Configuration file templates
- Default configurations
- Security policies (seccomp, AppArmor)
- Environment-specific configs

**Why separate?**
- Easy to find and modify
- Can be templated/generated
- Security configurations need careful review

### `/scripts`

Scripts to perform various build, install, analysis, etc. operations.

**What goes here?**
- Build scripts
- Test scripts
- Installation scripts
- Code generation scripts

**Why separate?**
- Keeps root clean
- Easy to find automation
- Can be language-agnostic (shell, python, etc.)

### `/tests`

Additional external test applications and test data.

**What goes here?**
- Integration tests
- End-to-end tests
- Test data and fixtures
- Test utilities

**Why separate from `/internal`?**
- Internal package tests stay with the code (`*_test.go`)
- Integration tests that need external dependencies go here
- Test data that's too large for inline fixtures

## What's NOT in This Project (But You Might See Elsewhere)

### `/api`

- OpenAPI/Swagger specs
- JSON schema files
- Protocol definition files (protobuf)

We don't have this yet, but could add for API documentation.

### `/web`

- Web application specific components
- Static web assets
- Server-side templates

Not applicable for our API-only service.

### `/docs`

- User documentation
- Architecture diagrams
- Design documents

Currently using root-level markdown files (README.md, etc.).

### `/vendor`

- Application dependencies (vendored)

We use Go modules without vendoring.

### `/build`

- Packaging and Continuous Integration

We use `Makefile` in root instead.

### `/examples`

- Examples for your applications or public libraries

Could add this for API usage examples.

## Comparison with Other Projects

### Kubernetes

```
kubernetes/
├── cmd/              # Multiple executables (kubectl, kubelet, etc.)
├── pkg/              # Public libraries
├── staging/          # Staging area for published packages
├── cluster/          # Cluster deployment configs
└── build/            # Build scripts
```

### Prometheus

```
prometheus/
├── cmd/              # promtool, prometheus executables
├── pkg/              # Public packages (labels, timestamp)
├── documentation/    # User docs and examples
└── scripts/          # Build/test scripts
```

### Our Project (will-it-compile)

```
will-it-compile/
├── cmd/              # api executable
├── internal/         # Private code
├── pkg/              # Public interfaces
├── deployments/      # Helm charts, K8s configs
├── images/           # Compiler container images
├── configs/          # Config files
├── scripts/          # Build scripts
└── tests/            # Integration tests
```

## Best Practices

### Do's ✅

1. **Keep `/cmd` thin** - Just create and configure, then call `/internal`
2. **Use `/internal` liberally** - Most application code goes here
3. **Be selective with `/pkg`** - Only stable, well-designed APIs
4. **Document public APIs** - Anything in `/pkg` should be documented
5. **Separate concerns** - Each package should have a clear purpose
6. **Follow Go conventions** - Package names match directory names

### Don'ts ❌

1. **Don't put business logic in `/cmd`** - Keep main.go simple
2. **Don't expose internal details in `/pkg`** - Think about API stability
3. **Don't mix deployment and source** - Keep infrastructure separate
4. **Don't create deep hierarchies** - Prefer flat, clear structure
5. **Don't put everything in `/pkg`** - Most code should be `/internal`

## Why This Matters

1. **Familiarity**: Other Go developers know where to find things
2. **Tooling**: Many tools expect this structure
3. **Collaboration**: Clear boundaries between public/private code
4. **Scalability**: Easy to add new executables, libraries, deployments
5. **Imports**: Go's `/internal` enforcement prevents misuse

## Migration from Other Layouts

If you have a different structure:

1. **Identify entry points** → Move to `/cmd`
2. **Identify private code** → Move to `/internal`
3. **Identify public APIs** → Move to `/pkg`
4. **Identify deployment configs** → Move to `/deployments`
5. **Update imports** → Fix all import paths

## Additional Resources

- [Standard Go Project Layout](https://github.com/golang-standards/project-layout)
- [Go Project Structure](https://go.dev/doc/code)
- [Effective Go](https://go.dev/doc/effective_go)
- [Go Blog: Organizing Go Code](https://go.dev/blog/organizing-go-code)

## Our Specific Choices

### Why `/deployments` over `/deploy`?

- More explicit and self-documenting
- Recommended by golang-standards/project-layout
- Used by many large Go projects

### Why separate `/images`?

- These are build dependencies, not the main application
- Clear separation: "application code" vs. "build tooling"
- Easy to manage multiple compiler environments

### Why `/configs` for seccomp?

- Security policies are configuration
- Version controlled with the code
- Easy to find and audit

## Project-Specific Notes

### Runtime Abstraction

Our runtime abstraction is split:
- **Interface** in `/pkg/runtime/` - Can be implemented externally
- **Implementations** in `/internal/runtime/` - Private adapters

This allows:
- External projects to implement custom runtimes
- Internal flexibility to change implementations
- Clear public contract (CompilationRuntime interface)

### Models Package

`/pkg/models/` contains:
- Request/Response types (used by API clients)
- Domain entities (CompilationJob, Environment)
- Enums (Language, Compiler, Standard)

These are public because API clients need them.

### Testing Strategy

- **Unit tests**: Next to the code (`*_test.go`)
- **Integration tests**: In `/tests/integration/`
- **Mocks**: In same package as interface

This follows Go best practices for test organization.

---

Following these conventions makes the project:
- **Easier to understand** for new contributors
- **More maintainable** over time
- **Compatible** with Go tooling expectations
- **Professional** and production-ready
