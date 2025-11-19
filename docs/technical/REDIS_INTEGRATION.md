# Redis Integration Guide

This document explains the Redis integration in will-it-compile, including architecture, configuration, and usage.

## Overview

Redis integration provides **persistent job storage** for the will-it-compile service, replacing the in-memory storage used in development. This enables:

- ✅ **Persistence**: Jobs survive server restarts
- ✅ **Horizontal scaling**: Multiple API instances can share job state
- ✅ **TTL management**: Automatic cleanup of old jobs
- ✅ **Production readiness**: Suitable for multi-instance deployments

## Architecture

### Storage Abstraction

The project uses a `JobStore` interface to abstract storage implementations:

```go
type JobStore interface {
    Store(job models.CompilationJob) error
    Get(jobID string) (models.CompilationJob, bool)
    StoreResult(jobID string, result models.CompilationResult) error
    GetResult(jobID string) (models.CompilationResult, bool)
    Close() error
}
```

**Implementations:**
- `internal/storage/memory/store.go` - In-memory (development)
- `internal/storage/redis/store.go` - Redis (production)

### Redis Data Model

Redis stores jobs and results using **hashes** for efficient structured storage:

```
Key Pattern                    Type    TTL    Purpose
---------------------------------------------------------------------------
job:{job_id}                   Hash    24h    Job metadata (status, timestamps)
result:{job_id}                Hash    24h    Compilation result
job:index:status:{status}      Set     24h    Jobs by status (queued/processing/completed)
```

**Job Hash Fields:**
- `id` - Job UUID
- `request` - JSON-encoded CompilationRequest
- `status` - Job status (queued/processing/completed/failed)
- `created_at` - RFC3339 timestamp
- `started_at` - RFC3339 timestamp (nullable)
- `completed_at` - RFC3339 timestamp (nullable)

**Result Hash Fields:**
- `success` - Boolean
- `compiled` - Boolean
- `stdout` - Standard output
- `stderr` - Standard error
- `exit_code` - Integer
- `duration` - Nanoseconds

## Configuration

### Environment Variables

```bash
# Enable/disable Redis
REDIS_ENABLED=true              # Set to 'false' for in-memory storage

# Connection
REDIS_ADDR=localhost:6379       # Redis server address
REDIS_PASSWORD=                 # Password (empty if no auth)
REDIS_DB=0                      # Database number (0-15)

# Performance
REDIS_POOL_SIZE=20              # Connection pool size
REDIS_JOB_TTL_HOURS=24          # Time-to-live for jobs

# Worker Pool
MAX_WORKERS=5                   # Concurrent workers
QUEUE_SIZE=100                  # Job queue buffer size
```

### Using .env File

1. Copy the example file:
   ```bash
   cp .env.example .env
   ```

2. Edit `.env` with your settings:
   ```bash
   REDIS_ENABLED=true
   REDIS_ADDR=localhost:6379
   ```

3. The server automatically loads these on startup.

### Docker Compose Configuration

The `docker-compose.yml` includes a Redis service:

```yaml
services:
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    command: redis-server --appendonly yes
    volumes:
      - redis-data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s

  api:
    environment:
      - REDIS_ENABLED=true
      - REDIS_ADDR=redis:6379
    depends_on:
      redis:
        condition: service_healthy
```

**Key features:**
- AOF persistence (`--appendonly yes`)
- Volume for data persistence
- Health checks for dependency management

## Usage

### Development (In-Memory)

For local development without Redis:

```bash
# Run with in-memory storage
REDIS_ENABLED=false go run cmd/api/main.go
```

Or leave Redis disabled in `.env`:
```bash
REDIS_ENABLED=false
```

### Development (With Redis)

Using Docker Compose:

```bash
# Start Redis + API
docker compose up

# Logs
docker compose logs -f redis
docker compose logs -f api

# Stop
docker compose down

# Clean up volumes
docker compose down -v
```

### Production Deployment

**Option 1: Standalone Redis**

```bash
# Install Redis
apt-get install redis-server

# Configure
vi /etc/redis/redis.conf
# Set: appendonly yes, maxmemory-policy allkeys-lru

# Start API
REDIS_ENABLED=true \
REDIS_ADDR=localhost:6379 \
REDIS_JOB_TTL_HOURS=24 \
./bin/will-it-compile-api
```

**Option 2: Redis Cloud (Managed)**

```bash
# Use managed Redis service
REDIS_ENABLED=true \
REDIS_ADDR=redis-12345.cloud.redislabs.com:12345 \
REDIS_PASSWORD=your-password \
./bin/will-it-compile-api
```

**Option 3: Kubernetes**

See `deployments/helm/` for Helm chart with Redis StatefulSet.

## Testing

### Unit Tests

Tests use `miniredis` (in-memory Redis mock):

```bash
# Run Redis storage tests
go test ./internal/storage/redis/

# Run all storage tests
go test ./internal/storage/...

# With coverage
go test -cover ./internal/storage/redis/
```

### Integration Tests

Test with real Redis:

```bash
# Start Redis
docker run -d -p 6379:6379 redis:7-alpine

# Run tests
REDIS_ENABLED=true go test -v ./tests/integration/

# Clean up
docker stop $(docker ps -q --filter ancestor=redis:7-alpine)
```

### Manual Testing

