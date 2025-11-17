# Docker Quick Start

Quick reference for running **will-it-compile** with Docker locally.

**Production Deployment?** Use Kubernetes. See [DOCKER.md](./DOCKER.md) for details.

## TL;DR

```bash
# Start everything (local development)
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

### Debugging

```bash
# Shell into containers
docker compose exec api sh
docker compose exec web sh

# View resource usage
docker stats

# Check health
curl http://localhost:8080/health
curl http://localhost:3000/
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
├── docker-compose.yml            # Local development stack
├── .dockerignore                 # Exclude files from API build
├── web/
│   ├── Dockerfile                # Production web (nginx) - for K8s
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
| `web` | 3000 | React frontend (Vite dev server) |
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

## Production Deployment

**Don't use Docker Compose for production!** Use Kubernetes instead:

```bash
# Build and push images to your registry
docker build -t your-registry/will-it-compile-api:v1.0.0 .
docker build -t your-registry/will-it-compile-web:v1.0.0 ./web
docker build -t your-registry/will-it-compile-cpp:gcc-13 ./images/cpp

docker push your-registry/will-it-compile-api:v1.0.0
docker push your-registry/will-it-compile-web:v1.0.0
docker push your-registry/will-it-compile-cpp:gcc-13

# Deploy with Helm
cd deployments/helm
helm install will-it-compile ./will-it-compile-chart \
  --namespace will-it-compile \
  --create-namespace
```

See:
- [`docs/architecture/KUBERNETES_ARCHITECTURE.md`](./docs/architecture/KUBERNETES_ARCHITECTURE.md)
- [`deployments/DEPLOYMENT_GUIDE.md`](./deployments/DEPLOYMENT_GUIDE.md)

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

## Need More Help?

- **Comprehensive Guide**: [DOCKER.md](./DOCKER.md)
- **Kubernetes Deployment**: [docs/architecture/KUBERNETES_ARCHITECTURE.md](./docs/architecture/KUBERNETES_ARCHITECTURE.md)
- **API Documentation**: [README.md](./README.md)
- **Project Architecture**: [CLAUDE.md](./CLAUDE.md)

---

**Local Development Only** - For production, use Kubernetes
