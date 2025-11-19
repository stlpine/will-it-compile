# Deployment Guide

This guide covers deploying will-it-compile to both local development and production Kubernetes environments.

## Prerequisites

### All Environments
- Kubernetes cluster 1.24+
- Helm 3.0+
- `kubectl` configured to access your cluster
- Compiler Docker images built and available

### Redis Storage

The Helm chart includes **built-in Redis support** for persistent job storage. This is enabled by default and provides:

- ✅ **Persistent storage**: Jobs survive pod restarts
- ✅ **Horizontal scaling**: Multiple API instances share job state
- ✅ **Production ready**: Automatic TTL cleanup and connection pooling

**Configuration Options:**

1. **Embedded Redis** (Default): Chart deploys Redis StatefulSet
   - Development: Ephemeral storage (no persistence)
   - Production: Persistent storage with authentication

2. **In-Memory Mode**: Disable Redis for single-replica testing only
   - Set `redis.enabled=false` in values-dev.yaml
   - Not suitable for production or multi-replica deployments

See the **Redis Configuration** section below for details.

### Compiler Images

**Option 1: Use Official Docker Hub Images** (Recommended)

The project uses official Docker images from Docker Hub:

```bash
# API server image (published automatically)
stlpine/will-it-compile-api:latest

# Compiler images (official Docker Hub images)
gcc:13                    # C/C++ (Debian-based)
golang:1.22-alpine        # Go (Alpine-based)
rust:1.75-alpine          # Rust (Alpine-based)
```

No additional setup needed - Helm charts use these images by default.

**Option 2: Build and Push Custom API Server Image**

If you need to build the API server from source:

```bash
# Build API server image
docker build -f Dockerfile -t your-registry.io/will-it-compile/api:latest .

# Push to your registry
docker push your-registry.io/will-it-compile/api:latest

# Update Helm values to use your registry
helm install will-it-compile ./deployments/helm/will-it-compile \
  --set image.repository=your-registry.io/will-it-compile/api
```

**Note**: Compiler images are pulled directly from Docker Hub. No custom builds required.

## Local Development Deployment

### Option 1: Using kind (Kubernetes in Docker)

```bash
# Create a kind cluster
kind create cluster --name will-it-compile

# Option A: Use public Docker Hub images (automatic pull)
helm install will-it-compile ./deployments/helm/will-it-compile \
  --values ./deployments/helm/will-it-compile/values-dev.yaml

# Option B: Pre-load local images for offline testing
kind load docker-image stlpine/will-it-compile-api:latest --name will-it-compile
kind load docker-image gcc:13 --name will-it-compile
kind load docker-image golang:1.22-alpine --name will-it-compile
kind load docker-image rust:1.75-alpine --name will-it-compile

# Access the service
kubectl port-forward svc/will-it-compile 8080:80

# Test
curl http://localhost:8080/health
```

### Option 2: Using Minikube

```bash
# Start minikube
minikube start

# Use minikube's Docker daemon
eval $(minikube docker-env)

# Build images (they'll be available in minikube)
make docker-build

# Install chart
helm install will-it-compile ./deployments/helm/will-it-compile \
  --values ./deployments/helm/will-it-compile/values-dev.yaml

# Access via NodePort
minikube service will-it-compile
```

## Production Deployment

### 1. Container Registry Options

**Option A: Use Docker Hub (stlpine) - Recommended for Quick Start**

Images are automatically published on every main branch commit. No setup needed.

```bash
# Images used:
# - stlpine/will-it-compile-api:latest (API server, auto-published)
# - gcc:13 (official C/C++ compiler, from Docker Hub)
# - golang:1.22-alpine (official Go compiler, from Docker Hub)
# - rust:1.75-alpine (official Rust compiler, from Docker Hub)

# Deploy directly using Helm defaults
helm install will-it-compile ./deployments/helm/will-it-compile \
  --namespace will-it-compile --create-namespace \
  --values ./deployments/helm/will-it-compile/values-production.yaml
```

**Option B: Use Your Own Private Registry**

