package compiler

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/stlpine/will-it-compile/pkg/models"
	"gopkg.in/yaml.v3"
)

// Config represents the parsed configuration from environments.yaml
type Config struct {
	Environments []EnvironmentConfig `yaml:"environments"`
	Limits       LimitsConfig        `yaml:"limits"`
	RateLimits   RateLimitsConfig    `yaml:"rate_limits"`
}

// EnvironmentConfig represents a language environment configuration
type EnvironmentConfig struct {
	Language  string            `yaml:"language"`
	Compilers []CompilerConfig  `yaml:"compilers"`
}

// CompilerConfig represents a compiler configuration
type CompilerConfig struct {
	Name          string   `yaml:"name"`
	Version       string   `yaml:"version"`
	Image         string   `yaml:"image"`
	Standards     []string `yaml:"standards"`
	Architectures []string `yaml:"architectures"`
	OSes          []string `yaml:"oses"`
}

// LimitsConfig represents resource limits
type LimitsConfig struct {
	MaxSourceSizeMB            int `yaml:"max_source_size_mb"`
	MaxCompilationTimeSeconds  int `yaml:"max_compilation_time_seconds"`
	MaxOutputSizeMB            int `yaml:"max_output_size_mb"`
	MaxMemoryMB                int `yaml:"max_memory_mb"`
	MaxCPUQuota                int `yaml:"max_cpu_quota"`
}

// RateLimitsConfig represents rate limiting configuration
type RateLimitsConfig struct {
	RequestsPerMinute int `yaml:"requests_per_minute"`
	Burst             int `yaml:"burst"`
}

// LoadConfig loads the configuration from a YAML file
func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate the configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if len(c.Environments) == 0 {
		return fmt.Errorf("no environments defined")
	}

	for i, env := range c.Environments {
		if env.Language == "" {
			return fmt.Errorf("environment[%d]: language is required", i)
		}
		if len(env.Compilers) == 0 {
			return fmt.Errorf("environment[%d]: no compilers defined for language %s", i, env.Language)
		}
		for j, comp := range env.Compilers {
			if comp.Name == "" {
				return fmt.Errorf("environment[%d].compiler[%d]: compiler name is required", i, j)
			}
			if comp.Version == "" {
				return fmt.Errorf("environment[%d].compiler[%d]: compiler version is required", i, j)
			}
			if comp.Image == "" {
				return fmt.Errorf("environment[%d].compiler[%d]: image is required", i, j)
			}
		}
	}

	return nil
}

// ToEnvironmentSpecs converts the configuration to a map of EnvironmentSpec
func (c *Config) ToEnvironmentSpecs() (map[string]models.EnvironmentSpec, error) {
	envSpecs := make(map[string]models.EnvironmentSpec)

	for _, envConfig := range c.Environments {
		// Parse language
		language := models.Language(envConfig.Language)
		if !language.Valid() {
			return nil, fmt.Errorf("unsupported language: %s", envConfig.Language)
		}
		language = language.Normalize()

		for _, compConfig := range envConfig.Compilers {
			// Build compiler identifier (e.g., "gcc-13")
			compilerID := fmt.Sprintf("%s-%s", compConfig.Name, compConfig.Version)
			compiler := models.Compiler(compilerID)

			// For MVP, we only validate known compilers, but allow others for future expansion
			// This allows the YAML to define future compilers without code changes

			// Get default standard (first one if available)
			var defaultStandard models.Standard
			if len(compConfig.Standards) > 0 {
				defaultStandard = models.Standard(compConfig.Standards[0])
			}

			// Get default architecture (first one if available)
			var defaultArch models.Architecture
			if len(compConfig.Architectures) > 0 {
				defaultArch = models.Architecture(compConfig.Architectures[0])
			} else {
				defaultArch = models.ArchX86_64 // Fallback to x86_64
			}

			// Get default OS (first one if available)
			var defaultOS models.OS
			if len(compConfig.OSes) > 0 {
				defaultOS = models.OS(compConfig.OSes[0])
			} else {
				defaultOS = models.OSLinux // Fallback to Linux
			}

			// Create environment key (e.g., "cpp-gcc-13")
			envKey := fmt.Sprintf("%s-%s", language, compilerID)

			// Create environment spec
			envSpecs[envKey] = models.EnvironmentSpec{
				Language:     language,
				Compiler:     compiler,
				Version:      compConfig.Version,
				Standard:     defaultStandard,
				Architecture: defaultArch,
				OS:           defaultOS,
				ImageTag:     compConfig.Image,
			}
		}
	}

	return envSpecs, nil
}

// GetDefaultConfigPath returns the default path to the configuration file
func GetDefaultConfigPath() string {
	// Try to find the config file relative to the project root
	// This works when running from project directory or bin directory
	candidates := []string{
		"configs/environments.yaml",                    // When running from project root
		"../configs/environments.yaml",                 // When running from bin/
		"../../configs/environments.yaml",              // When running from test directories
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			absPath, _ := filepath.Abs(candidate)
			return absPath
		}
	}

	// Fallback to default location
	return "configs/environments.yaml"
}

// LoadDefaultConfig loads the configuration from the default location
func LoadDefaultConfig() (*Config, error) {
	configPath := GetDefaultConfigPath()
	return LoadConfig(configPath)
}
