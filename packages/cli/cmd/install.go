package cmd

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/kkaddal-bc/maestro/packages/cli/internal/fetcher"
	"github.com/kkaddal-bc/maestro/packages/cli/internal/installer"
	"github.com/kkaddal-bc/maestro/packages/cli/internal/manifest"
	"github.com/kkaddal-bc/maestro/packages/cli/internal/targets"
	"github.com/spf13/cobra"
)

type skillsFetcher interface {
	FetchManifest() (*manifest.Manifest, error)
	FetchSkillsArchive(version string) (io.ReadCloser, error)
}

var (
	newSkillsFetcher = func() skillsFetcher {
		return fetcher.New()
	}
	detectInstallTargets = targets.Detect
	runInstaller         = installer.Install
)

func newInstallCommand() *cobra.Command {
	cmd := newNotImplementedCommand("install", "Install related commands")

	cmd.AddCommand(newInstallSkillsCommand())

	return cmd
}

func newInstallSkillsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "skills [skill-name]",
		Short: "Install maestro skills",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runInstallSkills,
	}

	return cmd
}

func runInstallSkills(cmd *cobra.Command, args []string) error {
	client := newSkillsFetcher()

	manifestData, err := client.FetchManifest()
	if err != nil {
		return err
	}

	requested, err := selectSkills(manifestData, args)
	if err != nil {
		return err
	}

	archive, err := client.FetchSkillsArchive(manifestData.Version)
	if err != nil {
		return err
	}
	defer archive.Close()

	gz, err := gzip.NewReader(archive)
	if err != nil {
		return fmt.Errorf("open skills archive: %w", err)
	}
	defer gz.Close()

	installTargets := detectInstallTargets()
	result, err := runInstaller(requested, gz, installTargets)
	if err != nil {
		return err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	printInstallSummary(cmd.OutOrStdout(), home, installTargets, result)
	return nil
}

func selectSkills(manifestData *manifest.Manifest, args []string) ([]string, error) {
	if len(args) == 1 {
		name := args[0]
		if !manifestHasSkill(manifestData, name) {
			return nil, fmt.Errorf("unknown skill %q", name)
		}
		return []string{name}, nil
	}

	if len(manifestData.Skills) == 0 {
		return nil, errors.New("manifest contains no skills")
	}

	skills := make([]string, 0, len(manifestData.Skills))
	for _, skill := range manifestData.Skills {
		skills = append(skills, skill.Name)
	}
	return skills, nil
}

func manifestHasSkill(manifestData *manifest.Manifest, name string) bool {
	for _, skill := range manifestData.Skills {
		if skill.Name == name {
			return true
		}
	}
	return false
}

func printInstallSummary(out io.Writer, home string, activeTargets []targets.Target, result installer.Result) {
	active := map[string]struct{}{}
	for _, target := range activeTargets {
		active[target.Path] = struct{}{}
	}

	installedByTarget := map[string][]string{}
	for _, item := range result.Installed {
		installedByTarget[item.Target] = append(installedByTarget[item.Target], item.Skill)
	}

	for _, target := range targets.Known(home) {
		if _, ok := active[target.Path]; ok {
			for _, skill := range installedByTarget[target.Path] {
				fmt.Fprintf(out, "✓ installed %s → %s\n", skill, displayTargetPath(home, target.Path))
			}
			continue
		}

		fmt.Fprintf(out, "- skipped %s (not found)\n", displayTargetPath(home, target.Path))
	}
}

func displayTargetPath(home, targetPath string) string {
	rel, err := filepath.Rel(home, targetPath)
	if err != nil || strings.HasPrefix(rel, "..") {
		return targetPath
	}
	rel = filepath.ToSlash(rel)
	if rel == "." {
		return "~/"
	}
	return "~/" + rel + "/"
}