```bash
# Build and push API server image
docker build -t your-registry.io/will-it-compile/api:v1.0.0 .
docker push your-registry.io/will-it-compile/api:v1.0.0

# Note: Compiler images use official Docker Hub images (gcc:13, golang:1.22-alpine, rust:1.75-alpine)
# No need to build or push compiler images unless you need custom modifications

# Deploy with custom registry
helm install will-it-compile ./deployments/helm/will-it-compile \
  --namespace will-it-compile --create-namespace \
  --values ./deployments/helm/will-it-compile/values-production.yaml \
  --set image.repository=your-registry.io/will-it-compile/api \
  --set compilerImages.cpp.repository=your-registry.io/will-it-compile/cpp-gcc
```

### 2. Create Namespace

```bash
kubectl create namespace will-it-compile
```

### 3. Create Image Pull Secret (if using private registry)

```bash
kubectl create secret docker-registry regcred \
  --namespace will-it-compile \
  --docker-server=your-registry.io \
  --docker-username=your-username \
  --docker-password=your-password \
  --docker-email=your-email@example.com
```

### 4. Update Production Values

Edit `deployments/helm/will-it-compile/values-production.yaml`:

```yaml
image:
  repository: your-registry.io/will-it-compile/api
  tag: "v1.0.0"

compilerImages:
  cpp:
    repository: your-registry.io/will-it-compile/cpp-gcc
    tag: "13-alpine"

imagePullSecrets:
  - name: regcred

ingress:
  enabled: true
  hosts:
    - host: compile.your-domain.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: will-it-compile-tls
      hosts:
        - compile.your-domain.com
```

### 5. Install with Production Values

```bash
helm install will-it-compile ./deployments/helm/will-it-compile \
  --namespace will-it-compile \
  --values ./deployments/helm/will-it-compile/values-production.yaml
```

### 6. Verify Deployment

```bash
# Check all resources
kubectl get all -n will-it-compile

# Check pods are running
kubectl get pods -n will-it-compile

# Check RBAC
kubectl get role,rolebinding,serviceaccount -n will-it-compile

# Check logs
kubectl logs -n will-it-compile -l app=will-it-compile

# Test health endpoint
kubectl port-forward -n will-it-compile svc/will-it-compile 8080:80
curl http://localhost:8080/health
```

## Redis Configuration

### Embedded Redis

The Helm chart deploys Redis as a StatefulSet with automatic configuration:

**Development** (values-dev.yaml):
```yaml
redis:
  enabled: true
  persistence:
    enabled: false  # Ephemeral storage
  auth:
    enabled: false  # No authentication
  resources:
    limits:
      cpu: 200m
      memory: 256Mi
```

**Production** (values-production.yaml):
```yaml
redis:
  enabled: true
  persistence:
    enabled: true
    size: 20Gi  # Persistent storage
  auth:
    enabled: true  # Password authentication (auto-generated)
  resources:
    limits:
      cpu: 1000m
      memory: 1Gi
```

**Deployment:**
```bash
# Development
helm install will-it-compile ./deployments/helm/will-it-compile \
  --namespace will-it-compile --create-namespace \
  --values ./deployments/helm/will-it-compile/values-dev.yaml

# Production
helm install will-it-compile ./deployments/helm/will-it-compile \
  --namespace will-it-compile --create-namespace \
  --values ./deployments/helm/will-it-compile/values-production.yaml
```

**Verify Redis Deployment:**
```bash
# Check Redis pod
kubectl get pods -n will-it-compile -l app.kubernetes.io/component=redis

# Check Redis logs
kubectl logs -n will-it-compile -l app.kubernetes.io/component=redis

# Test Redis connection (from API pod)
kubectl exec -n will-it-compile -it <api-pod-name> -- sh -c 'redis-cli -h will-it-compile-redis-client ping'

# Retrieve Redis password (production with auth enabled)
kubectl get secret will-it-compile-redis-secret -n will-it-compile -o jsonpath='{.data.password}' | base64 -d
```

### In-Memory Mode (Testing Only)

For local testing with a single replica, you can disable Redis:

