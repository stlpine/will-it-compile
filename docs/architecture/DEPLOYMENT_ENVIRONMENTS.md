# Deployment Environments Guide

## Overview

`will-it-compile` can run in different environments. This document explains what works where and how to deploy properly.

## Current MVP Implementation

### Architecture
```
API Server → Docker Client → Docker Daemon → Creates Containers
```

### What It Does
- Uses Docker API to dynamically create containers for each compilation
- Each container is isolated with security constraints
- Containers are ephemeral (created, run, deleted)

## Supported Deployment Environments

### ✅ 1. Local Development (Current)

**Requirements:**
- Docker Desktop or Docker Engine installed
- Docker daemon running

**Works perfectly because:**
- Direct access to Docker socket (`/var/run/docker.sock`)
- Full Docker API available

**Setup:**
```bash
# Verify Docker is running
docker ps

# Build images
make docker-build

# Run server
make run
```

### ✅ 2. Single VM/Server Deployment

**Requirements:**
- Linux VM with Docker installed
- Docker daemon running

**Works because:**
- Same as local development
- Server has Docker installed

**Setup:**
```bash
# On server
sudo apt-get install docker.io
sudo systemctl start docker

# Deploy application
./bin/will-it-compile-api
```

### ✅ 3. Docker Compose

**Requirements:**
- Docker Compose

**Example `docker-compose.yml`:**
```yaml
version: '3.8'
services:
  api:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock  # ⚠️ Required but has security implications
    environment:
      - PORT=8080
```

**⚠️ Security Note:** Mounting Docker socket gives container root-level access to host.

### ❌ 4. Kubernetes (NOT CURRENTLY SUPPORTED)

**Why it doesn't work:**

```
┌─────────────────────────┐
│  Kubernetes Node        │
│                         │
│  ┌──────────────┐       │
│  │  API Pod     │       │
│  │              │       │
│  │  Docker      │───X───│ No Docker daemon!
│  │  Client      │       │
│  └──────────────┘       │
│                         │
│  containerd/CRI-O       │ ← Only this exists
│  (Container Runtime)    │
└─────────────────────────┘
```

**Problems:**
1. No Docker daemon in pods
2. Only container runtime (containerd, CRI-O) at node level
3. Mounting host Docker socket breaks security model
4. Violates pod isolation principles

**Solution:** See [`KUBERNETES_ARCHITECTURE.md`](./KUBERNETES_ARCHITECTURE.md) for K8s-native implementation using Jobs API.

## Quick Comparison

| Environment | Works? | Difficulty | Security | Production Ready? |
|------------|--------|------------|----------|-------------------|
| Local Dev | ✅ Yes | Easy | Medium | No (dev only) |
| Single VM | ✅ Yes | Easy | Medium | Yes (small scale) |
| Docker Compose | ✅ Yes | Easy | Low ⚠️ | No (socket mount) |
| **Kubernetes** | ❌ **No** | **Hard** | **N/A** | **Needs redesign** |
| AWS ECS | ✅ Yes | Medium | Medium | Yes (with DinD) |
| Cloud Run | ❌ No | - | - | No |

## For Kubernetes: Two Approaches

### Approach 1: Kubernetes Jobs API (Recommended)

**Architecture:**
```
API Pod → K8s API → Creates Job → Ephemeral Pod runs compilation
```

**Benefits:**
- Native K8s integration
- Proper security boundaries
- Uses K8s RBAC
- No Docker socket needed

**Implementation:** See [KUBERNETES_ARCHITECTURE.md](./KUBERNETES_ARCHITECTURE.md)

### Approach 2: Worker Pool Pattern

**Architecture:**
```
API Pod → Redis Queue → Pre-created Worker Pods
                        ↓
                    gVisor/Firecracker sandboxing
```

**Benefits:**
- Better performance (no pod startup)
- Higher throughput
- More control

**Drawbacks:**
- More complex
- Requires gVisor or similar

## Migration Path

### Phase 1: MVP (Current) ✅
- Docker-based
- Local development
- Single-server deployment

### Phase 2: Kubernetes Support (Next)
- Implement runtime abstraction
- Add Kubernetes Jobs implementation
- Auto-detect environment

### Phase 3: Optimization
- Worker pool pattern
- Pre-warming
- Better scheduling

### Phase 4: Advanced
- gVisor/Firecracker
- WebAssembly runtime
- Multi-cloud support

## Decision Guide

**Choose Docker Approach (Current) if:**
- ✅ Running on single server/VM
- ✅ Local development
- ✅ Small scale (<100 requests/min)
- ✅ Don't need horizontal scaling

**Need Kubernetes Approach if:**
- ✅ Running on Kubernetes cluster
- ✅ Need horizontal scaling
- ✅ Multi-tenant environment
- ✅ Production-grade deployment

## AWS ECS Deployment

**Note:** AWS ECS supports Docker-in-Docker with some configuration:

```json
{
  "containerDefinitions": [{
    "name": "will-it-compile",
    "image": "will-it-compile/api:latest",
    "privileged": false,
    "dockerSocketMount": true
  }]
}
```

However, this still has security considerations similar to Docker Compose.

## Security Considerations

### Docker Socket Mounting (Current Approach)

**Risk Level:** 🟡 Medium to High

Mounting `/var/run/docker.sock` gives:
- ✅ Full access to create containers
- ❌ Can escape to host system
- ❌ Can access other containers
- ❌ Effectively root on host

**Mitigations:**
1. Run in isolated VM
2. Use Docker user namespace remapping
3. Apply strict resource limits
4. Monitor container creation
5. Use network policies

### Kubernetes Jobs Approach

**Risk Level:** 🟢 Low

- ✅ Proper security boundaries
- ✅ RBAC controls
- ✅ Pod Security Standards
- ✅ Network policies
- ✅ No elevated privileges needed

## Monitoring and Observability

### Docker-based Deployment
```bash
# Monitor containers
docker ps -a

# Check resource usage
docker stats

# View logs
docker logs <container-id>
```

### Kubernetes Deployment
```bash
# Monitor jobs
kubectl get jobs

# Check pods
kubectl get pods -l app=will-it-compile

# View logs
kubectl logs -l job-name=compile-<job-id>
```

## Troubleshooting

### "No Docker daemon" in Kubernetes
**Cause:** Trying to use Docker client in K8s pod
**Solution:** Implement K8s Jobs approach (see KUBERNETES_ARCHITECTURE.md)

### "Permission denied" accessing Docker socket
**Cause:** User doesn't have Docker permissions
**Solution:** Add user to `docker` group: `sudo usermod -aG docker $USER`

### "Cannot connect to Docker daemon"
**Cause:** Docker daemon not running
**Solution:** `sudo systemctl start docker`

## Next Steps

1. **For local development:** Continue using current Docker approach
2. **For production deployment:** Evaluate your environment
3. **For Kubernetes:** Implement Jobs API approach
4. **For high scale:** Consider worker pool pattern

## References

- [KUBERNETES_ARCHITECTURE.md](./KUBERNETES_ARCHITECTURE.md) - Full K8s implementation guide
- [Docker Security](https://docs.docker.com/engine/security/)
- [Kubernetes Jobs](https://kubernetes.io/docs/concepts/workloads/controllers/job/)
- [Pod Security Standards](https://kubernetes.io/docs/concepts/security/pod-security-standards/)
