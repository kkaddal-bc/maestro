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
	cmd := newNotImplementedCommand("list", "List related commands")

	cmd.AddCommand(newListSkillsCommand())

	return cmd
}

func newListSkillsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "skills",
		Short: "List maestro skills",
		Args:  cobra.NoArgs,
		RunE:  runListSkills,
	}

	return cmd
}

func runListSkills(cmd *cobra.Command, _ []string) error {
	client := newSkillsFetcher()

	manifestData, err := client.FetchManifest()
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

func printListSkills(out io.Writer, home string, manifestData *manifest.Manifest, activeTargets []targets.Target) {
	headers := []string{"SKILL", "DESCRIPTION"}
	for _, target := range activeTargets {
		headers = append(headers, displayTargetPath(home, target.Path))
	}
	fmt.Fprintln(out, strings.Join(headers, "\t"))

	for _, skill := range manifestData.Skills {
		row := []string{skill.Name, skill.Description}
		for _, target := range activeTargets {
			row = append(row, skillStatus(target.Path, skill.Name))
		}
		fmt.Fprintln(out, strings.Join(row, "\t"))
	}
}

func skillStatus(targetPath, skillName string) string {
	info, err := os.Stat(filepath.Join(targetPath, skillName))
	if err == nil && info.IsDir() {
		return "installed"
	}
	return "not installed"
}
