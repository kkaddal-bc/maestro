package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewRootCommand(version string) *cobra.Command {
	root := newNotImplementedCommand("maestro", "Maestro CLI")
	root.Version = version
	root.AddCommand(newInstallCommand(), newListCommand(), newUpdateCommand())

	return root
}

func printNotImplemented(cmd *cobra.Command) {
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "not implemented")
}
