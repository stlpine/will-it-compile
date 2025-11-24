# Deployment Guide

This guide covers deploying will-it-compile to both local development and production Kubernetes environments.

## Prerequisites

### All Environments
- Kubernetes cluster 1.24+
- Helm 3.0+
- `kubectl` configured to access your cluster
- Compiler Docker images built and available

### Redis Storage

The Helm chart includes **built-in Redis support** for shared job storage. This is enabled by default and provides:

- ✅ **Shared cache**: Multiple API instances can share job state
- ✅ **Horizontal scaling**: Deploy multiple replicas
- ✅ **Automatic cleanup**: TTL-based job expiration
- ⚠️ **Ephemeral storage**: Data lost on pod restart (by design)

**Configuration Options:**

1. **Embedded Redis** (Default): Chart deploys Redis Deployment
   - Ephemeral in-memory cache
   - Data lost on pod restart/shutdown
   - Good for horizontal scaling during uptime

2. **In-Memory Mode**: Disable Redis entirely
   - Set `redis.enabled=false` in values
   - Each API pod has separate in-memory storage
   - Only suitable for single-replica deployments

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
  --values ./deployments/helm/will-it-compile/values-prod.yaml
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
  --values ./deployments/helm/will-it-compile/values-prod.yaml \
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
  --values ./deployments/helm/will-it-compile/values-prod.yaml
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

## Accessing the Web Frontend

The Helm chart includes a React-based web frontend for an intuitive UI experience.

### Development Environment

In development, the web frontend is exposed via **NodePort** for easy local access:

**Method 1: Port Forwarding (Recommended)**
```bash
# Forward web frontend
kubectl port-forward svc/will-it-compile-web 3000:80

# Access in browser
open http://localhost:3000

# Forward API (if testing direct API calls)
kubectl port-forward svc/will-it-compile 8080:80
```

**Method 2: NodePort Direct Access**
```bash
# Get the NodePort
kubectl get svc will-it-compile-web -o jsonpath='{.spec.ports[0].nodePort}'

# With minikube
minikube service will-it-compile-web

# With kind - you'll need to access via localhost with port mapping
# (kind requires port mapping configured during cluster creation)
```

**Method 3: Enable Ingress (Optional)**

For a production-like local setup, you can enable ingress in development:

```bash
# Install nginx ingress controller (if not already installed)
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml

# Update values-dev.yaml to enable ingress
# Or install with ingress enabled
helm install will-it-compile ./deployments/helm/will-it-compile \
  --values ./deployments/helm/will-it-compile/values-dev.yaml \
  --set ingress.enabled=true \
  --set ingress.className=nginx \
  --set ingress.hosts[0].host=will-it-compile.local \
  --set ingress.hosts[0].paths[0].path=/api \
  --set ingress.hosts[0].paths[0].pathType=Prefix \
  --set ingress.hosts[0].paths[0].backend=api \
  --set ingress.hosts[0].paths[1].path=/ \
  --set ingress.hosts[0].paths[1].pathType=Prefix \
  --set ingress.hosts[0].paths[1].backend=web

# Add to /etc/hosts
echo "127.0.0.1 will-it-compile.local" | sudo tee -a /etc/hosts

# Access
open http://will-it-compile.local
```

### Production Environment

In production, the web frontend is accessed via **Ingress** with HTTPS:

**Default Configuration** (`values-production.yaml`):
```yaml
ingress:
  enabled: true
  className: nginx
  hosts:
    - host: compile.example.com
      paths:
        - path: /api          # Routes to API backend
          pathType: Prefix
          backend: api
        - path: /             # Routes to web frontend
          pathType: Prefix
          backend: web
  tls:
    - secretName: will-it-compile-tls
      hosts:
        - compile.example.com
```

**Setup Steps:**

1. **Install Ingress Controller** (if not already installed):
```bash
# NGINX Ingress Controller
helm install ingress-nginx ingress-nginx/ingress-nginx \
  --namespace ingress-nginx --create-namespace

# Verify installation
kubectl get pods -n ingress-nginx
```

2. **Install cert-manager** (for automatic TLS certificates):
```bash
# Install cert-manager
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml

# Create Let's Encrypt issuer (production)
cat <<EOF | kubectl apply -f -
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: your-email@example.com
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - http01:
        ingress:
          class: nginx
EOF
```

3. **Configure DNS**:
```bash
# Get ingress external IP
kubectl get ingress -n will-it-compile

# Point your domain to the ingress IP
# Example: compile.example.com -> <EXTERNAL-IP>
```

