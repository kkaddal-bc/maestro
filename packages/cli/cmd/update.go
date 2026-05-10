package cmd

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/kkaddal-bc/maestro/packages/cli/internal/fetcher"
	"github.com/kkaddal-bc/maestro/packages/cli/internal/installer"
	"github.com/kkaddal-bc/maestro/packages/cli/internal/manifest"
	"github.com/kkaddal-bc/maestro/packages/cli/internal/targets"
	"github.com/spf13/cobra"
)

type updateSkillsFetcher interface {
	FetchManifest() (*manifest.Manifest, error)
	FetchSkillsArchive(version string) (io.ReadCloser, error)
}

var (
	newUpdateSkillsFetcher = func() updateSkillsFetcher {
		return fetcher.New()
	}
)

func newUpdateCommand() *cobra.Command {
	cmd := newNotImplementedCommand("update", "Update related commands")

	cmd.AddCommand(newUpdateSkillsCommand())

	return cmd
}

func newUpdateSkillsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "skills",
		Short: "Update maestro skills",
		Args:  cobra.NoArgs,
		RunE:  runUpdateSkills,
	}
}

func runUpdateSkills(cmd *cobra.Command, _ []string) error {
	client := newUpdateSkillsFetcher()

	manifestData, err := client.FetchManifest()
	if err != nil {
		return err
	}

	installTargets := detectInstallTargets()
	selected := installedSkills(manifestData, installTargets)
	if len(selected) == 0 {
		printUpToDateSummary(cmd.OutOrStdout())
		return nil
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

	result, err := installer.Update(selected, gz, installTargets)
	if err != nil {
		return err
	}
	if result.UpToDate {
		printUpToDateSummary(cmd.OutOrStdout())
		return nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	printUpdateSummary(cmd.OutOrStdout(), home, result)
	return nil
}

func installedSkills(manifestData *manifest.Manifest, installTargets []targets.Target) []string {
	installed := map[string]struct{}{}
	for _, skill := range manifestData.Skills {
		for _, target := range installTargets {
			if skillInstalled(target.Path, skill.Name) {
				installed[skill.Name] = struct{}{}
				break
			}
		}
	}

	skills := make([]string, 0, len(installed))
	for _, skill := range manifestData.Skills {
		if _, ok := installed[skill.Name]; ok {
			skills = append(skills, skill.Name)
		}
	}
	return skills
}

func skillInstalled(targetRoot, skillName string) bool {
	info, err := os.Stat(filepath.Join(targetRoot, skillName))
	return err == nil && info.IsDir()
}

func printUpdateSummary(out io.Writer, home string, result installer.UpdateResult) {
	updatedByTarget := map[string][]string{}
	for _, item := range result.Updated {
		updatedByTarget[item.Target] = append(updatedByTarget[item.Target], item.Skill)
	}

	for _, target := range targets.Known(home) {
		if skills, ok := updatedByTarget[target.Path]; ok {
			for _, skill := range skills {
				fmt.Fprintf(out, "✓ updated %s → %s\n", skill, displayTargetPath(home, target.Path))
			}
		}
	}
}

func printUpToDateSummary(out io.Writer) {
	_, _ = fmt.Fprintln(out, "skills are already up to date")
}
