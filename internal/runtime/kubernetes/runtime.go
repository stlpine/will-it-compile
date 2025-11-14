package kubernetes

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/stlpine/will-it-compile/pkg/runtime"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Sentinel errors for Kubernetes runtime.
var (
	ErrWatchChannelClosed = errors.New("watch channel closed unexpectedly")
)

const (
	// Resource limits for compilation pods.
	MaxMemory = "128Mi"
	MaxCPU    = "500m"
	ReqMemory = "64Mi"
	ReqCPU    = "100m"

	// Max output size (1MB).
	MaxOutputSize = 1 * 1024 * 1024

	// TTL for completed jobs (5 minutes).
	JobTTLSeconds = 300
)

// KubernetesRuntime implements CompilationRuntime using Kubernetes Jobs
// This is used for production deployments in Kubernetes clusters.
type KubernetesRuntime struct {
	clientset *kubernetes.Clientset
	namespace string
}

// NewKubernetesRuntime creates a new Kubernetes-based compilation runtime.
func NewKubernetesRuntime(namespace string) (*KubernetesRuntime, error) {
	// Use in-cluster config (when running inside K8s)
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get in-cluster config: %w (are you running inside Kubernetes?)", err)
	}

	// Create Kubernetes clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes clientset: %w", err)
	}

	if namespace == "" {
		namespace = "default"
	}

	return &KubernetesRuntime{
		clientset: clientset,
		namespace: namespace,
	}, nil
}

// Compile runs compilation using Kubernetes Jobs.
func (k *KubernetesRuntime) Compile(ctx context.Context, config runtime.CompilationConfig) (*runtime.CompilationOutput, error) {
	startTime := time.Now()

	// 1. Create ConfigMap with source code
	if err := k.createSourceConfigMap(ctx, config); err != nil {
		return nil, fmt.Errorf("failed to create source configmap: %w", err)
	}

	// 2. Create and run Job
	job, err := k.createCompilationJob(ctx, config)
	if err != nil {
		// Use context without cancel to allow cleanup even if parent context is cancelled
		cleanupCtx := context.WithoutCancel(ctx)
		k.cleanup(cleanupCtx, config.JobID) // Clean up configmap on error
		return nil, fmt.Errorf("failed to create job: %w", err)
	}

	// 3. Wait for job completion with timeout
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second // Default timeout
	}

	output, timedOut, err := k.waitForJobCompletion(ctx, job.Name, timeout)
	if err != nil {
		cleanupCtx := context.WithoutCancel(ctx)
		k.cleanup(cleanupCtx, config.JobID)
		return nil, fmt.Errorf("failed waiting for job: %w", err)
	}

	// 4. Schedule cleanup (async to not block response)
	// Use context without cancel to allow cleanup to complete even if parent is cancelled
	cleanupCtx := context.WithoutCancel(ctx)
	go k.cleanup(cleanupCtx, config.JobID)

	output.Duration = time.Since(startTime)
	output.TimedOut = timedOut

	return output, nil
}

// ImageExists checks if a container image exists
// Note: In K8s, we rely on image pull policy and let K8s handle image verification.
func (k *KubernetesRuntime) ImageExists(ctx context.Context, imageTag string) (bool, error) {
	// In Kubernetes, we can't easily check if an image exists without pulling it
	// Instead, we'll return true and let the Job creation fail if image doesn't exist
	// This is acceptable because:
	// 1. Images should be pre-pulled or available in registry
	// 2. Image pull failures are caught during job execution
	// 3. We can add image validation in deployment checks
	return true, nil
}

// Close cleans up any resources.
func (k *KubernetesRuntime) Close() error {
	// Kubernetes clientset doesn't need explicit cleanup
	return nil
}

// createSourceConfigMap creates a ConfigMap containing the source code.
func (k *KubernetesRuntime) createSourceConfigMap(ctx context.Context, config runtime.CompilationConfig) error {
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "source-" + config.JobID,
			Namespace: k.namespace,
			Labels: map[string]string{
				"app":        "will-it-compile",
				"component":  "source",
				"job-id":     config.JobID,
				"managed-by": "will-it-compile",
			},
		},
		Data: map[string]string{
			"source.cpp": config.SourceCode,
		},
	}

	_, err := k.clientset.CoreV1().ConfigMaps(k.namespace).Create(ctx, configMap, metav1.CreateOptions{})
	return err
}

// createCompilationJob creates a Kubernetes Job for compilation.
func (k *KubernetesRuntime) createCompilationJob(ctx context.Context, config runtime.CompilationConfig) (*batchv1.Job, error) {
	backoffLimit := int32(0) // Don't retry failed jobs
	ttlSeconds := int32(JobTTLSeconds)

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "compile-" + config.JobID,
			Namespace: k.namespace,
			Labels: map[string]string{
				"app":        "will-it-compile",
				"component":  "compiler",
				"job-id":     config.JobID,
				"managed-by": "will-it-compile",
			},
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: &ttlSeconds,
			BackoffLimit:            &backoffLimit,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":        "will-it-compile",
						"component":  "compiler",
						"job-id":     config.JobID,
						"managed-by": "will-it-compile",
					},
				},
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
							Name:    "compiler",
							Image:   config.ImageTag,
							Command: []string{"/usr/bin/compile.sh"},
							Env:     k.convertEnv(config.Env),
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse(MaxCPU),
									corev1.ResourceMemory: resource.MustParse(MaxMemory),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse(ReqCPU),
									corev1.ResourceMemory: resource.MustParse(ReqMemory),
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
										Name: "source-" + config.JobID,
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

	return k.clientset.BatchV1().Jobs(k.namespace).Create(ctx, job, metav1.CreateOptions{})
}

