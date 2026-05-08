package main

import (
	"context"
	"fmt"
	"os"

	"github.com/kkaddal-bc/maestro/cli/internal/maestrocli"
)

func main() {
	if err := maestrocli.Execute(context.Background(), os.Args[1:], os.Stdout, os.Stderr); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
