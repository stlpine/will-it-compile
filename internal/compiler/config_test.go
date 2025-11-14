package compiler

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stlpine/will-it-compile/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	// Find the actual config file
	configPath := "../../../configs/environments.yaml"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Try alternative path (when running from different directory)
		configPath = "../../configs/environments.yaml"
	}

	config, err := LoadConfig(configPath)
	require.NoError(t, err, "Should load valid config file")
	require.NotNil(t, config)

	// Verify structure
	assert.NotEmpty(t, config.Environments, "Should have environments")
	assert.NotZero(t, config.Limits.MaxSourceSizeMB, "Should have limits")
	assert.NotZero(t, config.RateLimits.RequestsPerMinute, "Should have rate limits")
}

func TestLoadConfig_InvalidPath(t *testing.T) {
	config, err := LoadConfig("/nonexistent/path/to/config.yaml")
	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "failed to read config file")
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	// Create a temporary invalid YAML file
	tmpDir := t.TempDir()
	invalidFile := filepath.Join(tmpDir, "invalid.yaml")
	err := os.WriteFile(invalidFile, []byte("invalid: yaml: content: ["), 0644)
	require.NoError(t, err)

	config, err := LoadConfig(invalidFile)
	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "failed to parse config file")
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid_config",
			config: Config{
				Environments: []EnvironmentConfig{
					{
						Language: "cpp",
						Compilers: []CompilerConfig{
							{
								Name:    "gcc",
								Version: "13",
								Image:   "will-it-compile/cpp-gcc:13-alpine",
							},
						},
					},
				},
			},
			expectErr: false,
		},
		{
			name: "no_environments",
			config: Config{
				Environments: []EnvironmentConfig{},
			},
			expectErr: true,
			errMsg:    "no environments defined",
		},
		{
			name: "missing_language",
			config: Config{
				Environments: []EnvironmentConfig{
					{
						Language: "",
						Compilers: []CompilerConfig{
							{Name: "gcc", Version: "13", Image: "test"},
						},
					},
				},
			},
			expectErr: true,
			errMsg:    "language is required",
		},
		{
			name: "no_compilers",
			config: Config{
				Environments: []EnvironmentConfig{
					{
						Language:  "cpp",
						Compilers: []CompilerConfig{},
					},
				},
			},
			expectErr: true,
			errMsg:    "no compilers defined",
		},
		{
			name: "missing_compiler_name",
			config: Config{
				Environments: []EnvironmentConfig{
					{
						Language: "cpp",
						Compilers: []CompilerConfig{
							{Name: "", Version: "13", Image: "test"},
						},
					},
				},
			},
			expectErr: true,
			errMsg:    "compiler name is required",
		},
		{
			name: "missing_compiler_version",
			config: Config{
				Environments: []EnvironmentConfig{
					{
						Language: "cpp",
						Compilers: []CompilerConfig{
							{Name: "gcc", Version: "", Image: "test"},
						},
					},
				},
			},
			expectErr: true,
			errMsg:    "compiler version is required",
		},
		{
			name: "missing_image",
			config: Config{
				Environments: []EnvironmentConfig{
					{
						Language: "cpp",
						Compilers: []CompilerConfig{
							{Name: "gcc", Version: "13", Image: ""},
						},
					},
				},
			},
			expectErr: true,
			errMsg:    "image is required",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.config.Validate()
			if tc.expectErr {
				assert.Error(t, err)
				if tc.errMsg != "" {
					assert.Contains(t, err.Error(), tc.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigToEnvironmentSpecs(t *testing.T) {
	config := Config{
		Environments: []EnvironmentConfig{
			{
				Language: "cpp",
				Compilers: []CompilerConfig{
					{
						Name:          "gcc",
						Version:       "13",
						Image:         "will-it-compile/cpp-gcc:13-alpine",
						Standards:     []string{"c++11", "c++14", "c++17", "c++20", "c++23"},
						Architectures: []string{"x86_64"},
						OSes:          []string{"linux"},
					},
				},
			},
		},
	}

	envSpecs, err := config.ToEnvironmentSpecs()
	require.NoError(t, err)
	require.NotEmpty(t, envSpecs)

	// Should create "cpp-gcc-13" environment
	spec, exists := envSpecs["cpp-gcc-13"]
	assert.True(t, exists, "Should have cpp-gcc-13 environment")
	assert.Equal(t, models.LanguageCpp, spec.Language)
	assert.Equal(t, models.Compiler("gcc-13"), spec.Compiler)
	assert.Equal(t, "13", spec.Version)
	assert.Equal(t, models.Standard("c++11"), spec.Standard) // First standard is default
	assert.Equal(t, models.ArchX86_64, spec.Architecture)
	assert.Equal(t, models.OSLinux, spec.OS)
	assert.Equal(t, "will-it-compile/cpp-gcc:13-alpine", spec.ImageTag)
}

func TestConfigToEnvironmentSpecs_UnsupportedLanguage(t *testing.T) {
	config := Config{
		Environments: []EnvironmentConfig{
			{
				Language: "invalid-lang",
				Compilers: []CompilerConfig{
					{
						Name:    "gcc",
						Version: "13",
						Image:   "test",
					},
				},
			},
		},
	}

	envSpecs, err := config.ToEnvironmentSpecs()
	assert.Error(t, err)
	assert.Nil(t, envSpecs)
	assert.Contains(t, err.Error(), "unsupported language")
}

func TestConfigToEnvironmentSpecs_DefaultValues(t *testing.T) {
	config := Config{
		Environments: []EnvironmentConfig{
			{
				Language: "cpp",
				Compilers: []CompilerConfig{
					{
						Name:    "gcc",
						Version: "13",
						Image:   "test-image",
						// No standards, architectures, or OSes specified
					},
				},
			},
		},
	}

	envSpecs, err := config.ToEnvironmentSpecs()
	require.NoError(t, err)

	spec := envSpecs["cpp-gcc-13"]
	assert.Equal(t, models.ArchX86_64, spec.Architecture, "Should default to x86_64")
	assert.Equal(t, models.OSLinux, spec.OS, "Should default to Linux")
}

func TestGetDefaultConfigPath(t *testing.T) {
	path := GetDefaultConfigPath()
	assert.NotEmpty(t, path)
	// Just verify it returns a path - actual existence depends on where test runs
}

func TestLoadDefaultConfig(t *testing.T) {
	// This test may fail if not run from project root or if config doesn't exist
	// We make it resilient by checking if the file exists first
	configPath := GetDefaultConfigPath()
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Skip("Config file not found at default path, skipping")
	}

	config, err := LoadDefaultConfig()
	require.NoError(t, err)
	require.NotNil(t, config)
}

func TestGetHardcodedEnvironments(t *testing.T) {
	envs := getHardcodedEnvironments()
	assert.NotEmpty(t, envs, "Should have hardcoded environments")

	// Should have at least cpp-gcc-13
	spec, exists := envs["cpp-gcc-13"]
	assert.True(t, exists, "Should have cpp-gcc-13 environment")
	assert.Equal(t, models.LanguageCpp, spec.Language)
	assert.Equal(t, models.CompilerGCC13, spec.Compiler)
	assert.Equal(t, "will-it-compile/cpp-gcc:13-alpine", spec.ImageTag)
}
