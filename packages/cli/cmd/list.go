package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/kkaddal-bc/maestro/packages/cli/internal/manifest"
	"github.com/kkaddal-bc/maestro/packages/cli/internal/targets"
	"github.com/spf13/cobra"
)

func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List maestro skills",
		Args:  cobra.NoArgs,
		RunE:  runListSkills,
	}

	return cmd
}

func runListSkills(cmd *cobra.Command, _ []string) error {
	fetcher := newSkillsFetcher()

	manifestData, err := fetcher.FetchManifest()
	if err != nil {
		return err
	}

	installTargets := detectInstallTargets()
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	printListSkills(cmd.OutOrStdout(), home, manifestData, installTargets)
	return nil
}

func printListSkills(out io.Writer, home string, manifestData *manifest.Manifest, installTargets []targets.Target) {
	headers := []string{"SKILL", "DESCRIPTION"}
	for _, target := range installTargets {
		headers = append(headers, displayTargetPath(home, target.Path))
	}
	fmt.Fprintln(out, strings.Join(headers, "\t"))

	for _, skill := range manifestData.Skills {
		row := []string{skill.Name, skill.Description}
		for _, target := range installTargets {
			row = append(row, skillStatus(target.Path, skill.Name))
		}
		fmt.Fprintln(out, strings.Join(row, "\t"))
	}
}

func skillStatus(targetPath, skillName string) string {
	info, err := os.Stat(filepath.Join(targetPath, skillName))
	if err != nil || !info.IsDir() {
		return "not installed"
	}
	return "installed"
}
