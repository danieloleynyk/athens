package main

import (
	"github.com/gomods/athens/cmd/cli/commands"
	"os"
)

func main() {
	if err := commands.NewDefaultCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
