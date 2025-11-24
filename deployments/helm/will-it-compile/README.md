# will-it-compile Helm Chart

This Helm chart deploys the will-it-compile service to a Kubernetes cluster.

## Prerequisites

- Kubernetes 1.24+
- Helm 3.0+
- Compiler images built and available in the cluster

## Installing the Chart

### Quick Start

```bash
# Install with default values
helm install will-it-compile ./deployments/helm/will-it-compile

# Install with custom values
helm install will-it-compile ./deployments/helm/will-it-compile \
  --values ./deployments/helm/will-it-compile/values-prod.yaml

# Install in a specific namespace
helm install will-it-compile ./deployments/helm/will-it-compile \
  --namespace will-it-compile --create-namespace
```

## Uninstalling the Chart

```bash
helm uninstall will-it-compile
```

## Configuration

The following table lists the configurable parameters and their default values.

### Application Settings

| Parameter | Description | Default |
|-----------|-------------|---------|
| `replicaCount` | Number of replicas | `2` |
| `image.repository` | API server image repository | `will-it-compile/api` |
| `image.tag` | API server image tag | `latest` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |

### Compiler Images

| Parameter | Description | Default |
|-----------|-------------|---------|
| `compilerImages.cpp.repository` | C++ compiler image repository | `will-it-compile/cpp-gcc` |
| `compilerImages.cpp.tag` | C++ compiler image tag | `13-alpine` |

### Service Account & RBAC

| Parameter | Description | Default |
|-----------|-------------|---------|
| `serviceAccount.create` | Create service account | `true` |
| `serviceAccount.name` | Service account name | `""` (generated) |
| `serviceAccount.annotations` | Service account annotations | `{}` |

### Service

| Parameter | Description | Default |
|-----------|-------------|---------|
| `service.type` | Service type | `ClusterIP` |
| `service.port` | Service port | `80` |
| `service.targetPort` | Container port | `8080` |

### Ingress

| Parameter | Description | Default |
|-----------|-------------|---------|
| `ingress.enabled` | Enable ingress | `false` |
| `ingress.className` | Ingress class name | `""` |
| `ingress.annotations` | Ingress annotations | `{}` |
| `ingress.hosts` | Ingress hosts configuration | See values.yaml |
| `ingress.tls` | Ingress TLS configuration | `[]` |

### Resources

| Parameter | Description | Default |
|-----------|-------------|---------|
| `resources.limits.cpu` | CPU limit | `1000m` |
| `resources.limits.memory` | Memory limit | `512Mi` |
| `resources.requests.cpu` | CPU request | `200m` |
| `resources.requests.memory` | Memory request | `256Mi` |

### Autoscaling

| Parameter | Description | Default |
|-----------|-------------|---------|
| `autoscaling.enabled` | Enable HPA | `false` |
| `autoscaling.minReplicas` | Minimum replicas | `2` |
| `autoscaling.maxReplicas` | Maximum replicas | `10` |
| `autoscaling.targetCPUUtilizationPercentage` | Target CPU % | `70` |
| `autoscaling.targetMemoryUtilizationPercentage` | Target Memory % | `80` |

### Network Policy

| Parameter | Description | Default |
|-----------|-------------|---------|
| `networkPolicy.enabled` | Enable network policy | `true` |
| `networkPolicy.policyTypes` | Policy types | `["Ingress", "Egress"]` |

### Pod Disruption Budget

| Parameter | Description | Default |
|-----------|-------------|---------|
| `podDisruptionBudget.enabled` | Enable PDB | `true` |
| `podDisruptionBudget.minAvailable` | Minimum available pods | `1` |

### Security

| Parameter | Description | Default |
|-----------|-------------|---------|
| `podSecurityContext.runAsNonRoot` | Run as non-root | `true` |
| `podSecurityContext.runAsUser` | User ID | `1000` |
| `podSecurityContext.fsGroup` | FS group | `1000` |
| `securityContext.allowPrivilegeEscalation` | Allow privilege escalation | `false` |
| `securityContext.capabilities.drop` | Dropped capabilities | `["ALL"]` |

