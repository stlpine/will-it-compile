package commands

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/stlpine/will-it-compile/internal/compiler"
	"github.com/stlpine/will-it-compile/internal/runtime/docker"
	"github.com/stlpine/will-it-compile/pkg/models"
)

// Sentinel errors for compile command.
var (
	ErrCompilationFailed  = errors.New("compilation failed")
	ErrUnsupportedFileExt = errors.New("unsupported file extension")
)

var compileCmd = &cobra.Command{
	Use:   "compile <file>",
	Short: "Compile a source code file",
	Long: `Compile a source code file in an isolated Docker container.

The file type is detected automatically from the extension.
Compilation happens in a secure, sandboxed environment with
resource limits and no network access.`,
	Example: `  # Compile a C++ file
  will-it-compile compile mycode.cpp

  # Compile with specific C++ standard
  will-it-compile compile mycode.cpp --std=c++20

  # Compile with specific compiler
  will-it-compile compile mycode.cpp --compiler=gcc-13

  # Verbose output
  will-it-compile compile mycode.cpp --verbose`,
	Args: cobra.ExactArgs(1),
	RunE: runCompile,
}

var (
	compileStandard   string
	compileCompiler   string
	compileTimeout    int
	compileShowStdout bool
	compileShowStderr bool
)

func init() {
	rootCmd.AddCommand(compileCmd)

	// Flags
	compileCmd.Flags().StringVar(&compileStandard, "std", "", "language standard (e.g., c++20, c++17)")
	compileCmd.Flags().StringVar(&compileCompiler, "compiler", "", "compiler to use (e.g., gcc-13)")
	compileCmd.Flags().IntVar(&compileTimeout, "timeout", 30, "compilation timeout in seconds")
	compileCmd.Flags().BoolVar(&compileShowStdout, "stdout", true, "show compilation stdout")
	compileCmd.Flags().BoolVar(&compileShowStderr, "stderr", true, "show compilation stderr")
}

func runCompile(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	// Validate file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		printError("file not found: %s", filePath)
		return err
	}

	printVerbose(cmd, "Reading source file: %s", filePath)

	// Read source code
	sourceCode, err := os.ReadFile(filePath)
	if err != nil {
		printError("failed to read file: %v", err)
		return err
	}

	// Detect language from extension
	language, err := detectLanguage(filePath)
	if err != nil {
		printError("%v", err)
		return err
	}

	printVerbose(cmd, "Detected language: %s", language)

	// Encode source code
	encodedCode := base64.StdEncoding.EncodeToString(sourceCode)

	// Build compilation request
	request := models.CompilationRequest{
		Code:     encodedCode,
		Language: language,
	}

	if compileStandard != "" {
		request.Standard = models.Standard(compileStandard)
	}

	if compileCompiler != "" {
		request.Compiler = models.Compiler(compileCompiler)
	}

	printInfo(cmd, "Compiling %s...", filepath.Base(filePath))
	printVerbose(cmd, "Language: %s, Compiler: %s, Standard: %s", request.Language, request.Compiler, request.Standard)

	// Create Docker runtime (CLI always uses Docker)
	dockerRuntime, err := docker.NewDockerRuntime()
	if err != nil {
		printError("failed to create Docker runtime: %v", err)
		return err
	}
	defer dockerRuntime.Close()

	// Create compiler
	comp := compiler.NewCompilerWithRuntime(dockerRuntime)

	// Create compilation job
	job := models.CompilationJob{
		ID:      uuid.New().String(),
		Request: request,
	}

	// Run compilation with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(compileTimeout)*time.Second)
	defer cancel()

	startTime := time.Now()
	result := comp.Compile(ctx, job)
	duration := time.Since(startTime)

	// Print results
	printVerbose(cmd, "Compilation completed in %v", duration)

	if result.Success {
		if result.Compiled {
			printInfo(cmd, "✓ Compilation successful (exit code: %d, duration: %v)", result.ExitCode, result.Duration)
		} else {
			printInfo(cmd, "✗ Compilation failed (exit code: %d, duration: %v)", result.ExitCode, result.Duration)
		}
	} else {
		printError("compilation error: %s", result.Error)
		return ErrCompilationFailed
	}

	// Show stdout
	if compileShowStdout && result.Stdout != "" {
		if !isQuiet(cmd) {
			fmt.Println("\n--- Stdout ---")
			fmt.Println(result.Stdout)
		}
	}

	// Show stderr
	if compileShowStderr && result.Stderr != "" {
		if !isQuiet(cmd) {
			fmt.Println("\n--- Stderr ---")
			fmt.Println(result.Stderr)
		}
	}

	// Exit with error if compilation failed
	if !result.Compiled {
		return ErrCompilationFailed
	}

	return nil
}

// detectLanguage detects the programming language from file extension.
func detectLanguage(filePath string) (models.Language, error) {
	ext := filepath.Ext(filePath)

	switch ext {
	case ".cpp", ".cc", ".cxx", ".c++":
		return models.LanguageCpp, nil
	case ".c":
		return models.LanguageC, nil
	default:
		return "", fmt.Errorf("%w: %s (supported: .cpp, .cc, .cxx, .c++, .c)", ErrUnsupportedFileExt, ext)
	}
}
