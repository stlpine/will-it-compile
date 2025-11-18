package compiler

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	internalruntime "github.com/stlpine/will-it-compile/internal/runtime"
	"github.com/stlpine/will-it-compile/pkg/models"
	"github.com/stlpine/will-it-compile/pkg/runtime"
)

// Sentinel errors for compiler package.
var (
	ErrMissingRequiredImages  = errors.New("missing required Docker images")
	ErrSourceCodeTooLarge     = errors.New("source code too large (max 1MB)")
	ErrUnsupportedLanguage    = errors.New("unsupported language")
	ErrUnsupportedEnvironment = errors.New("unsupported environment")
)

// Compiler handles code compilation in isolated environments.
type Compiler struct {
	runtime      runtime.CompilationRuntime
	environments map[string]models.EnvironmentSpec
}

// NewCompiler creates a new compiler instance with auto-detected runtime
// It loads environment configuration from YAML, with hardcoded fallback.
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
			_ = rt.Close() //nolint:errcheck // already in error path
			return nil, fmt.Errorf("failed to parse environment specs: %w", err)
		}
	}

	compiler := &Compiler{
		runtime:      rt,
		environments: environments,
	}

	// Verify required images exist at startup
	if err := compiler.verifyImages(context.Background()); err != nil {
		_ = rt.Close() //nolint:errcheck // already in error path
		return nil, err
	}

	return compiler, nil
}

// NewCompilerWithRuntime creates a compiler with a custom runtime (useful for testing and explicit runtime selection)
// Uses hardcoded configuration by default, but can be customized after creation.
func NewCompilerWithRuntime(rt runtime.CompilationRuntime) *Compiler {
	return &Compiler{
		runtime:      rt,
		environments: getHardcodedEnvironments(),
	}
}

// getHardcodedEnvironments returns the hardcoded fallback environment configuration
// This is used when YAML config cannot be loaded, or for testing.
func getHardcodedEnvironments() map[string]models.EnvironmentSpec {
	return map[string]models.EnvironmentSpec{
		"cpp-gcc-13": {
			Language:     models.LanguageCpp,
			Compiler:     models.CompilerGCC13,
			Version:      "13",
			Standard:     models.StandardCpp20,
			Architecture: models.ArchX86_64,
			OS:           models.OSLinux,
			ImageTag:     "gcc:13",
		},
		"c-gcc-13": {
			Language:     models.LanguageC,
			Compiler:     models.CompilerGCC13,
			Version:      "13",
			Standard:     models.StandardC17,
			Architecture: models.ArchX86_64,
			OS:           models.OSLinux,
			ImageTag:     "gcc:13",
		},
		"go-go": {
			Language:     models.LanguageGo,
			Compiler:     models.CompilerGo,
			Version:      "1.23",
			Architecture: models.ArchX86_64,
			OS:           models.OSLinux,
			ImageTag:     "golang:1.23-alpine",
		},
		"rust-rustc": {
			Language:     models.LanguageRust,
			Compiler:     models.CompilerRustc,
			Version:      "1.80",
			Architecture: models.ArchX86_64,
			OS:           models.OSLinux,
			ImageTag:     "rust:1.80-alpine",
		},
	}
}

// Close cleans up resources.
func (c *Compiler) Close() error {
	return c.runtime.Close()
}

// verifyImages checks that all required images exist.
func (c *Compiler) verifyImages(ctx context.Context) error {
	// In integration tests (CI), we only pull gcc:9 to speed up the pipeline.
	// MINIMAL_IMAGE_VALIDATION=true checks only for gcc:9 instead of all images.
	// This must match the image pulled in .github/workflows/pr-ci.yml
	if os.Getenv("MINIMAL_IMAGE_VALIDATION") == "true" {
		minimalImage := "gcc:9"
		exists, err := c.runtime.ImageExists(ctx, minimalImage)
		if err != nil {
			return fmt.Errorf("failed to check image %s: %w", minimalImage, err)
		}
		if !exists {
			return fmt.Errorf("%w: %s\n\nPlease pull official images using:\n  make docker-pull\n  or: docker pull %s",
				ErrMissingRequiredImages, minimalImage, minimalImage)
		}
		return nil
	}

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
		var imageListSb106 strings.Builder
		for _, img := range missingImages {
			imageListSb106.WriteString("\n  - " + img)
		}
		imageList += imageListSb106.String()
		return fmt.Errorf("%w:%s\n\nPlease pull official images using:\n  make docker-pull\n  or: docker pull gcc:13",
			ErrMissingRequiredImages, imageList)
	}

	return nil
}

