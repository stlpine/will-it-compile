package compiler

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/stlpine/will-it-compile/pkg/models"
	"github.com/stlpine/will-it-compile/pkg/runtime"
	internalruntime "github.com/stlpine/will-it-compile/internal/runtime"
)

// Compiler handles code compilation in isolated environments
type Compiler struct {
	runtime      runtime.CompilationRuntime
	environments map[string]models.EnvironmentSpec
}

// NewCompiler creates a new compiler instance with auto-detected runtime
// It loads environment configuration from YAML, with hardcoded fallback
func NewCompiler() (*Compiler, error) {
	// Auto-detect runtime (Docker or Kubernetes)
	rt, err := internalruntime.NewRuntimeAuto("")
	if err != nil {
		return nil, fmt.Errorf("failed to create runtime: %w", err)
	}

	// Load environments from YAML configuration
	config, err := LoadDefaultConfig()
	var environments map[string]models.EnvironmentSpec

	if err != nil {
		// Fallback to hardcoded configuration
		fmt.Printf("Warning: Failed to load config from YAML (%v), using hardcoded configuration\n", err)
		environments = getHardcodedEnvironments()
	} else {
		environments, err = config.ToEnvironmentSpecs()
		if err != nil {
			rt.Close()
			return nil, fmt.Errorf("failed to parse environment specs: %w", err)
		}
	}

	compiler := &Compiler{
		runtime:      rt,
		environments: environments,
	}

	// Verify required images exist at startup
	if err := compiler.verifyImages(context.Background()); err != nil {
		rt.Close()
		return nil, err
	}

	return compiler, nil
}

// NewCompilerWithRuntime creates a compiler with a custom runtime (useful for testing and explicit runtime selection)
// Uses hardcoded configuration by default, but can be customized after creation
func NewCompilerWithRuntime(rt runtime.CompilationRuntime) *Compiler {
	return &Compiler{
		runtime:      rt,
		environments: getHardcodedEnvironments(),
	}
}

// getHardcodedEnvironments returns the hardcoded fallback environment configuration
// This is used when YAML config cannot be loaded, or for testing
func getHardcodedEnvironments() map[string]models.EnvironmentSpec {
	return map[string]models.EnvironmentSpec{
		"cpp-gcc-13": {
			Language:     models.LanguageCpp,
			Compiler:     models.CompilerGCC13,
			Version:      "13",
			Standard:     models.StandardCpp20,
			Architecture: models.ArchX86_64,
			OS:           models.OSLinux,
			ImageTag:     "will-it-compile/cpp-gcc:13-alpine",
		},
	}
}

// Close cleans up resources
func (c *Compiler) Close() error {
	return c.runtime.Close()
}

// verifyImages checks that all required images exist
func (c *Compiler) verifyImages(ctx context.Context) error {
	missingImages := []string{}

	// Check each environment's image
	for envKey, envSpec := range c.environments {
		exists, err := c.runtime.ImageExists(ctx, envSpec.ImageTag)
		if err != nil {
			return fmt.Errorf("failed to check image %s: %w", envSpec.ImageTag, err)
		}
		if !exists {
			missingImages = append(missingImages, fmt.Sprintf("%s (%s)", envSpec.ImageTag, envKey))
		}
	}

	if len(missingImages) > 0 {
		imageList := ""
		for _, img := range missingImages {
			imageList += "\n  - " + img
		}
		return fmt.Errorf("missing required images:%s\n\nPlease build images using:\n  make docker-build\n  or: cd images/cpp && ./build.sh",
			imageList)
	}

	return nil
}

// Compile compiles the given code and returns the result
func (c *Compiler) Compile(ctx context.Context, job models.CompilationJob) models.CompilationResult {
	startTime := time.Now()

	// Validate the request
	if err := c.validateRequest(job.Request); err != nil {
		return models.CompilationResult{
			JobID:    job.ID,
			Success:  false,
			Compiled: false,
			Error:    err.Error(),
			Duration: time.Since(startTime),
		}
	}

	// Decode source code
	sourceCode, err := base64.StdEncoding.DecodeString(job.Request.Code)
	if err != nil {
		return models.CompilationResult{
			JobID:    job.ID,
			Success:  false,
			Compiled: false,
			Error:    "invalid base64 encoding",
			Duration: time.Since(startTime),
		}
	}

	// Select environment
	envSpec, err := c.selectEnvironment(job.Request)
	if err != nil {
		return models.CompilationResult{
			JobID:    job.ID,
			Success:  false,
			Compiled: false,
			Error:    err.Error(),
			Duration: time.Since(startTime),
		}
	}

	// Prepare runtime configuration
	config := runtime.CompilationConfig{
		JobID:      job.ID,
		ImageTag:   envSpec.ImageTag,
		SourceCode: string(sourceCode),
		WorkDir:    "/workspace",
		Env: []string{
			fmt.Sprintf("CPP_STANDARD=%s", envSpec.Standard),
			"SOURCE_FILE=/workspace/source.cpp",
			"COMPILE_TIMEOUT=25",
		},
		Timeout: 30 * time.Second,
	}

	// Run compilation
	output, err := c.runtime.Compile(ctx, config)
	if err != nil {
		return models.CompilationResult{
			JobID:    job.ID,
			Success:  false,
			Compiled: false,
			Error:    fmt.Sprintf("compilation failed: %v", err),
			Duration: time.Since(startTime),
		}
	}

	// Build result
	result := models.CompilationResult{
		JobID:    job.ID,
		Success:  true,
		Compiled: output.ExitCode == 0,
		Stdout:   output.Stdout,
		Stderr:   output.Stderr,
		ExitCode: output.ExitCode,
		Duration: output.Duration,
	}

	if output.TimedOut {
		result.Error = "compilation timeout"
	}

	return result
}

// validateRequest validates the compilation request
func (c *Compiler) validateRequest(req models.CompilationRequest) error {
	// Use the built-in Validate method
	if err := req.Validate(); err != nil {
		return err
	}

	// Check code size (base64 encoded)
	if len(req.Code) > 2*1024*1024 { // ~1.5MB source after decoding
		return fmt.Errorf("source code too large (max 1MB)")
	}

	// Validate language support (for MVP, only cpp is supported)
	normalizedLang := req.Language.Normalize()
	if normalizedLang != models.LanguageCpp {
		return fmt.Errorf("unsupported language: %s (supported: cpp)", req.Language)
	}

	return nil
}

// selectEnvironment selects the appropriate environment for compilation
func (c *Compiler) selectEnvironment(req models.CompilationRequest) (models.EnvironmentSpec, error) {
	// Normalize language
	language := req.Language.Normalize()

	// For MVP, we only support one environment
	compiler := req.Compiler
	if compiler == "" {
		compiler = models.CompilerGCC13
	}

	// Build environment key
	envKey := fmt.Sprintf("%s-%s", language, compiler)

	env, exists := c.environments[envKey]
	if !exists {
		return models.EnvironmentSpec{}, fmt.Errorf("unsupported environment: %s with %s", language, compiler)
	}

	// Override standard if specified
	if req.Standard != "" {
		env.Standard = req.Standard
	}

	return env, nil
}

// GetSupportedEnvironments returns a list of supported environments
func (c *Compiler) GetSupportedEnvironments() []models.Environment {
	return []models.Environment{
		{
			Language:  "cpp",
			Compilers: []string{"gcc-13"},
			Standards: []string{"c++11", "c++14", "c++17", "c++20", "c++23"},
			OSes:      []string{"linux"},
			Arches:    []string{"x86_64"},
		},
	}
}
