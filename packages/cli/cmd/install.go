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
	"github.com/kkaddal-bc/maestro/packages/cli/internal/installpicker"
	"github.com/kkaddal-bc/maestro/packages/cli/internal/manifest"
	"github.com/kkaddal-bc/maestro/packages/cli/internal/style"
	"github.com/kkaddal-bc/maestro/packages/cli/internal/targets"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

const installSkillFlagName = "skill"

type skillsFetcher interface {
	FetchManifest() (*manifest.Manifest, error)
	FetchSkillsArchive(version string) (io.ReadCloser, error)
}

type skillPicker interface {
	Pick([]string) ([]string, error)
}

var (
	newSkillsFetcher = func() skillsFetcher {
		return fetcher.New()
	}
	detectInstallTargets = targets.Detect
	runInstaller         = installer.Install
	newInstallPicker     = func(in io.Reader, out io.Writer) skillPicker {
		return installpicker.New(installpicker.NewHuhSelector(in, out))
	}
	isTerminal = func(r io.Reader) bool {
		file, ok := r.(*os.File)
		if !ok {
			return false
		}
		return term.IsTerminal(int(file.Fd()))
	}
)

func newInstallCommand() *cobra.Command {
	cmd := newInstallCommandBase("install", "Install skills into configured targets", false)
	cmd.AddCommand(newInstallSkillsCommand())

	return cmd
}

func newInstallSkillsCommand() *cobra.Command {
	return newInstallCommandBase("skills", "Install maestro skills", true)
}

func newInstallCommandBase(use, short string, hidden bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		Args:  cobra.NoArgs,
		RunE:  runInstallSkills,
	}
	cmd.Hidden = hidden
	addInstallSkillFlag(cmd)

	return cmd
}

func addInstallSkillFlag(cmd *cobra.Command) {
	cmd.Flags().String(installSkillFlagName, "", "Install a single skill by name")
}

func runInstallSkills(cmd *cobra.Command, _ []string) error {
	client := newSkillsFetcher()

	manifestData, err := client.FetchManifest()
	if err != nil {
		return err
	}

	installTargets := detectInstallTargets()
	requested, err := selectSkillsToInstall(cmd, manifestData, installTargets)
	if err != nil {
		return err
	}
	if len(requested) == 0 {
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

func selectSkillsToInstall(cmd *cobra.Command, manifestData *manifest.Manifest, installTargets []targets.Target) ([]string, error) {
	out := cmd.OutOrStdout()

	skillName, err := cmd.Flags().GetString(installSkillFlagName)
	if err != nil {
		return nil, err
	}
	if skillName != "" {
		if !manifestHasSkill(manifestData, skillName) {
			return nil, fmt.Errorf("unknown skill %q", skillName)
		}
		return excludeInstalledSkills(out, []string{skillName}, installTargets), nil
	}

	skills := manifestSkillNames(manifestData)
	if len(skills) == 0 {
		return nil, errors.New("manifest contains no skills")
	}
	if isTerminal(cmd.InOrStdin()) {
		picker := newInstallPicker(cmd.InOrStdin(), out)
		selected, err := picker.Pick(skills)
		if err != nil {
			return nil, err
		}
		return excludeInstalledSkills(out, selected, installTargets), nil
	}

	return excludeInstalledSkills(out, skills, installTargets), nil
}

func manifestSkillNames(manifestData *manifest.Manifest) []string {
	skills := make([]string, 0, len(manifestData.Skills))
	for _, skill := range manifestData.Skills {
		skills = append(skills, skill.Name)
	}
	return skills
}

func manifestHasSkill(manifestData *manifest.Manifest, name string) bool {
	for _, skill := range manifestData.Skills {
		if skill.Name == name {
			return true
		}
	}
	return false
}

func excludeInstalledSkills(out io.Writer, requested []string, installTargets []targets.Target) []string {
	selected := make([]string, 0, len(requested))
	for _, skill := range requested {
		if skillInstalledOnTargets(installTargets, skill) {
			fmt.Fprintf(out, "- skipped %s (already installed)\n", skill)
			continue
		}
		selected = append(selected, skill)
	}
	return selected
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
				fmt.Fprintf(out, "%s %s → %s\n",
					style.Accent.Render("✓ installed"),
					skill,
					displayTargetPath(home, target.Path),
				)
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
