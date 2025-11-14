package runtime

import (
	"fmt"
	"os"

	pkgruntime "github.com/stlpine/will-it-compile/pkg/runtime"
	dockerruntime "github.com/stlpine/will-it-compile/internal/runtime/docker"
	k8sruntime "github.com/stlpine/will-it-compile/internal/runtime/kubernetes"
)

// RuntimeType represents the type of runtime to use
type RuntimeType string

const (
	RuntimeTypeDocker     RuntimeType = "docker"
	RuntimeTypeKubernetes RuntimeType = "kubernetes"
	RuntimeTypeAuto       RuntimeType = "auto"
)

// NewRuntime creates a new CompilationRuntime based on the specified type
// If runtimeType is "auto", it will auto-detect the environment
func NewRuntime(runtimeType RuntimeType, namespace string) (pkgruntime.CompilationRuntime, error) {
	switch runtimeType {
	case RuntimeTypeDocker:
		return dockerruntime.NewDockerRuntime()

	case RuntimeTypeKubernetes:
		if namespace == "" {
			namespace = os.Getenv("NAMESPACE")
			if namespace == "" {
				namespace = "default"
			}
		}
		return k8sruntime.NewKubernetesRuntime(namespace)

	case RuntimeTypeAuto:
		return NewRuntimeAuto(namespace)

	default:
		return nil, fmt.Errorf("unknown runtime type: %s", runtimeType)
	}
}

// NewRuntimeAuto automatically detects the environment and creates the appropriate runtime
func NewRuntimeAuto(namespace string) (pkgruntime.CompilationRuntime, error) {
	// Check if running inside Kubernetes
	// The presence of KUBERNETES_SERVICE_HOST indicates we're in a K8s pod
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		fmt.Println("Detected Kubernetes environment, using Kubernetes Jobs runtime")

		if namespace == "" {
			namespace = os.Getenv("NAMESPACE")
			if namespace == "" {
				namespace = "default"
			}
		}

		return k8sruntime.NewKubernetesRuntime(namespace)
	}

	// Default to Docker for local development
	fmt.Println("Using Docker runtime for local development")
	return dockerruntime.NewDockerRuntime()
}

// GetRuntimeType returns the runtime type that would be selected by auto-detection
func GetRuntimeType() RuntimeType {
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		return RuntimeTypeKubernetes
	}
	return RuntimeTypeDocker
}
