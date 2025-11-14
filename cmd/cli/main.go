package main

import (
	"os"

	"github.com/stlpine/will-it-compile/cmd/cli/commands"
)

func main() {
	if err := commands.Execute(); err != nil {
		os.Exit(1)
	}
}