// waitForJobCompletion waits for a Job to complete and returns its output.
func (k *KubernetesRuntime) waitForJobCompletion(ctx context.Context, jobName string, timeout time.Duration) (*runtime.CompilationOutput, bool, error) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Watch for job completion
	watcher, err := k.clientset.BatchV1().Jobs(k.namespace).Watch(ctx, metav1.ListOptions{
		FieldSelector:  "metadata.name=" + jobName,
		TimeoutSeconds: ptr(int64(timeout.Seconds())),
	})
	if err != nil {
		return nil, false, fmt.Errorf("failed to watch job: %w", err)
	}
	defer watcher.Stop()

	// Wait for completion or timeout
	for {
		select {
		case event, ok := <-watcher.ResultChan():
			if !ok {
				// Channel closed, check if it was due to timeout
				if errors.Is(ctx.Err(), context.DeadlineExceeded) {
					return &runtime.CompilationOutput{
						Stderr:   "Compilation timeout",
						ExitCode: 137, // SIGKILL
					}, true, nil
				}
				return nil, false, ErrWatchChannelClosed
			}

			job, ok := event.Object.(*batchv1.Job)
			if !ok {
				continue
			}

			// Check if job succeeded
			if job.Status.Succeeded > 0 {
				// Use context without cancel to allow output collection even if parent context is cancelled
				outputCtx := context.WithoutCancel(ctx)
				output, err := k.getJobOutput(outputCtx, jobName)
				return output, false, err
			}

			// Check if job failed
			if job.Status.Failed > 0 {
				// Use context without cancel to allow output collection even if parent context is cancelled
				outputCtx := context.WithoutCancel(ctx)
				output, _ := k.getJobOutput(outputCtx, jobName) //nolint:errcheck // best effort output collection
				if output == nil {
					output = &runtime.CompilationOutput{
						Stderr:   "Job failed to execute",
						ExitCode: 1,
					}
				}
				return output, false, nil
			}

		case <-ctx.Done():
			// Timeout occurred
			return &runtime.CompilationOutput{
				Stderr:   "Compilation timeout",
				ExitCode: 137,
			}, true, nil
		}
	}
}

// getJobOutput retrieves the output from a completed job.
func (k *KubernetesRuntime) getJobOutput(ctx context.Context, jobName string) (*runtime.CompilationOutput, error) {
	// Get pods created by the job
	pods, err := k.clientset.CoreV1().Pods(k.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "job-name=" + jobName,
	})
	if err != nil || len(pods.Items) == 0 {
		return nil, fmt.Errorf("failed to get job pods: %w", err)
	}

	pod := pods.Items[0]

	// Get container exit code
	exitCode := 0
	if len(pod.Status.ContainerStatuses) > 0 {
		if terminated := pod.Status.ContainerStatuses[0].State.Terminated; terminated != nil {
			exitCode = int(terminated.ExitCode)
		}
	}

	// Get logs
	logOptions := &corev1.PodLogOptions{
		Container: "compiler",
	}

	req := k.clientset.CoreV1().Pods(k.namespace).GetLogs(pod.Name, logOptions)
	logStream, err := req.Stream(ctx)
	if err != nil {
		return &runtime.CompilationOutput{
			Stderr:   fmt.Sprintf("Failed to get logs: %v", err),
			ExitCode: exitCode,
		}, nil
	}
	defer logStream.Close() //nolint:errcheck // read-only operation

	// Read logs with size limit
	buf := make([]byte, MaxOutputSize)
	n, _ := io.ReadFull(logStream, buf) //nolint:errcheck // best effort read
	if n == 0 {
		// Try reading whatever is available
		n, _ = logStream.Read(buf) //nolint:errcheck // best effort read
	}

	output := string(buf[:n])

	// Split stdout/stderr (if needed, for now we treat all as stdout)
	return &runtime.CompilationOutput{
		Stdout:   output,
		Stderr:   "",
		ExitCode: exitCode,
	}, nil
}

// cleanup removes the ConfigMap and Job resources.
func (k *KubernetesRuntime) cleanup(ctx context.Context, jobID string) {
	deletePolicy := metav1.DeletePropagationForeground

	// Delete job (pods will be deleted automatically due to TTL)
	jobName := "compile-" + jobID
	_ = k.clientset.BatchV1().Jobs(k.namespace).Delete(ctx, jobName, metav1.DeleteOptions{ //nolint:errcheck // best effort cleanup
		PropagationPolicy: &deletePolicy,
	})

	// Delete configmap
	configMapName := "source-" + jobID
	_ = k.clientset.CoreV1().ConfigMaps(k.namespace).Delete(ctx, configMapName, metav1.DeleteOptions{}) //nolint:errcheck // best effort cleanup
}

// convertEnv converts []string environment variables to []corev1.EnvVar.
func (k *KubernetesRuntime) convertEnv(envVars []string) []corev1.EnvVar {
	result := make([]corev1.EnvVar, 0, len(envVars))
	for _, e := range envVars {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			result = append(result, corev1.EnvVar{
				Name:  parts[0],
				Value: parts[1],
			})
		}
	}
	return result
}

// ptr is a helper to get pointer to a value.
func ptr[T any](v T) *T {
	return &v
}

// Ensure KubernetesRuntime implements CompilationRuntime.
var _ runtime.CompilationRuntime = (*KubernetesRuntime)(nil)
