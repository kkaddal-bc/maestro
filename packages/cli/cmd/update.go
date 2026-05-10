package cmd

import "github.com/spf13/cobra"

func newUpdateCommand() *cobra.Command {
	cmd := newNotImplementedCommand("update", "Update related commands")

	cmd.AddCommand(newUpdateSkillsCommand())

	return cmd
}

func newUpdateSkillsCommand() *cobra.Command {
	return newNotImplementedCommand("skills", "Update maestro skills")
}
