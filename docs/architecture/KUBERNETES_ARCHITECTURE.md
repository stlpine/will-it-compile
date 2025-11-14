# Kubernetes Architecture for will-it-compile

## Problem Statement

The current Docker-based architecture does not work in Kubernetes environments because:
1. No Docker daemon runs inside pods
2. Only container runtimes (containerd, CRI-O) exist at the node level
3. Mounting host Docker socket breaks security boundaries

## Solution: Kubernetes Jobs API

### Architecture Overview

```
┌─────────────┐
│  API Pod    │
│             │
│  Receives   │──┐
│  Request    │  │
└─────────────┘  │
                 │ Creates K8s Job
                 ↓
         ┌──────────────┐
         │ Kubernetes   │
         │ API Server   │
         └──────────────┘
                 │
                 │ Schedules
                 ↓
         ┌──────────────┐
         │ Compiler Pod │ (Ephemeral)
         │              │
         │ - Runs once  │
         │ - Compiles   │
         │ - Exits      │
         └──────────────┘
                 │
                 │ Stores result
                 ↓
         ┌──────────────┐
         │  Redis/DB    │
         └──────────────┘
```

## Implementation Strategy

### Phase 1: Create Kubernetes Client Interface

```go
// pkg/runtime/interface.go
package runtime

import "context"

// CompilationRuntime abstracts away the execution environment
type CompilationRuntime interface {
    // Compile runs code compilation in an isolated environment
    Compile(ctx context.Context, config CompilationConfig) (*CompilationOutput, error)

    // Cleanup removes any temporary resources
    Cleanup(ctx context.Context, jobID string) error
}

// CompilationConfig holds configuration for compilation
type CompilationConfig struct {
    JobID      string
    ImageTag   string
    SourceCode string
    Env        []string
    Timeout    time.Duration
}

// CompilationOutput holds the result
type CompilationOutput struct {
    Stdout   string
    Stderr   string
    ExitCode int
    Duration time.Duration
}
```

### Phase 2: Docker Implementation (Local Development)

```go
// internal/runtime/docker/runtime.go
package docker

import (
    "context"
    "github.com/stlpine/will-it-compile/pkg/runtime"
)

type DockerRuntime struct {
    client *docker.Client
}

func NewDockerRuntime() (*DockerRuntime, error) {
    // Existing Docker implementation
}

func (d *DockerRuntime) Compile(ctx context.Context, config runtime.CompilationConfig) (*runtime.CompilationOutput, error) {
    // Use existing Docker code
}
```

### Phase 3: Kubernetes Implementation (Production)