```bash
# Start server with Redis
docker compose up

# Submit job
curl -X POST http://localhost:8080/api/v1/compile \
  -H "Content-Type: application/json" \
  -d '{
    "source_code": "int main() { return 0; }",
    "language": "cpp",
    "compiler": "gcc-13"
  }'

# Check Redis data
docker exec -it will-it-compile-redis redis-cli
> KEYS job:*
> HGETALL job:<job-id>
> HGETALL result:<job-id>
> TTL job:<job-id>
```

## Monitoring

### Redis Metrics

```bash
# Connect to Redis CLI
docker exec -it will-it-compile-redis redis-cli

# Check stats
> INFO stats
> INFO memory
> INFO clients

# Monitor commands
> MONITOR

# Count keys
> DBSIZE
```

### Application Logs

The server logs Redis connection status on startup:

```
2025-01-19 10:30:00 Starting will-it-compile API server
2025-01-19 10:30:00 Environment: production
2025-01-19 10:30:00 Port: 8080
2025-01-19 10:30:00 Redis enabled: true
2025-01-19 10:30:00 Redis address: redis:6379
2025-01-19 10:30:00 Job TTL: 24h0m0s
2025-01-19 10:30:00 Redis job store initialized successfully (TTL: 24h0m0s)
```

## Troubleshooting

### Connection Refused

**Symptom:** `failed to connect to Redis at localhost:6379: connection refused`

**Solutions:**
1. Ensure Redis is running: `docker ps` or `redis-cli ping`
2. Check `REDIS_ADDR` is correct
3. Verify firewall rules
4. Check Redis logs: `docker logs will-it-compile-redis`

### Authentication Failed

**Symptom:** `NOAUTH Authentication required`

**Solution:** Set `REDIS_PASSWORD` environment variable:
```bash
REDIS_PASSWORD=your-password
```

### Jobs Not Persisting

**Symptom:** Jobs disappear after server restart

**Check:**
1. Redis AOF enabled: `docker exec ... redis-cli CONFIG GET appendonly`
2. Volume mounted: `docker volume ls | grep redis`
3. TTL not too short: Check `REDIS_JOB_TTL_HOURS`

### High Memory Usage

**Symptom:** Redis using too much memory

**Solutions:**
1. Reduce TTL: `REDIS_JOB_TTL_HOURS=12`
2. Set max memory policy in Redis config:
   ```
   maxmemory 256mb
   maxmemory-policy allkeys-lru
   ```
3. Monitor with `INFO memory`

### Slow Performance

**Symptom:** High latency for job operations

**Check:**
1. Redis connection pool: Increase `REDIS_POOL_SIZE=50`
2. Network latency: Use `redis-cli --latency`
3. Redis slow log: `SLOWLOG GET 10`

## Migration from In-Memory

### Step 1: Enable Redis (Shadow Mode)

Start with Redis enabled but verify behavior:

```bash
REDIS_ENABLED=true
REDIS_ADDR=localhost:6379
```

### Step 2: Monitor

Check logs for Redis errors:
```bash
tail -f /var/log/will-it-compile/api.log | grep -i redis
```

### Step 3: Validate

Test job lifecycle:
1. Submit job
2. Check Redis has data
3. Restart server
4. Verify job still exists

### Step 4: Production

Deploy with Redis fully enabled.

## Best Practices

### Development
- ✅ Use in-memory storage (`REDIS_ENABLED=false`)
- ✅ Use Docker Compose for local Redis testing
- ✅ Keep TTL short (1-2 hours) to avoid clutter

### Production
- ✅ Use managed Redis (AWS ElastiCache, Redis Cloud)
- ✅ Enable AOF persistence
- ✅ Set appropriate TTL (24-48 hours)
- ✅ Monitor memory usage
- ✅ Use connection pooling (20-50 connections)
- ✅ Enable health checks
- ✅ Set up alerts for connection failures
- ✅ Regular backups (via Redis BGSAVE or RDB snapshots)

### Security
- ✅ Use authentication (`requirepass` in Redis config)
- ✅ Bind to private network only
- ✅ Use TLS for connections (if supported)
- ✅ Limit network access via firewall

## Performance Tuning

### Connection Pool

Adjust based on worker count:
```bash
# Formula: POOL_SIZE = MAX_WORKERS * 2 + 10
MAX_WORKERS=20
REDIS_POOL_SIZE=50
```

### TTL Optimization

Balance storage vs. persistence needs:
- **Short TTL (6-12h)**: Less storage, faster cleanup
- **Long TTL (48-72h)**: Better for debugging, audit trails

### Memory Limits

Configure Redis `maxmemory` and eviction policy:
```conf
maxmemory 512mb
maxmemory-policy allkeys-lru  # Evict least recently used
```

## Future Enhancements

Phase 3 roadmap includes:

- [ ] **Redis Streams** for job queue (replace in-memory channel)
- [ ] **Pub/Sub** for real-time job updates
- [ ] **Redis Cluster** for horizontal scaling
- [ ] **Sentinel** for high availability
- [ ] **Prometheus metrics** for Redis monitoring

## References

- [Redis Documentation](https://redis.io/documentation)
- [go-redis GitHub](https://github.com/redis/go-redis)
- [Redis Best Practices](https://redis.io/docs/manual/patterns/)
- [Project Architecture](../architecture/IMPLEMENTATION_PLAN.md)
