package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kkaddal-bc/maestro/packages/cli/internal/releaseassets"
)

func main() {
	var (
		version   = flag.String("version", "", "release version")
		skillsDir = flag.String("skills-dir", "", "skills directory")
		outputDir = flag.String("output-dir", "", "output directory")
	)
	flag.Parse()

	if *version == "" || *skillsDir == "" || *outputDir == "" {
		fmt.Fprintln(os.Stderr, "version, skills-dir, and output-dir are required")
		os.Exit(2)
	}

	if err := releaseassets.WriteArtifacts(*version, *skillsDir, *outputDir); err != nil {
		fmt.Fprintf(os.Stderr, "write artifacts: %v\n", err)
		os.Exit(1)
	}
}
