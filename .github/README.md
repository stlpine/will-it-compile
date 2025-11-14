# GitHub Actions Workflows

This directory contains GitHub Actions workflows for continuous integration and code quality checks.

## Workflows

### 1. Pull Request Checks (`pull_request.yml`)

**Trigger**: On pull requests to any branch

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
- **Docker**: Validates Docker images
  - Builds Docker images
  - Tests Docker images with sample code
- **Status Check**: Final check to ensure all jobs passed

**Concurrency**: Auto-cancels previous runs when new commits are pushed to the same PR (saves CI resources)

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
- **Security**: Security scanning
  - Runs `gosec` security scanner
  - Uploads SARIF results to GitHub Security
- **Verify**: Dependency verification
  - Verifies Go modules
  - Checks for vulnerabilities using `govulncheck`
- **Success**: Final status check

**Concurrency**: Does NOT cancel previous runs - ensures every main branch commit is fully validated

**Artifacts**:
- Coverage reports (retained for 30 days)

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

Workflows upload artifacts for analysis:
- **CI Workflow**: Coverage reports (30 days)

Binary artifacts are not uploaded as they can be easily rebuilt with `make build`.

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

### Concurrency Cancellation

**Problem**: PR workflow cancels my run

**Solution**: This is expected behavior. When you push new commits to a PR, the previous workflow run is cancelled to save CI resources. Only the latest commit's workflow will run to completion.

---

## Future Enhancements

Potential workflow improvements for later:

- [ ] Add code coverage reporting to PRs (e.g., Codecov)
- [ ] Deploy to staging environment on main branch
- [ ] Add performance benchmarking
- [ ] Scheduled security scans (weekly)
- [ ] Automated dependency updates (Dependabot)
- [ ] Slack/Discord notifications for build failures
- [ ] Release automation (when release strategy is decided)

---

## References

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [golangci-lint GitHub Action](https://github.com/golangci/golangci-lint-action)
- [Go setup action](https://github.com/actions/setup-go)

---

**Last Updated**: 2025-11-14
**Workflows**: Pull Request Checks, CI Pipeline
