package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Version information (set by build flags).
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "will-it-compile",
	Short: "A secure code compilation checker",
	Long: `will-it-compile is a CLI tool for checking if code compiles
in isolated, secure environments.

It supports multiple languages and compilers, running each compilation
in a sandboxed Docker container with strict resource limits.`,
	Version: version,
	Example: `  # Compile a C++ file
  will-it-compile compile mycode.cpp

  # Compile with specific standard
  will-it-compile compile mycode.cpp --std=c++20

  # List supported environments
  will-it-compile environments

  # Show version
  will-it-compile version`,
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Set custom version template
	rootCmd.SetVersionTemplate(fmt.Sprintf("will-it-compile version %s (commit: %s, built: %s)\n", version, commit, buildDate))

	// Global flags
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "quiet mode (errors only)")
}

// isVerbose returns true if verbose flag is set.
func isVerbose(cmd *cobra.Command) bool {
	verbose, _ := cmd.Flags().GetBool("verbose")
	return verbose
}

// isQuiet returns true if quiet flag is set.
func isQuiet(cmd *cobra.Command) bool {
	quiet, _ := cmd.Flags().GetBool("quiet")
	return quiet
}

// printInfo prints informational messages (unless quiet mode).
func printInfo(cmd *cobra.Command, format string, args ...interface{}) {
	if !isQuiet(cmd) {
		fmt.Fprintf(os.Stdout, format+"\n", args...)
	}
}

// printVerbose prints verbose messages (only in verbose mode).
func printVerbose(cmd *cobra.Command, format string, args ...interface{}) {
	if isVerbose(cmd) {
		fmt.Fprintf(os.Stdout, "[VERBOSE] "+format+"\n", args...)
	}
}

// printError prints error messages.
func printError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
}
