# GitHub Actions Workflows

This directory contains GitHub Actions workflows for CI/CD automation.

## Workflows

### 1. Pull Request Checks (`pull_request.yml`)

**Trigger**: On pull requests to `main` or `develop` branches

**Purpose**: Ensures code quality and functionality before merging

**Jobs**:
- **Lint**: Code formatting and linting checks
  - Verifies code is formatted (`go fmt`)
  - Runs `golangci-lint` for code quality
- **Unit Tests**: Fast tests without Docker dependencies
  - Runs unit tests using `make test-unit`
- **Integration Tests**: Full integration tests with Docker
  - Builds Docker images
  - Runs integration tests using `make test-integration`
- **Build**: Compiles all binaries
  - Builds API server, CLI, and TUI
  - Uploads binaries as artifacts
- **Docker**: Validates Docker images
  - Builds Docker images
  - Tests Docker images with sample code
- **Status Check**: Final check to ensure all jobs passed

**Usage**: Automatically runs on PR creation and updates

---

### 2. CI Pipeline (`ci.yml`)

**Trigger**:
- On push to `main` branch
- Manual trigger via `workflow_dispatch`

**Purpose**: Continuous integration for the main branch

**Jobs**:
- **Lint**: Code quality checks (same as PR workflow)
- **Test**: Comprehensive testing with coverage
  - Runs all tests with coverage using `make test-coverage`
  - Uploads coverage reports (HTML and data)
- **Build**: Builds all components
  - Compiles all binaries
  - Builds and tests Docker images
  - Uploads binaries with commit SHA
- **Security**: Security scanning
  - Runs `gosec` security scanner
  - Uploads SARIF results to GitHub Security
- **Verify**: Dependency verification
  - Verifies Go modules
  - Checks for vulnerabilities using `govulncheck`
- **Success**: Final status check

**Artifacts**:
- Coverage reports (retained for 30 days)
- Binaries tagged with commit SHA (retained for 90 days)

---

### 3. Release (`release.yml`)

**Trigger**: On pushing version tags (e.g., `v1.0.0`, `v2.1.3`)

**Purpose**: Automated release builds and GitHub release creation

**Jobs**:
- **Test**: Runs full test suite before building release
- **Build**: Cross-platform binary compilation
  - Builds for: Linux (amd64, arm64), macOS (amd64, arm64), Windows (amd64)
  - Creates archives (.tar.gz for Unix, .zip for Windows)
  - Injects version info, commit SHA, and build date into binaries
- **Release**: Creates GitHub release
  - Downloads all built binaries
  - Generates changelog from git commits
  - Creates release with downloadable assets
  - Marks as pre-release for `-rc`, `-beta`, `-alpha` tags

**How to Create a Release**:
```bash
# Tag the commit
git tag -a v1.0.0 -m "Release v1.0.0"

# Push the tag
git push origin v1.0.0

# GitHub Actions will automatically:
# 1. Run tests
# 2. Build binaries for all platforms
# 3. Create a GitHub release with binaries
```

**Release Assets**:
- `will-it-compile-api-{os}-{arch}[.exe].tar.gz` - API server
- `will-it-compile-{os}-{arch}[.exe].tar.gz` - CLI tool
- `will-it-compile-tui-{os}-{arch}[.exe].tar.gz` - TUI client

---

## Requirements

### GitHub Actions Runners

All workflows use `ubuntu-latest` runners with:
- Go 1.24
- Docker support
- 7GB RAM, 2-core CPU (GitHub-hosted default)

### Secrets

Currently, no secrets are required. The workflows use:
- `GITHUB_TOKEN` - Automatically provided by GitHub Actions

### Permissions

The workflows request minimal permissions:
- `contents: read` - Read repository contents (default for most workflows)
- `contents: write` - Write releases (release.yml only)
- `pull-requests: read` - Read PR info (pull_request.yml only)

---

## Workflow Features

### Caching

All workflows use Go module caching to speed up builds:
```yaml
uses: actions/setup-go@v5
with:
  go-version: '1.24'
  cache: true  # Caches Go modules
```

### Parallelization

The PR and CI workflows run jobs in parallel where possible:
- Linting runs independently
- Unit tests run separately from integration tests
- Build jobs can run concurrently

### Artifacts

Workflows upload artifacts for debugging and deployment:
- **PR Workflow**: Binaries (7 days)
- **CI Workflow**: Binaries (90 days), Coverage reports (30 days)
- **Release Workflow**: Cross-platform binaries (as release assets)

---

## Local Testing

Before pushing, test locally to catch issues early:

```bash
# Run what the workflows will run
make deps          # Download dependencies
make fmt           # Format code
make lint          # Run linter
make test-unit     # Unit tests
make docker-build  # Build Docker images
make test          # All tests
make build         # Build binaries
```

---

## Troubleshooting

### Workflow Fails on `make lint`

**Problem**: `golangci-lint` not installed or outdated

**Solution**: The workflow uses the official `golangci-lint-action` which downloads the latest version automatically. No local installation needed for CI.

### Workflow Fails on Docker Build

**Problem**: Docker image build fails

**Solution**:
- Check `images/cpp/Dockerfile` for syntax errors
- Verify base image is available
- Test locally: `cd images/cpp && ./build.sh`

### Integration Tests Timeout

**Problem**: Tests take too long or hang

**Solution**:
- Check for infinite loops in test code
- Increase timeout in workflow (default is 30 seconds per compilation)
- Review `internal/docker/client.go:23` for `MaxCompilationTime`

### Release Build Fails for Specific Platform

**Problem**: Cross-compilation fails for certain OS/arch combinations

**Solution**:
- Check if the platform combination is supported by Go
- Review the `matrix` strategy in `release.yml`
- Test locally with: `GOOS=linux GOARCH=arm64 go build ./cmd/api/`

---

## Future Enhancements

Potential workflow improvements:

- [ ] Add code coverage reporting to PRs (e.g., Codecov)
- [ ] Deploy to staging environment on main branch
- [ ] Push Docker images to registry on release
- [ ] Add performance benchmarking
- [ ] Scheduled security scans (weekly)
- [ ] Automated dependency updates (Dependabot)
- [ ] Slack/Discord notifications for build failures

---

## References

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [golangci-lint GitHub Action](https://github.com/golangci/golangci-lint-action)
- [Go setup action](https://github.com/actions/setup-go)
- [Docker Buildx action](https://github.com/docker/setup-buildx-action)

---

**Last Updated**: 2025-11-14
