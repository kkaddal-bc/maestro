package cmd

import "github.com/spf13/cobra"

func newInstallCommand() *cobra.Command {
	cmd := newNotImplementedCommand("install", "Install related commands")

	cmd.AddCommand(newInstallSkillsCommand())

	return cmd
}

func newInstallSkillsCommand() *cobra.Command {
	cmd := newNotImplementedCommand("skills [skill-name]", "Install maestro skills")
	cmd.Args = cobra.MaximumNArgs(1)

	return cmd
}
