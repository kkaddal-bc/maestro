package cmd

import "github.com/spf13/cobra"

func newInstallCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install related commands",
		Run: func(cmd *cobra.Command, args []string) {
			printNotImplemented(cmd)
		},
	}

	cmd.AddCommand(newInstallSkillsCommand())

	return cmd
}

func newInstallSkillsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "skills [skill-name]",
		Short: "Install maestro skills",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			printNotImplemented(cmd)
		},
	}
}
