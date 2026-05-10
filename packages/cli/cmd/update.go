package cmd

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/kkaddal-bc/maestro/packages/cli/internal/fetcher"
	"github.com/kkaddal-bc/maestro/packages/cli/internal/installer"
	"github.com/kkaddal-bc/maestro/packages/cli/internal/manifest"
	"github.com/kkaddal-bc/maestro/packages/cli/internal/style"
	"github.com/kkaddal-bc/maestro/packages/cli/internal/targets"
	"github.com/spf13/cobra"
)

const updateSkillFlagName = "skill"

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
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update installed skills to latest versions",
		Args:  cobra.NoArgs,
		RunE:  runUpdate,
	}
	addUpdateSkillFlag(cmd)
	cmd.AddCommand(newUpdateSkillsCommand())

	return cmd
}

func newUpdateSkillsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "skills",
		Short: "Update maestro skills",
		Args:  cobra.NoArgs,
		RunE:  runUpdate,
	}
	addUpdateSkillFlag(cmd)
	return cmd
}

func runUpdate(cmd *cobra.Command, _ []string) error {
	client := newUpdateSkillsFetcher()

	skillsManifest, err := client.FetchManifest()
	if err != nil {
		return err
	}

	activeTargets := detectInstallTargets()
	selected, err := selectUpdateSkills(cmd, skillsManifest, activeTargets)
	if err != nil {
		return err
	}
	if len(selected) == 0 {
		printUpToDateSummary(cmd.OutOrStdout())
		return nil
	}

	archive, err := client.FetchSkillsArchive(skillsManifest.Version)
	if err != nil {
		return err
	}
	defer archive.Close()

	gz, err := gzip.NewReader(archive)
	if err != nil {
		return fmt.Errorf("open skills archive: %w", err)
	}
	defer gz.Close()

	result, err := installer.Update(selected, gz, activeTargets)
	if err != nil {
		return err
	}
	if result.UpToDate {
		printUpToDateSummary(cmd.OutOrStdout())
		return nil
	}

	printUpdateSummary(cmd.OutOrStdout(), result)
	return nil
}

func selectUpdateSkills(cmd *cobra.Command, manifestData *manifest.Manifest, installTargets []targets.Target) ([]string, error) {
	requested, err := requestedUpdateSkill(cmd)
	if err != nil {
		return nil, err
	}
	if requested == "" {
		return installedSkillsForTargets(manifestData, installTargets), nil
	}

	if !manifestHasSkill(manifestData, requested) {
		return nil, fmt.Errorf("unknown skill %q", requested)
	}
	if !skillInstalledOnTargets(installTargets, requested) {
		return nil, fmt.Errorf("skill %q is not installed on any active target", requested)
	}
	return []string{requested}, nil
}

func requestedUpdateSkill(cmd *cobra.Command) (string, error) {
	requested, err := cmd.Flags().GetString(updateSkillFlagName)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(requested), nil
}

func installedSkillsForTargets(manifestData *manifest.Manifest, installTargets []targets.Target) []string {
	skills := make([]string, 0, len(manifestData.Skills))
	for _, skill := range manifestData.Skills {
		if skillInstalledOnTargets(installTargets, skill.Name) {
			skills = append(skills, skill.Name)
		}
	}
	return skills
}

func skillInstalledOnTargets(activeTargets []targets.Target, skillName string) bool {
	for _, target := range activeTargets {
		if skillInstalled(target.Path, skillName) {
			return true
		}
	}
	return false
}

func skillInstalled(targetRoot, skillName string) bool {
	info, err := os.Stat(filepath.Join(targetRoot, skillName))
	return err == nil && info.IsDir()
}

func printUpdateSummary(out io.Writer, result installer.UpdateResult) {
	for _, item := range result.Updated {
		fmt.Fprintln(out, style.Success.Render(fmt.Sprintf("✓ Updated %s", item.Skill)))
	}
}

func printUpToDateSummary(out io.Writer) {
	fmt.Fprintln(out, style.Secondary.Render("- all skills are up to date"))
}

func addUpdateSkillFlag(cmd *cobra.Command) {
	cmd.Flags().String(updateSkillFlagName, "", "Update a single skill by name")
}