4. **Deploy with Custom Domain**:
```bash
helm install will-it-compile ./deployments/helm/will-it-compile \
  --namespace will-it-compile --create-namespace \
  --values ./deployments/helm/will-it-compile/values-prod.yaml \
  --set ingress.hosts[0].host=compile.yourdomain.com \
  --set ingress.tls[0].hosts[0]=compile.yourdomain.com
```

5. **Access**:
```bash
# Web UI (root path)
https://compile.yourdomain.com

# API endpoints
https://compile.yourdomain.com/api/v1/health
https://compile.yourdomain.com/api/v1/environments
```

### Verifying Web Frontend

**Health Check:**
```bash
# Web frontend health
curl http://localhost:3000/          # Dev (via port-forward)
curl https://compile.yourdomain.com/  # Production

# API health
curl http://localhost:8080/health    # Dev (via port-forward)
curl https://compile.yourdomain.com/api/v1/health  # Production
```

**Check Deployment Status:**
```bash
# Check web pods
kubectl get pods -l app.kubernetes.io/component=web

# Check web service
kubectl get svc -l app.kubernetes.io/component=web

# Check ingress routing
kubectl describe ingress will-it-compile

# View web logs
kubectl logs -l app.kubernetes.io/component=web --tail=50
```

### Troubleshooting Web Access

**Problem: Can't access web frontend**

1. Check if web pods are running:
```bash
kubectl get pods -l app.kubernetes.io/component=web
kubectl describe pod <web-pod-name>
```

2. Check service endpoints:
```bash
kubectl get endpoints will-it-compile-web
```

3. Check ingress configuration:
```bash
kubectl get ingress will-it-compile -o yaml
```

4. Test direct access to web pod:
```bash
kubectl port-forward <web-pod-name> 3000:80
curl http://localhost:3000
```

**Problem: API calls failing from web frontend**

1. Check API_SERVICE_URL environment variable in web pods:
```bash
kubectl get pods -l app.kubernetes.io/component=web -o jsonpath='{.items[0].spec.containers[0].env}'
```

2. Verify API service is reachable from web pods:
```bash
kubectl exec -it <web-pod-name> -- wget -O- http://will-it-compile:80/health
```

**Problem: Ingress not routing correctly**

1. Check ingress controller logs:
```bash
kubectl logs -n ingress-nginx -l app.kubernetes.io/component=controller
```

2. Verify backend services exist:
```bash
kubectl get svc will-it-compile will-it-compile-web
```

3. Test ingress rules:
```bash
# Should route to API
curl -H "Host: compile.yourdomain.com" http://<INGRESS-IP>/api/v1/health

# Should route to web
curl -H "Host: compile.yourdomain.com" http://<INGRESS-IP>/
```

## Redis Configuration

### Embedded Redis

The Helm chart deploys Redis as a **Deployment** (ephemeral cache):

**Key Characteristics:**
- ⚠️ **Data is NOT persistent** - Lost on pod restart/shutdown
- ✅ **Good for horizontal scaling** - Multiple API pods share cache during uptime
- ✅ **Simpler & faster** - No persistent volumes, faster pod restarts
- ✅ **Zero storage costs** - No PersistentVolumeClaims

**Development** (values-dev.yaml):
```yaml
redis:
  enabled: true
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
  --values ./deployments/helm/will-it-compile/values-prod.yaml
```

**Verify Redis Deployment:**
```bash
# Check Redis pod
kubectl get pods -n will-it-compile -l app.kubernetes.io/component=redis

# Check Redis logs
kubectl logs -n will-it-compile -l app.kubernetes.io/component=redis

# Test Redis connection (from API pod)
kubectl exec -n will-it-compile -it <api-pod-name> -- sh -c 'redis-cli -h will-it-compile-redis ping'

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
- Multiple replicas (jobs won't be shared across pods)

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
# Initializing Redis job store at will-it-compile-redis:6379
# Redis job store initialized successfully (TTL: 24h0m0s)
```

**Important Notes:**
- Redis uses ephemeral storage (Deployment, not StatefulSet)
- Data is lost when Redis pod restarts or is deleted
- This is by design for simplicity and lower resource usage
- Jobs are shared across API pods during uptime only

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
  --values ./deployments/helm/will-it-compile/values-prod.yaml \
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
  --values ./deployments/helm/will-it-compile/values-prod.yaml \
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
  --values ./deployments/helm/will-it-compile/values-prod.yaml \
  --set image.repository=yourregistry.azurecr.io/will-it-compile/api \
  --set compilerImages.cpp.repository=yourregistry.azurecr.io/will-it-compile/cpp-gcc
```

## Upgrading

```bash
# Update with new values
helm upgrade will-it-compile ./deployments/helm/will-it-compile \
  --namespace will-it-compile \
  --values ./deployments/helm/will-it-compile/values-prod.yaml

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
# values-prod.yaml
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
# values-prod.yaml
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