```bash
helm install will-it-compile ./deployments/helm/will-it-compile \
  --namespace will-it-compile --create-namespace \
  --values ./deployments/helm/will-it-compile/values-dev.yaml \
  --set redis.enabled=false

# Or in values-dev.yaml:
redis:
  enabled: false
```

⚠️ **Warning**: In-memory mode is NOT suitable for:
- Production deployments
- Multiple replicas (jobs won't be shared)
- Long-running deployments (jobs lost on restart)

### Redis Monitoring

**Check Redis storage:**
```bash
# Connect to Redis CLI
kubectl exec -n will-it-compile -it <redis-pod-name> -- redis-cli

# Inside Redis CLI:
> INFO memory
> DBSIZE
> KEYS job:*
> TTL job:<job-id>
> HGETALL job:<job-id>
```

**Monitor API logs for Redis:**
```bash
kubectl logs -n will-it-compile -l app=will-it-compile | grep -i redis

# Expected output on startup:
# Initializing Redis job store at will-it-compile-redis-client:6379
# Redis job store initialized successfully (TTL: 24h0m0s)
```

## Cloud-Specific Deployments

### Google Kubernetes Engine (GKE)

```bash
# Create GKE cluster
gcloud container clusters create will-it-compile \
  --zone us-central1-a \
  --num-nodes 3 \
  --machine-type n1-standard-2

# Get credentials
gcloud container clusters get-credentials will-it-compile --zone us-central1-a

# Install chart
helm install will-it-compile ./deployments/helm/will-it-compile \
  --namespace will-it-compile \
  --create-namespace \
  --values ./deployments/helm/will-it-compile/values-production.yaml \
  --set image.repository=gcr.io/your-project/will-it-compile/api \
  --set compilerImages.cpp.repository=gcr.io/your-project/will-it-compile/cpp-gcc
```

### Amazon EKS

```bash
# Create EKS cluster (using eksctl)
eksctl create cluster \
  --name will-it-compile \
  --region us-west-2 \
  --nodegroup-name standard-workers \
  --node-type t3.medium \
  --nodes 3

# Install chart
helm install will-it-compile ./deployments/helm/will-it-compile \
  --namespace will-it-compile \
  --create-namespace \
  --values ./deployments/helm/will-it-compile/values-production.yaml \
  --set image.repository=YOUR_ACCOUNT.dkr.ecr.us-west-2.amazonaws.com/will-it-compile/api \
  --set compilerImages.cpp.repository=YOUR_ACCOUNT.dkr.ecr.us-west-2.amazonaws.com/will-it-compile/cpp-gcc
```

### Azure Kubernetes Service (AKS)

```bash
# Create resource group
az group create --name will-it-compile-rg --location eastus

# Create AKS cluster
az aks create \
  --resource-group will-it-compile-rg \
  --name will-it-compile \
  --node-count 3 \
  --node-vm-size Standard_DS2_v2 \
  --enable-managed-identity

# Get credentials
az aks get-credentials --resource-group will-it-compile-rg --name will-it-compile

# Install chart
helm install will-it-compile ./deployments/helm/will-it-compile \
  --namespace will-it-compile \
  --create-namespace \
  --values ./deployments/helm/will-it-compile/values-production.yaml \
  --set image.repository=yourregistry.azurecr.io/will-it-compile/api \
  --set compilerImages.cpp.repository=yourregistry.azurecr.io/will-it-compile/cpp-gcc
```

## Upgrading

```bash
# Update with new values
helm upgrade will-it-compile ./deployments/helm/will-it-compile \
  --namespace will-it-compile \
  --values ./deployments/helm/will-it-compile/values-production.yaml

# Update to new image version
helm upgrade will-it-compile ./deployments/helm/will-it-compile \
  --namespace will-it-compile \
  --reuse-values \
  --set image.tag=v1.1.0
```

## Rollback

```bash
# List releases
helm history will-it-compile -n will-it-compile

# Rollback to previous version
helm rollback will-it-compile -n will-it-compile

# Rollback to specific revision
helm rollback will-it-compile 3 -n will-it-compile
```

## Monitoring

### Check Compilation Jobs

```bash
# List all compilation jobs
kubectl get jobs -n will-it-compile -l managed-by=will-it-compile

# Get job details
kubectl describe job compile-<job-id> -n will-it-compile

# View job logs
kubectl logs -n will-it-compile -l job-id=<job-id>

# Clean up old jobs (handled automatically by TTL, but manual cleanup if needed)
kubectl delete jobs -n will-it-compile -l managed-by=will-it-compile --field-selector status.successful=1
```

### View Application Logs

```bash
# Stream logs from all pods
kubectl logs -n will-it-compile -l app=will-it-compile -f --all-containers=true

# Logs from specific pod
kubectl logs -n will-it-compile <pod-name>
```

### Check Resource Usage

```bash
# Pod resource usage
kubectl top pods -n will-it-compile

# Node resource usage
kubectl top nodes

# HPA status (if enabled)
kubectl get hpa -n will-it-compile
```

## Troubleshooting

### Pods Not Starting

```bash
# Check pod status
kubectl describe pod -n will-it-compile <pod-name>

# Common issues:
# 1. ImagePullBackOff - check image exists and pull secrets are correct
# 2. CrashLoopBackOff - check logs for application errors
# 3. Pending - check resource availability and node selectors
```

### RBAC Errors

```bash
# Check service account
kubectl get sa will-it-compile -n will-it-compile

# Check role
kubectl get role -n will-it-compile

# Check role binding
kubectl describe rolebinding will-it-compile-compiler-binding -n will-it-compile

# Test permissions
kubectl auth can-i create jobs --as=system:serviceaccount:will-it-compile:will-it-compile -n will-it-compile
```

### Compilation Jobs Failing

```bash
# Check if compiler image exists
kubectl run test-compiler --image=gcc:13 --restart=Never -- /bin/sh -c "echo test"
kubectl logs test-compiler
kubectl delete pod test-compiler

# Check job pod events
kubectl describe pod -n will-it-compile -l job-name=compile-<job-id>

# Verify ConfigMap creation
kubectl get configmaps -n will-it-compile -l managed-by=will-it-compile
```

### Network Issues

```bash
# Check network policy
kubectl get networkpolicy -n will-it-compile

# Temporarily disable network policy for debugging
helm upgrade will-it-compile ./deployments/helm/will-it-compile \
  --namespace will-it-compile \
  --reuse-values \
  --set networkPolicy.enabled=false

# Test connectivity from pod
kubectl exec -n will-it-compile -it <pod-name> -- wget -O- http://kubernetes.default.svc
```

## Uninstalling

```bash
# Uninstall release (keeps namespace)
helm uninstall will-it-compile -n will-it-compile

# Clean up compilation jobs and ConfigMaps
kubectl delete jobs,configmaps -n will-it-compile -l managed-by=will-it-compile

# Delete namespace (if desired)
kubectl delete namespace will-it-compile
```

## Performance Tuning

### Adjust Resource Limits

For high-traffic environments:

```yaml
# values-production.yaml
resources:
  limits:
    cpu: 4000m
    memory: 2Gi
  requests:
    cpu: 1000m
    memory: 1Gi

autoscaling:
  enabled: true
  minReplicas: 5
  maxReplicas: 50
  targetCPUUtilizationPercentage: 60
```

### Node Affinity for Compilation Workloads

Use dedicated nodes for compilation jobs:

```yaml
# values-production.yaml
nodeSelector:
  workload: compilation

tolerations:
  - key: "workload"
    operator: "Equal"
    value: "compilation"
    effect: "NoSchedule"
```

Then label nodes:
```bash
kubectl label nodes <node-name> workload=compilation
kubectl taint nodes <node-name> workload=compilation:NoSchedule
```

## Security Best Practices

1. **Use specific image tags** instead of `latest`
2. **Enable network policies** in production
3. **Use private container registries** with pull secrets
4. **Enable Pod Security Standards** at namespace level
5. **Regular security scans** of container images
6. **Rotate secrets** regularly
7. **Enable audit logging** for API access
8. **Implement rate limiting** at ingress level

## Additional Resources

- [Helm Documentation](https://helm.sh/docs/)
- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [Project README](../README.md)
- [Architecture Documentation](../KUBERNETES_ARCHITECTURE.md)