```go
// internal/runtime/kubernetes/runtime.go
package kubernetes

import (
    "context"
    "fmt"

    batchv1 "k8s.io/api/batch/v1"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/rest"
)

type KubernetesRuntime struct {
    clientset *kubernetes.Clientset
    namespace string
}

func NewKubernetesRuntime(namespace string) (*KubernetesRuntime, error) {
    // Create in-cluster config
    config, err := rest.InClusterConfig()
    if err != nil {
        return nil, fmt.Errorf("failed to get in-cluster config: %w", err)
    }

    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        return nil, fmt.Errorf("failed to create clientset: %w", err)
    }

    return &KubernetesRuntime{
        clientset: clientset,
        namespace: namespace,
    }, nil
}

func (k *KubernetesRuntime) Compile(ctx context.Context, config runtime.CompilationConfig) (*runtime.CompilationOutput, error) {
    // 1. Create ConfigMap with source code
    configMap := &corev1.ConfigMap{
        ObjectMeta: metav1.ObjectMeta{
            Name: fmt.Sprintf("source-%s", config.JobID),
            Labels: map[string]string{
                "app":    "will-it-compile",
                "job-id": config.JobID,
            },
        },
        Data: map[string]string{
            "source.cpp": config.SourceCode,
        },
    }

    _, err := k.clientset.CoreV1().ConfigMaps(k.namespace).Create(ctx, configMap, metav1.CreateOptions{})
    if err != nil {
        return nil, fmt.Errorf("failed to create configmap: %w", err)
    }

    // 2. Create Job
    job := k.createCompilationJob(config)

    createdJob, err := k.clientset.BatchV1().Jobs(k.namespace).Create(ctx, job, metav1.CreateOptions{})
    if err != nil {
        return nil, fmt.Errorf("failed to create job: %w", err)
    }

    // 3. Wait for job completion
    output, err := k.waitForJobCompletion(ctx, createdJob.Name, config.Timeout)
    if err != nil {
        return nil, err
    }

    // 4. Cleanup (async in production)
    defer k.Cleanup(context.Background(), config.JobID)

    return output, nil
}

func (k *KubernetesRuntime) createCompilationJob(config runtime.CompilationConfig) *batchv1.Job {
    backoffLimit := int32(0)
    ttlSeconds := int32(300) // Clean up after 5 minutes

    return &batchv1.Job{
        ObjectMeta: metav1.ObjectMeta{
            Name: fmt.Sprintf("compile-%s", config.JobID),
            Labels: map[string]string{
                "app":    "will-it-compile",
                "job-id": config.JobID,
            },
        },
        Spec: batchv1.JobSpec{
            TTLSecondsAfterFinished: &ttlSeconds,
            BackoffLimit:            &backoffLimit,
            Template: corev1.PodTemplateSpec{
                Spec: corev1.PodSpec{
                    RestartPolicy: corev1.RestartPolicyNever,
                    SecurityContext: &corev1.PodSecurityContext{
                        RunAsNonRoot: ptr(true),
                        RunAsUser:    ptr(int64(1000)),
                        FSGroup:      ptr(int64(1000)),
                        SeccompProfile: &corev1.SeccompProfile{
                            Type: corev1.SeccompProfileTypeRuntimeDefault,
                        },
                    },
                    Containers: []corev1.Container{
                        {
                            Name:  "compiler",
                            Image: config.ImageTag,
                            Command: []string{"/usr/bin/compile.sh"},
                            Env: convertEnv(config.Env),
                            Resources: corev1.ResourceRequirements{
                                Limits: corev1.ResourceList{
                                    corev1.ResourceCPU:    resource.MustParse("500m"),
                                    corev1.ResourceMemory: resource.MustParse("128Mi"),
                                },
                                Requests: corev1.ResourceList{
                                    corev1.ResourceCPU:    resource.MustParse("100m"),
                                    corev1.ResourceMemory: resource.MustParse("64Mi"),
                                },
                            },
                            SecurityContext: &corev1.SecurityContext{
                                AllowPrivilegeEscalation: ptr(false),
                                RunAsNonRoot:             ptr(true),
                                RunAsUser:                ptr(int64(1000)),
                                Capabilities: &corev1.Capabilities{
                                    Drop: []corev1.Capability{"ALL"},
                                },
                                ReadOnlyRootFilesystem: ptr(false),
                            },
                            VolumeMounts: []corev1.VolumeMount{
                                {
                                    Name:      "source",
                                    MountPath: "/workspace",
                                    ReadOnly:  true,
                                },
                                {
                                    Name:      "tmp",
                                    MountPath: "/tmp",
                                },
                            },
                        },
                    },
                    Volumes: []corev1.Volume{
                        {
                            Name: "source",
                            VolumeSource: corev1.VolumeSource{
                                ConfigMap: &corev1.ConfigMapVolumeSource{
                                    LocalObjectReference: corev1.LocalObjectReference{
                                        Name: fmt.Sprintf("source-%s", config.JobID),
                                    },
                                },
                            },
                        },
                        {
                            Name: "tmp",
                            VolumeSource: corev1.VolumeSource{
                                EmptyDir: &corev1.EmptyDirVolumeSource{
                                    Medium:    corev1.StorageMediumMemory,
                                    SizeLimit: resource.NewQuantity(64*1024*1024, resource.BinarySI),
                                },
                            },
                        },
                    },
                },
            },
        },
    }
}

func (k *KubernetesRuntime) waitForJobCompletion(ctx context.Context, jobName string, timeout time.Duration) (*runtime.CompilationOutput, error) {
    // Watch for job completion
    watcher, err := k.clientset.BatchV1().Jobs(k.namespace).Watch(ctx, metav1.ListOptions{
        FieldSelector: fmt.Sprintf("metadata.name=%s", jobName),
        TimeoutSeconds: ptr(int64(timeout.Seconds())),
    })
    if err != nil {
        return nil, err
    }
    defer watcher.Stop()

    for event := range watcher.ResultChan() {
        job, ok := event.Object.(*batchv1.Job)
        if !ok {
            continue
        }

        if job.Status.Succeeded > 0 {
            // Job completed successfully, get logs
            return k.getJobOutput(ctx, jobName)
        }

        if job.Status.Failed > 0 {
            return &runtime.CompilationOutput{
                ExitCode: 1,
                Stderr:   "Job failed",
            }, nil
        }
    }

    return nil, fmt.Errorf("job timeout")
}

func (k *KubernetesRuntime) getJobOutput(ctx context.Context, jobName string) (*runtime.CompilationOutput, error) {
    // Get pods created by the job
    pods, err := k.clientset.CoreV1().Pods(k.namespace).List(ctx, metav1.ListOptions{
        LabelSelector: fmt.Sprintf("job-name=%s", jobName),
    })
    if err != nil || len(pods.Items) == 0 {
        return nil, fmt.Errorf("failed to get job pods: %w", err)
    }

    pod := pods.Items[0]

    // Get logs
    logOptions := &corev1.PodLogOptions{
        Container: "compiler",
    }

    req := k.clientset.CoreV1().Pods(k.namespace).GetLogs(pod.Name, logOptions)
    logs, err := req.Stream(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to get logs: %w", err)
    }
    defer logs.Close()

    // Read logs with size limit
    buf := make([]byte, 1*1024*1024) // 1MB limit
    n, _ := logs.Read(buf)
    output := string(buf[:n])

    // Parse exit code from pod status
    exitCode := 0
    if pod.Status.ContainerStatuses != nil && len(pod.Status.ContainerStatuses) > 0 {
        if terminated := pod.Status.ContainerStatuses[0].State.Terminated; terminated != nil {
            exitCode = int(terminated.ExitCode)
        }
    }

    return &runtime.CompilationOutput{
        Stdout:   output,
        ExitCode: exitCode,
    }, nil
}

func (k *KubernetesRuntime) Cleanup(ctx context.Context, jobID string) error {
    deletePolicy := metav1.DeletePropagationForeground

    // Delete job (will delete associated pods)
    jobName := fmt.Sprintf("compile-%s", jobID)
    err := k.clientset.BatchV1().Jobs(k.namespace).Delete(ctx, jobName, metav1.DeleteOptions{
        PropagationPolicy: &deletePolicy,
    })
    if err != nil {
        return err
    }

    // Delete configmap
    configMapName := fmt.Sprintf("source-%s", jobID)
    return k.clientset.CoreV1().ConfigMaps(k.namespace).Delete(ctx, configMapName, metav1.DeleteOptions{})
}

func ptr[T any](v T) *T {
    return &v
}

func convertEnv(envVars []string) []corev1.EnvVar {
    result := make([]corev1.EnvVar, len(envVars))
    for i, e := range envVars {
        parts := strings.SplitN(e, "=", 2)
        result[i] = corev1.EnvVar{
            Name:  parts[0],
            Value: parts[1],
        }
    }
    return result
}
```

