# Docker Setup Guide

This document provides instructions for running **will-it-compile** locally using Docker and Docker Compose for development.

**For Production Deployment**: Use Kubernetes. See [`docs/architecture/KUBERNETES_ARCHITECTURE.md`](./docs/architecture/KUBERNETES_ARCHITECTURE.md) and [`deployments/helm/`](./deployments/helm/).

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Quick Start](#quick-start)
3. [Development Setup](#development-setup)
4. [Docker Images](#docker-images)
5. [Architecture](#architecture)
6. [Troubleshooting](#troubleshooting)
7. [Advanced Configuration](#advanced-configuration)

## Prerequisites

### Required Software

- **Docker**: Version 20.10+ ([Install Docker](https://docs.docker.com/get-docker/))
- **Docker Compose**: Version 2.0+ ([Install Docker Compose](https://docs.docker.com/compose/install/))

### System Requirements

- **RAM**: Minimum 4GB (8GB+ recommended for concurrent compilations)
- **Storage**: 2GB for Docker images
- **OS**: Linux, macOS, or Windows with WSL2

### Verify Installation

```bash
# Check Docker version
docker --version

# Check Docker Compose version
docker compose version

# Verify Docker daemon is running
docker ps
```

## Quick Start

The fastest way to get the entire stack running locally:

```bash
# Clone the repository
git clone https://github.com/stlpine/will-it-compile.git
cd will-it-compile

# Build and start all services
docker compose up -d

# View logs
docker compose logs -f

# Access the services
# - Web UI: http://localhost:3000
# - API: http://localhost:8080
# - Health check: http://localhost:8080/health
```

Stop the services:

```bash
docker compose down
```

## Development Setup

### Local Development with Hot Reload

For active development, Docker Compose provides hot-reloading for both backend and frontend:

```bash
# Start development environment
docker compose up

# Or run in detached mode
docker compose up -d

# Watch logs
docker compose logs -f api
docker compose logs -f web

# Rebuild after dependency changes
docker compose up --build

# Stop services
docker compose down
```

### Development Features

- **Hot Reload**: Source code changes automatically reload
  - **API**: Go source in `cmd/`, `internal/`, `pkg/` directories
  - **Web**: React source in `web/src/` directory
- **Debug Logs**: `LOG_LEVEL=debug` enabled
- **Source Mounts**: Code changes reflect immediately without rebuilding

### Individual Service Management

```bash
# Start only the API server
docker compose up api

# Start only the web frontend
docker compose up web

# Rebuild specific service
docker compose up --build api
docker compose up --build web

# View service logs
docker compose logs -f api
docker compose logs -f web

# Execute commands in running container
docker compose exec api sh
docker compose exec web sh
```

### Building Compiler Images

The C++ compiler Docker image is automatically built when you start the stack, but you can build it manually:

```bash
# Build C++ compiler image
cd images/cpp
./build.sh

# Or using docker compose
docker compose build compiler-cpp

# Verify the image
docker images | grep will-it-compile/cpp-gcc
```

## Docker Images

### Image Overview

| Image | Purpose | Base | Size (approx) | Usage |
|-------|---------|------|---------------|-------|
| `will-it-compile-api` | API Server | `alpine:3.19` | ~50MB | Local dev + K8s |
| `will-it-compile-web` | Web Frontend | `nginx:alpine` | ~30MB | Local dev + K8s |
| `will-it-compile/cpp-gcc:13-alpine` | C++ Compiler | `alpine:3.19` | ~200MB | Compilation jobs |

### Building Images for Kubernetes

When deploying to Kubernetes, you'll need to build and push images to a registry:

```bash
# Build API server image
docker build -t your-registry/will-it-compile-api:v1.0.0 .

# Build web frontend image
docker build -t your-registry/will-it-compile-web:v1.0.0 ./web

# Build C++ compiler image
docker build -t your-registry/will-it-compile-cpp:gcc-13 ./images/cpp

# Push to registry
docker push your-registry/will-it-compile-api:v1.0.0
docker push your-registry/will-it-compile-web:v1.0.0
docker push your-registry/will-it-compile-cpp:gcc-13
```

Then update your Kubernetes manifests or Helm values with the image URIs.

### Image Management

```bash
# List all project images
docker images | grep will-it-compile

# Remove all project images
docker images | grep will-it-compile | awk '{print $3}' | xargs docker rmi -f

# Prune unused images
docker image prune -a

# View image layers and size
docker history will-it-compile-api:latest
```

## Architecture

### Local Development Architecture

```
┌─────────────────┐
│   Web Frontend  │ :3000
│  (Vite Dev)     │
└────────┬────────┘
         │
         │ HTTP
         ▼
┌─────────────────┐
│   API Server    │ :8080
│      (Go)       │
└────────┬────────┘
         │
         │ Docker API
         ▼
┌─────────────────┐
│  Docker Daemon  │
│  (Host Socket)  │
└────────┬────────┘
         │
         │ Creates
         ▼
┌─────────────────┐
│ Compiler Image  │
│  (Ephemeral)    │
└─────────────────┘
```

### Docker Compose Services

- **Network**: `will-it-compile` (bridge)
- **Web → API**: Internal service-to-service communication
- **API → Docker**: Host socket mount (`/var/run/docker.sock`)
- **External Access**:
  - Web UI: Port 3000
  - API: Port 8080

### Volume Mounts (Development)

- **API**: Source code directories (`cmd/`, `internal/`, `pkg/`, `configs/`) mounted for hot reload
- **Web**: Source code directories (`src/`, `public/`) mounted for hot reload
- **Both**: Docker socket for container orchestration

## Troubleshooting

### Common Issues

#### 1. Port Already in Use

**Error**: `Bind for 0.0.0.0:8080 failed: port is already allocated`

**Solution**:
```bash
# Find process using the port
lsof -i :8080

# Kill the process
kill -9 <PID>

# Or change the port in docker-compose.yml
ports:
  - "8081:8080"  # Host:Container
```

#### 2. Docker Socket Permission Denied

**Error**: `permission denied while trying to connect to the Docker daemon socket`

**Solution**:
```bash
# Add your user to docker group (Linux)
sudo usermod -aG docker $USER
newgrp docker

# Or run with sudo
sudo docker compose up
```

#### 3. Image Build Failures

**Error**: `failed to solve: failed to fetch`

**Solution**:
```bash
# Clear build cache
docker builder prune -a

# Rebuild without cache
docker compose build --no-cache

# Check internet connectivity
ping google.com
```

#### 4. Container Crashes Immediately

**Error**: Container exits with code 137 (OOM) or 1

**Solution**:
```bash
# View logs
docker compose logs api

# Check Docker daemon memory
docker info | grep Memory

# Increase Docker Desktop memory (macOS/Windows)
# Docker Desktop → Settings → Resources → Memory
```

#### 5. Hot Reload Not Working

**Solution**:
```bash
# Ensure volumes are mounted correctly
docker compose config

# Rebuild containers
docker compose up --build

# For macOS/Windows, check file sharing settings
# Docker Desktop → Settings → Resources → File Sharing
```

### Health Checks

```bash
# Check API health
curl http://localhost:8080/health

# Check web accessibility
curl http://localhost:3000

# View health status in Docker
docker ps --format "table {{.Names}}\t{{.Status}}"
```

### Viewing Logs

```bash
# All services
docker compose logs -f

# Specific service
docker compose logs -f api
docker compose logs -f web

# Last 100 lines
docker compose logs --tail=100 api

# Logs since specific time
docker compose logs --since 2024-01-01T10:00:00 api
```

### Debugging Containers

```bash
# Execute shell in running container
docker compose exec api sh
docker compose exec web sh

# Run one-off command
docker compose run --rm api go version

# Inspect container
docker inspect will-it-compile-api

# View container resource usage
docker stats
```

## Advanced Configuration

### Custom Environment Variables

Create `.env` file in project root:

```env
# API Configuration
PORT=8080
LOG_LEVEL=debug

# Frontend Configuration
VITE_API_URL=http://localhost:8080

# Docker Configuration (advanced)
DOCKER_SOCKET=/var/run/docker.sock
```

Docker Compose automatically loads this file.

### Override Docker Compose Configuration

Create `docker-compose.override.yml` (auto-loaded):

```yaml
version: '3.8'

services:
  api:
    environment:
      - LOG_LEVEL=trace
    ports:
      - "8081:8080"
```

### Multi-Stage Builds

Both Dockerfiles use multi-stage builds:

1. **Builder Stage**: Compiles/builds application with all dev dependencies
2. **Runtime Stage**: Minimal image with only runtime dependencies

Benefits:
- Smaller image sizes (50-70% reduction)
- Faster deployments
- Reduced attack surface

### Production Deployment Options

#### Option 1: Kubernetes (Recommended)

For production, deploy using Kubernetes:

```bash
# Using Helm (recommended)
cd deployments/helm
helm install will-it-compile ./will-it-compile-chart \
  --namespace will-it-compile \
  --create-namespace

# Or using kubectl with manifests
kubectl apply -f deployments/k8s/
```

See:
- [`docs/architecture/KUBERNETES_ARCHITECTURE.md`](./docs/architecture/KUBERNETES_ARCHITECTURE.md)
- [`deployments/DEPLOYMENT_GUIDE.md`](./deployments/DEPLOYMENT_GUIDE.md)

#### Option 2: Single-Server Docker Deployment

For simple single-server deployments without Kubernetes:

```bash
# Build images
docker build -t will-it-compile-api:latest .
docker build -t will-it-compile-web:latest ./web

# Run API server
docker run -d \
  --name will-it-compile-api \
  --restart unless-stopped \
  -p 8080:8080 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -e LOG_LEVEL=info \
  will-it-compile-api:latest

# Run web frontend
docker run -d \
  --name will-it-compile-web \
  --restart unless-stopped \
  -p 80:80 \
  --link will-it-compile-api:api \
  will-it-compile-web:latest
```

**Note**: This approach is suitable for:
- Development/testing environments
- Small-scale single-server deployments
- Quick prototypes

For production with high availability, use Kubernetes.

## Useful Commands

### Docker Compose

```bash
# Start services
docker compose up -d

# Stop services (keep containers)
docker compose stop

# Stop and remove containers
docker compose down

# Stop and remove containers + volumes
docker compose down -v

# Rebuild images
docker compose build

# View running services
docker compose ps

# View service configuration
docker compose config
```

### Docker

```bash
# List all containers
docker ps -a

# Remove stopped containers
docker container prune

# List networks
docker network ls

# Inspect network
docker network inspect will-it-compile

# View container resource usage
docker stats

# Export container logs
docker logs will-it-compile-api > api.log 2>&1
```

## Performance Optimization

### Development

- Use `.dockerignore` to exclude unnecessary files
- Mount only essential directories
- Use `docker compose up` without `-d` to see logs immediately

### Image Building

- Enable BuildKit: `DOCKER_BUILDKIT=1 docker build`
- Use layer caching effectively
- Minimize image layers
- Use specific base image tags (not `latest`)

### Monitoring

```bash
# Real-time resource usage
docker stats

# Container logs with timestamps
docker compose logs -f -t

# Disk usage
docker system df

# Detailed disk usage
docker system df -v
```

## Next Steps

- **Local Development**: Start with `docker compose up` and begin coding
- **Production Deployment**: See [`docs/architecture/KUBERNETES_ARCHITECTURE.md`](./docs/architecture/KUBERNETES_ARCHITECTURE.md)
- **API Usage**: Review [README.md](./README.md) for API documentation
- **Architecture**: Check [CLAUDE.md](./CLAUDE.md) for project structure
- **Frontend Development**: See [web/README.md](./web/README.md)

## Support

For issues or questions:
- Check [Troubleshooting](#troubleshooting) section above
- Review Docker logs: `docker compose logs`
- Open an issue on GitHub
- Consult Docker documentation: https://docs.docker.com

---

**Last Updated**: 2025-11-17
**Docker Version**: 20.10+
**Docker Compose Version**: 2.0+
**Purpose**: Local development environment
**Production**: Deploy with Kubernetes (see docs/architecture/)
