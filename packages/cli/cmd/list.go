package cmd

import "github.com/spf13/cobra"

func newListCommand() *cobra.Command {
	cmd := newNotImplementedCommand("list", "List related commands")

	cmd.AddCommand(newListSkillsCommand())

	return cmd
}

func newListSkillsCommand() *cobra.Command {
	return newNotImplementedCommand("skills", "List maestro skills")
}
