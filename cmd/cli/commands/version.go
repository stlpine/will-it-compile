package commands

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  `Display version, build, and runtime information for will-it-compile.`,
	Example: `  # Show version
  will-it-compile version

  # Short version output
  will-it-compile version --short`,
	Run: runVersion,
}

var versionShort bool

func init() {
	rootCmd.AddCommand(versionCmd)

	versionCmd.Flags().BoolVarP(&versionShort, "short", "s", false, "print just the version number")
}

func runVersion(cmd *cobra.Command, args []string) {
	if versionShort {
		fmt.Println(version)
		return
	}

	fmt.Printf("will-it-compile version %s\n", version)
	fmt.Printf("  Git commit:    %s\n", commit)
	fmt.Printf("  Built:         %s\n", buildDate)
	fmt.Printf("  Go version:    %s\n", runtime.Version())
	fmt.Printf("  OS/Arch:       %s/%s\n", runtime.GOOS, runtime.GOARCH)
}
