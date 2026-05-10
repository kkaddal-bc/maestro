package cmd

import "github.com/spf13/cobra"

func newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update related commands",
		Run: func(cmd *cobra.Command, args []string) {
			printNotImplemented(cmd)
		},
	}

	cmd.AddCommand(newUpdateSkillsCommand())

	return cmd
}

func newUpdateSkillsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "skills",
		Short: "Update maestro skills",
		Run: func(cmd *cobra.Command, args []string) {
			printNotImplemented(cmd)
		},
	}
}
