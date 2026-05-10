package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewRootCommand(version string) *cobra.Command {
	root := &cobra.Command{
		Use:     "maestro",
		Short:   "Maestro CLI",
		Version: version,
		Run: func(cmd *cobra.Command, args []string) {
			printNotImplemented(cmd)
		},
	}

	root.AddCommand(newInstallCommand())
	root.AddCommand(newListCommand())
	root.AddCommand(newUpdateCommand())

	return root
}

func printNotImplemented(cmd *cobra.Command) {
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "not implemented")
}
