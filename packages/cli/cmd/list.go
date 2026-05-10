package cmd

import (
	"io"
	"os"
	"path/filepath"

	"github.com/kkaddal-bc/maestro/packages/cli/internal/listtable"
	"github.com/kkaddal-bc/maestro/packages/cli/internal/manifest"
	"github.com/kkaddal-bc/maestro/packages/cli/internal/targets"
	"github.com/spf13/cobra"
)

func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List skills and their installation status",
		Args:  cobra.NoArgs,
		RunE:  runList,
	}
	cmd.AddCommand(newListSkillsCommand())
	return cmd
}

func newListSkillsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "skills",
		Short: "List maestro skills",
		Args:  cobra.NoArgs,
		RunE:  runList,
	}

	return cmd
}

func runList(cmd *cobra.Command, _ []string) error {
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

	return printListSkills(cmd.OutOrStdout(), home, manifestData, installTargets)
}

func printListSkills(out io.Writer, home string, manifestData *manifest.Manifest, installTargets []targets.Target) error {
	headers := []string{"SKILL", "DESCRIPTION"}
	for _, target := range installTargets {
		headers = append(headers, displayTargetPath(home, target.Path))
	}

	return listtable.NewRenderer().Render(out, headers, buildListRows(manifestData, installTargets))
}

func buildListRows(manifestData *manifest.Manifest, installTargets []targets.Target) []listtable.Row {
	rows := make([]listtable.Row, 0, len(manifestData.Skills))
	for _, skill := range manifestData.Skills {
		statuses := make([]string, 0, len(installTargets))
		for _, target := range installTargets {
			statuses = append(statuses, skillStatus(target.Path, skill.Name))
		}

		rows = append(rows, listtable.Row{
			Name:        skill.Name,
			Description: skill.Description,
			Statuses:    statuses,
		})
	}

	return rows
}

func skillStatus(targetPath, skillName string) string {
	info, err := os.Stat(filepath.Join(targetPath, skillName))
	if err != nil || !info.IsDir() {
		return "not installed"
	}
	return "installed"
}