// Compile compiles the given code and returns the result.
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

	// Determine source filename based on language
	sourceFilename := c.getSourceFilename(envSpec.Language)

	// Build compile command based on language
	compileCmd := c.buildCompileCommand(envSpec, sourceFilename)

	// Prepare runtime configuration
	config := runtime.CompilationConfig{
		JobID:          job.ID,
		ImageTag:       envSpec.ImageTag,
		SourceCode:     string(sourceCode),
		SourceFilename: sourceFilename,
		CompileCommand: compileCmd,
		WorkDir:        "/workspace",
		Env: []string{
			fmt.Sprintf("STANDARD=%s", envSpec.Standard),
			fmt.Sprintf("SOURCE_FILE=/workspace/%s", sourceFilename),
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

// validateRequest validates the compilation request.
func (c *Compiler) validateRequest(req models.CompilationRequest) error {
	// Use the built-in Validate method
	if err := req.Validate(); err != nil {
		return err
	}

	// Check code size (base64 encoded)
	if len(req.Code) > 2*1024*1024 { // ~1.5MB source after decoding
		return ErrSourceCodeTooLarge
	}

	// Validate language support - check if we have environments for this language
	normalizedLang := req.Language.Normalize()

	// Check if any environment supports this language
	hasSupport := false
	for _, env := range c.environments {
		if env.Language == normalizedLang {
			hasSupport = true
			break
		}
	}

	if !hasSupport {
		return fmt.Errorf("%w: %s", ErrUnsupportedLanguage, req.Language)
	}

	return nil
}

// selectEnvironment selects the appropriate environment for compilation.
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
		return models.EnvironmentSpec{}, fmt.Errorf("%w: %s with %s", ErrUnsupportedEnvironment, language, compiler)
	}

	// Override standard if specified
	if req.Standard != "" {
		env.Standard = req.Standard
	}

	return env, nil
}

// buildCompileCommand builds the compilation command based on the environment.
func (c *Compiler) buildCompileCommand(env models.EnvironmentSpec, sourceFilename string) string {
	// Build command based on language
	// Note: stderr is NOT redirected to stdout so errors appear in stderr field
	switch env.Language {
	case models.LanguageCpp, models.LanguageC:
		// C/C++ compilation with GCC or Clang
		return fmt.Sprintf("g++ -std=%s /workspace/%s -o /workspace/output", env.Standard, sourceFilename)

	case models.LanguageGo:
		// Go compilation
		return fmt.Sprintf("go build -o /workspace/output /workspace/%s", sourceFilename)

	case models.LanguageRust:
		// Rust compilation
		return fmt.Sprintf("rustc /workspace/%s -o /workspace/output", sourceFilename)

	default:
		// Fallback to C++ (should not happen due to validation)
		return fmt.Sprintf("g++ -std=%s /workspace/%s -o /workspace/output", env.Standard, sourceFilename)
	}
}

// getSourceFilename returns the appropriate source filename based on language.
func (c *Compiler) getSourceFilename(language models.Language) string {
	switch language {
	case models.LanguageC:
		return "source.c"
	case models.LanguageCpp:
		return "source.cpp"
	case models.LanguageGo:
		return "main.go"
	case models.LanguageRust:
		return "main.rs"
	default:
		return "source.cpp"
	}
}

// GetSupportedEnvironments returns a list of supported environments.
func (c *Compiler) GetSupportedEnvironments() []models.Environment {
	// Group environment specs by language
	langMap := make(map[models.Language]*models.Environment)

	for _, envSpec := range c.environments {
		lang := envSpec.Language

		// Initialize environment for this language if not exists
		if langMap[lang] == nil {
			langMap[lang] = &models.Environment{
				Language:  string(lang),
				Compilers: []string{},
				Standards: []string{},
				OSes:      []string{},
				Arches:    []string{},
			}
		}

		env := langMap[lang]

		// Add compiler if not already in list
		compilerStr := string(envSpec.Compiler)
		if !contains(env.Compilers, compilerStr) {
			env.Compilers = append(env.Compilers, compilerStr)
		}

		// Add standard if not already in list
		standardStr := string(envSpec.Standard)
		if standardStr != "" && !contains(env.Standards, standardStr) {
			env.Standards = append(env.Standards, standardStr)
		}

		// Add OS if not already in list
		osStr := string(envSpec.OS)
		if !contains(env.OSes, osStr) {
			env.OSes = append(env.OSes, osStr)
		}

		// Add architecture if not already in list
		archStr := string(envSpec.Architecture)
		if !contains(env.Arches, archStr) {
			env.Arches = append(env.Arches, archStr)
		}
	}

	// Convert map to slice
	result := make([]models.Environment, 0, len(langMap))
	for _, env := range langMap {
		result = append(result, *env)
	}

	return result
}

// contains checks if a string slice contains a specific string.
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
