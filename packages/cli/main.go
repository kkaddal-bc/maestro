package main

import (
	"os"

	"github.com/kkaddal-bc/maestro/packages/cli/cmd"
)

func main() {
	if err := cmd.NewRootCommand("dev").Execute(); err != nil {
		os.Exit(1)
	}
}