### Phase 4: Runtime Selection

```go
// cmd/api/main.go
func main() {
    var compilationRuntime runtime.CompilationRuntime
    var err error

    // Detect environment
    if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
        // Running in Kubernetes
        namespace := os.Getenv("NAMESPACE")
        if namespace == "" {
            namespace = "default"
        }
        compilationRuntime, err = kubernetes.NewKubernetesRuntime(namespace)
    } else {
        // Running locally with Docker
        compilationRuntime, err = docker.NewDockerRuntime()
    }

    if err != nil {
        log.Fatalf("Failed to create runtime: %v", err)
    }

    // Pass runtime to compiler
    compiler := compiler.NewCompilerWithRuntime(compilationRuntime)

    // ... rest of setup
}
```

## Deployment Configuration

### Kubernetes Manifests

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: will-it-compile-api
spec:
  replicas: 3
  selector:
    matchLabels:
      app: will-it-compile-api
  template:
    metadata:
      labels:
        app: will-it-compile-api
    spec:
      serviceAccountName: will-it-compile
      containers:
      - name: api
        image: will-it-compile/api:latest
        env:
        - name: NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        - name: RUNTIME
          value: "kubernetes"
        ports:
        - containerPort: 8080
        resources:
          limits:
            cpu: "1"
            memory: "512Mi"
          requests:
            cpu: "100m"
            memory: "128Mi"
        securityContext:
          runAsNonRoot: true
          runAsUser: 1000
          allowPrivilegeEscalation: false
          capabilities:
            drop: ["ALL"]
          readOnlyRootFilesystem: true
