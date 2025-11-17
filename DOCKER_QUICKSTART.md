# Docker Quick Start

Quick reference for running **will-it-compile** with Docker.

## TL;DR

```bash
# Start everything (development)
docker compose up -d

# View logs
docker compose logs -f

# Stop everything
docker compose down
```

Access:
- **Web UI**: http://localhost:3000
- **API**: http://localhost:8080
- **Health**: http://localhost:8080/health

## Common Commands

### Development

```bash
# Start with logs visible
docker compose up

# Start in background
docker compose up -d

# Rebuild after code changes
docker compose up --build

# View logs
docker compose logs -f api     # API server
docker compose logs -f web     # Web frontend

# Stop services
docker compose down
```

### Production

```bash
# Start production stack
docker compose -f docker-compose.prod.yml up -d

# View production logs
docker compose -f docker-compose.prod.yml logs -f

# Stop production stack
docker compose -f docker-compose.prod.yml down
```

### Debugging

```bash
# Shell into containers
docker compose exec api sh
docker compose exec web sh

# View resource usage
docker stats

# Check health
curl http://localhost:8080/health
curl http://localhost:3000/health
```

### Cleanup

```bash
# Stop and remove containers
docker compose down

# Remove containers + volumes
docker compose down -v

# Remove all project images
docker images | grep will-it-compile | awk '{print $3}' | xargs docker rmi -f

# Full cleanup (careful!)
docker system prune -a
```

## File Structure

```
.
├── Dockerfile                    # API server (Go)
├── docker-compose.yml            # Development stack
├── docker-compose.prod.yml       # Production stack
├── .dockerignore                 # Exclude files from API build
├── web/
│   ├── Dockerfile                # Production web (nginx)
│   ├── Dockerfile.dev            # Development web (hot reload)
│   ├── nginx.conf                # Nginx configuration
│   └── .dockerignore             # Exclude files from web build
└── images/
    └── cpp/
        └── Dockerfile            # C++ compiler image
```

## Services

| Service | Port | Purpose |
|---------|------|---------|
| `api` | 8080 | Go API server |
| `web` | 3000 (dev) / 80 (prod) | React frontend |
| `compiler-cpp` | - | C++ compiler image builder |

## Environment Variables

Create `.env` file:

```env
# API
PORT=8080
LOG_LEVEL=debug

# Frontend
VITE_API_URL=http://localhost:8080
```

## Troubleshooting

### Port conflicts
```bash
# Change port in docker-compose.yml
ports:
  - "8081:8080"  # Use 8081 instead of 8080
```

### Permission issues
```bash
# Add user to docker group (Linux)
sudo usermod -aG docker $USER
newgrp docker
```

### Build failures
```bash
# Clear cache and rebuild
docker builder prune -a
docker compose build --no-cache
```

### Hot reload not working
```bash
# Rebuild with fresh volumes
docker compose down -v
docker compose up --build
```

## Need More Help?

See **[DOCKER.md](./DOCKER.md)** for comprehensive documentation.

## Quick Test

```bash
# Start services
docker compose up -d

# Wait for startup
sleep 10

# Test API
curl http://localhost:8080/health

# Test compilation (example)
curl -X POST http://localhost:8080/api/v1/compile \
  -H "Content-Type: application/json" \
  -d '{
    "language": "cpp",
    "source_code": "int main() { return 0; }",
    "compiler": "g++",
    "compiler_flags": ["-std=c++17"]
  }'

# View web UI
open http://localhost:3000  # macOS
xdg-open http://localhost:3000  # Linux
```

---

For detailed documentation, see [DOCKER.md](./DOCKER.md)
