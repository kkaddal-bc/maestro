package cmd

import "github.com/spf13/cobra"

func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List related commands",
		Run: func(cmd *cobra.Command, args []string) {
			printNotImplemented(cmd)
		},
	}

	cmd.AddCommand(newListSkillsCommand())

	return cmd
}

func newListSkillsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "skills",
		Short: "List maestro skills",
		Run: func(cmd *cobra.Command, args []string) {
			printNotImplemented(cmd)
		},
	}
}