---
# rbac.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: will-it-compile
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: will-it-compile-job-manager
rules:
- apiGroups: ["batch"]
  resources: ["jobs"]
  verbs: ["create", "get", "list", "watch", "delete"]
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["get", "list", "watch"]
- apiGroups: [""]
  resources: ["pods/log"]
  verbs: ["get"]
- apiGroups: [""]
  resources: ["configmaps"]
  verbs: ["create", "delete"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: will-it-compile-job-manager
subjects:
- kind: ServiceAccount
  name: will-it-compile
roleRef:
  kind: Role
  name: will-it-compile-job-manager
  apiGroup: rbac.authorization.k8s.io
```

## Security Considerations

### Pod Security Standards

1. **Restricted Profile**: Use Kubernetes restricted pod security profile
2. **Resource Limits**: CPU/Memory limits prevent resource exhaustion
3. **Network Policies**: Isolate compiler pods (no internet access)
4. **Ephemeral Storage**: Use emptyDir for temporary files
5. **RBAC**: Minimal permissions for API service account

### Image Security

```yaml
# NetworkPolicy to block internet access for compiler pods
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: compiler-pod-isolation
spec:
  podSelector:
    matchLabels:
      app: will-it-compile
      component: compiler
  policyTypes:
  - Egress
  egress:
  - to:
    - namespaceSelector:
        matchLabels:
          name: kube-system
    ports:
    - protocol: TCP
      port: 53  # DNS only
```

## Benefits of Kubernetes Approach

1. ✅ **Native Integration**: Uses K8s APIs properly
2. ✅ **Security**: No Docker socket mounting needed
3. ✅ **Isolation**: Each compilation in separate pod
4. ✅ **Resource Management**: K8s handles scheduling and limits
5. ✅ **Cleanup**: TTLSecondsAfterFinished handles cleanup
6. ✅ **Scalability**: K8s handles pod distribution
7. ✅ **Monitoring**: Standard K8s monitoring tools work

## Alternative: Pre-Created Worker Pools

For higher throughput, consider worker pool pattern:

```
API Pod → Job Queue (Redis) → Worker Pods (pre-created)
```

Workers:
- Run continuously
- Poll queue for work
- Execute in isolated processes/containers within pod
- Use gVisor or similar for sandboxing

## Migration Path

### Phase 1 (Current - MVP)
- ✅ Docker-based for local development
- ✅ Works on single machine

### Phase 2 (Kubernetes Support)
- Create runtime interface
- Implement Kubernetes runtime
- Auto-detect environment
- Deploy to K8s cluster

### Phase 3 (Optimization)
- Add job queue (Redis)
- Worker pool pattern
- Pre-warm pods
- Advanced scheduling

### Phase 4 (Advanced)
- gVisor/Firecracker for isolation
- WebAssembly runtime
- Multi-cloud support

## Dependencies

Add to `go.mod`:
```go
require (
    k8s.io/api v0.28.0
    k8s.io/apimachinery v0.28.0
    k8s.io/client-go v0.28.0
)
```

## Testing Strategy

### Local Development
- Continue using Docker
- No changes needed

### Kubernetes Testing
- Use kind (Kubernetes in Docker) for integration tests
- Mock Kubernetes client for unit tests

```bash
# Create kind cluster for testing
kind create cluster --name will-it-compile-test

# Run tests
RUNTIME=kubernetes go test ./...
```

## Documentation Updates Needed

1. Update CLAUDE.md with K8s architecture
2. Add deployment guide for K8s
3. Document RBAC requirements
4. Add troubleshooting for K8s issues

## References

- [Kubernetes Jobs](https://kubernetes.io/docs/concepts/workloads/controllers/job/)
- [Pod Security Standards](https://kubernetes.io/docs/concepts/security/pod-security-standards/)
- [gVisor](https://gvisor.dev/) - Application kernel for containers
- [Firecracker](https://firecracker-microvm.github.io/) - Lightweight VMs
