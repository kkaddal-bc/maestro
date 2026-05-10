package cmd

import "github.com/spf13/cobra"

func NewRootCommand(version string) *cobra.Command {
	root := &cobra.Command{
		Use:     "maestro",
		Short:   "Maestro CLI",
		Version: version,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	root.AddCommand(newInstallCommand(), newListCommand(), newUpdateCommand())

	return root
}