## Examples

### Production Deployment with Ingress

```bash
helm install will-it-compile ./deployments/helm/will-it-compile \
  --set ingress.enabled=true \
  --set ingress.hosts[0].host=compile.example.com \
  --set ingress.hosts[0].paths[0].path=/ \
  --set ingress.hosts[0].paths[0].pathType=Prefix \
  --set autoscaling.enabled=true
```

### Development Deployment

```bash
helm install will-it-compile ./deployments/helm/will-it-compile \
  --set replicaCount=1 \
  --set resources.requests.cpu=100m \
  --set resources.requests.memory=128Mi \
  --set networkPolicy.enabled=false
```

### Using Custom Image Registry

```bash
helm install will-it-compile ./deployments/helm/will-it-compile \
  --set image.repository=myregistry.io/will-it-compile/api \
  --set compilerImages.cpp.repository=myregistry.io/will-it-compile/cpp-gcc
```

## Preparing Compiler Images

Before deploying, note that compiler images are pulled from official Docker Hub repositories:

```bash
# Official compiler images (no build required):
# - gcc:13 (C/C++)
# - golang:1.22-alpine (Go)
# - rust:1.75-alpine (Rust)

# If using private registry for API server, use image pull secrets
kubectl create secret docker-registry regcred \
  --docker-server=myregistry.io \
  --docker-username=user \
  --docker-password=pass

helm install will-it-compile ./deployments/helm/will-it-compile \
  --set imagePullSecrets[0].name=regcred
```

## Verifying the Deployment

```bash
# Check pods
kubectl get pods -l app=will-it-compile

# Check service
kubectl get svc will-it-compile

# Test health endpoint
kubectl port-forward svc/will-it-compile 8080:80
curl http://localhost:8080/health

# Test compilation
curl -X POST http://localhost:8080/api/v1/compile \
  -H "Content-Type: application/json" \
  -d '{
    "code": "aW50IG1haW4oKSB7IHJldHVybiAwOyB9",
    "language": "cpp",
    "compiler": "gcc-13"
  }'
```

## Upgrading

```bash
# Upgrade with new values
helm upgrade will-it-compile ./deployments/helm/will-it-compile \
  --values ./values-prod.yaml

# Upgrade with inline changes
helm upgrade will-it-compile ./deployments/helm/will-it-compile \
  --set image.tag=v1.1.0
```

## Troubleshooting

### Pods not starting

```bash
# Check pod status
kubectl describe pod -l app=will-it-compile

# Check logs
kubectl logs -l app=will-it-compile

# Common issues:
# 1. Missing compiler images - ensure they're pre-pulled or in registry
# 2. RBAC permissions - check Role and RoleBinding are created
# 3. Resource limits - check if nodes have enough capacity
```

### Compilation jobs failing

```bash
# Check compilation job pods
kubectl get pods -l managed-by=will-it-compile

# View job logs
kubectl logs -l job-id=<job-id>

# Common issues:
# 1. RBAC permissions - check Job creation permissions
# 2. Image pull failures - verify compiler images exist
# 3. Resource quotas - check namespace resource limits
```

### Network issues

```bash
# Check network policy
kubectl get networkpolicy

# Temporarily disable network policy for debugging
helm upgrade will-it-compile ./deployments/helm/will-it-compile \
  --set networkPolicy.enabled=false
```

## Security Considerations

This chart implements several security best practices:

1. **Non-root execution**: Containers run as user 1000
2. **No privilege escalation**: Prevented at pod and container level
3. **Dropped capabilities**: All Linux capabilities dropped
4. **Seccomp profile**: Runtime default seccomp profile applied
5. **Network policy**: Restricts ingress/egress traffic
6. **RBAC**: Minimal permissions (only what's needed for compilation)
7. **Pod Disruption Budget**: Ensures high availability
8. **Read-only root filesystem**: Can be enabled (currently false for /tmp)

## License

See repository LICENSE file.
