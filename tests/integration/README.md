# Integration Tests

This directory contains integration tests that verify the complete system behavior with real external dependencies.

## Test Files

| File | Description | Dependencies |
|------|-------------|--------------|
| `api_test.go` | Basic API integration tests | Docker (compiler images) |
| `api_suite_test.go` | Suite-based API tests with setup/teardown | Docker |
| `async_compile_test.go` | Async compilation with virtualized time | Docker |
| `table_driven_test.go` | Comprehensive table-driven test scenarios | Docker |
| `redis_integration_test.go` | Redis storage integration tests | **Redis + Docker** |

## Testing Strategy

### Two-Tier Redis Testing

1. **Unit Tests** (`internal/storage/redis/store_test.go`)
   - Use `miniredis` (in-memory mock)
   - Fast, no external dependencies
   - Run in all environments

2. **Integration Tests** (`redis_integration_test.go`)
   - Use real Redis instance
   - Verify production behavior
   - Auto-skip if Redis unavailable locally
   - Always run in CI with Redis service

## Running Tests

### All Integration Tests

```bash
# Local (requires Docker for compiler images, optional Redis)
go test -v ./tests/integration/

# With Redis (recommended)
docker compose up redis -d
REDIS_ADDR=localhost:6379 go test -v ./tests/integration/

# CI (GitHub Actions - Redis service included)
# Tests run automatically with REDIS_ADDR=localhost:6379
```

### Specific Test Suites

```bash
# API tests only
go test -v -run TestAPI ./tests/integration/

# Redis tests only
docker compose up redis -d
REDIS_ADDR=localhost:6379 go test -v -run TestRedis ./tests/integration/

# Async tests only
go test -v -run TestAsync ./tests/integration/
```

### With Race Detection

```bash
go test -v -race ./tests/integration/
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `REDIS_ADDR` | `localhost:6379` | Redis server address for integration tests |
| `REDIS_PASSWORD` | `` | Redis password (if required) |
| `MINIMAL_IMAGE_VALIDATION` | `false` | Skip extensive Docker image checks (CI optimization) |

## CI/CD

### GitHub Actions Configuration

Integration tests run in CI with the following services:

```yaml
services:
  redis:
    image: redis:7-alpine
    ports:
      - 6379:6379
    options: >-
      --health-cmd "redis-cli ping"
      --health-interval 10s
```

**Environment Variables Set in CI:**
- `REDIS_ADDR=localhost:6379`
- `MINIMAL_IMAGE_VALIDATION=true`

## Test Behavior

### Redis Tests

**Local Development:**
- If Redis is **not available**: Tests are **skipped** with a message
- If Redis **is available**: Tests run against the Redis instance

**CI (GitHub Actions):**
- Redis service is **always available**
- Tests **always run** and must pass

### Skip Behavior Example

```go
func (s *RedisIntegrationSuite) SetupSuite() {
    err := s.client.Ping(ctx).Err()
    if err != nil {
        s.T().Skipf("Redis not available at %s: %v. Skipping Redis integration tests.", s.redisAddr, err)
    }
}
```

## Writing New Integration Tests

### For API Tests

```go
func TestMyFeature(t *testing.T) {
    // Use existing server setup or create new
    server, _ := api.NewServer()
    defer server.Close()

    // Test your feature
    // ...
}
```

### For Redis Tests

Add tests to `RedisIntegrationSuite`:

```go
func (s *RedisIntegrationSuite) TestMyRedisFeature() {
    // s.store is already initialized
    // s.client provides direct Redis access

    // Test your feature
    job := models.CompilationJob{...}
    err := s.store.Store(job)
    assert.NoError(s.T(), err)
}
```

## Best Practices

1. **Clean Up**: Always clean up resources (containers, Redis keys, etc.)
2. **Isolation**: Tests should not depend on each other
3. **Parallel Safe**: Avoid shared state or use unique keys
4. **Fast Feedback**: Keep tests fast; use minimal compiler images
5. **Graceful Skip**: Skip tests when dependencies unavailable (local dev)
6. **CI Required**: Tests must pass in CI with all services available

## Debugging

### View Redis Data During Tests

```bash
# Start Redis
docker compose up redis -d

# Run tests
REDIS_ADDR=localhost:6379 go test -v ./tests/integration/

# Inspect Redis
docker exec -it will-it-compile-redis redis-cli
> KEYS *
> HGETALL job:<job-id>
```

### Run with Verbose Output

```bash
go test -v -run TestRedis ./tests/integration/
```

### Check for Race Conditions

```bash
go test -race -run TestRedisConcurrent ./tests/integration/
```

## References

- [Redis Integration Guide](../../docs/technical/REDIS_INTEGRATION.md)
- [Testing with testify](https://github.com/stretchr/testify)
- [GitHub Actions Services](https://docs.github.com/en/actions/using-containerized-services/about-service-containers)
