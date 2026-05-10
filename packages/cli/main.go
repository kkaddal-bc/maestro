package main

import (
	"os"

	"github.com/kkaddal-bc/maestro/packages/cli/cmd"
)

var version = "dev"

func main() {
	if err := cmd.NewRootCommand(version).Execute(); err != nil {
		os.Exit(1)
	}
}
