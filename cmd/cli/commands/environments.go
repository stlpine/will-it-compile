package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stlpine/will-it-compile/internal/compiler"
	"github.com/stlpine/will-it-compile/internal/runtime/docker"
	"github.com/stlpine/will-it-compile/pkg/models"
)

var environmentsCmd = &cobra.Command{
	Use:     "environments",
	Aliases: []string{"env", "envs"},
	Short:   "List supported compilation environments",
	Long: `List all supported compilation environments including:
  - Programming languages
  - Available compilers
  - Language standards
  - Operating systems
  - Architectures`,
	Example: `  # List all environments
  will-it-compile environments

  # Short alias
  will-it-compile env`,
	RunE: runEnvironments,
}

var (
	envOutputFormat string
)

func init() {
	rootCmd.AddCommand(environmentsCmd)

	environmentsCmd.Flags().StringVarP(&envOutputFormat, "output", "o", "table", "output format: table, list, or json")
}

func runEnvironments(cmd *cobra.Command, args []string) error {
	printVerbose(cmd, "Fetching supported environments...")

	// Create Docker runtime
	dockerRuntime, err := docker.NewDockerRuntime()
	if err != nil {
		printError("failed to create Docker runtime: %v", err)
		return err
	}
	defer dockerRuntime.Close()

	// Create compiler
	comp := compiler.NewCompilerWithRuntime(dockerRuntime)

	// Get supported environments
	environments := comp.GetSupportedEnvironments()

	if len(environments) == 0 {
		printInfo(cmd, "No environments configured")
		return nil
	}

	// Print based on format
	switch envOutputFormat {
	case "table":
		printEnvironmentsTable(cmd, environments)
	case "list":
		printEnvironmentsList(cmd, environments)
	case "json":
		printEnvironmentsJSON(cmd, environments)
	default:
		printError("unknown output format: %s (use: table, list, json)", envOutputFormat)
		return fmt.Errorf("invalid output format")
	}

	return nil
}

func printEnvironmentsTable(cmd *cobra.Command, environments []models.Environment) {
	printInfo(cmd, "Supported Compilation Environments:\n")

	for _, env := range environments {
		printInfo(cmd, "Language: %s", env.Language)
		printInfo(cmd, "  Compilers:  %s", strings.Join(env.Compilers, ", "))
		printInfo(cmd, "  Standards:  %s", strings.Join(env.Standards, ", "))
		printInfo(cmd, "  OSes:       %s", strings.Join(env.OSes, ", "))
		printInfo(cmd, "  Arches:     %s", strings.Join(env.Arches, ", "))
		printInfo(cmd, "")
	}
}

func printEnvironmentsList(cmd *cobra.Command, environments []models.Environment) {
	for _, env := range environments {
		for _, compiler := range env.Compilers {
			for _, std := range env.Standards {
				printInfo(cmd, "%s (%s, %s)", env.Language, compiler, std)
			}
		}
	}
}

func printEnvironmentsJSON(cmd *cobra.Command, environments []models.Environment) {
	// Simple JSON output (could use encoding/json for better formatting)
	printInfo(cmd, "[")
	for i, env := range environments {
		printInfo(cmd, "  {")
		printInfo(cmd, "    \"language\": \"%s\",", env.Language)
		printInfo(cmd, "    \"compilers\": [%s],", joinQuoted(env.Compilers))
		printInfo(cmd, "    \"standards\": [%s],", joinQuoted(env.Standards))
		printInfo(cmd, "    \"oses\": [%s],", joinQuoted(env.OSes))
		printInfo(cmd, "    \"arches\": [%s]", joinQuoted(env.Arches))
		if i < len(environments)-1 {
			printInfo(cmd, "  },")
		} else {
			printInfo(cmd, "  }")
		}
	}
	printInfo(cmd, "]")
}

func joinQuoted(items []string) string {
	quoted := make([]string, len(items))
	for i, item := range items {
		quoted[i] = fmt.Sprintf("\"%s\"", item)
	}
	return strings.Join(quoted, ", ")
}
