package cmd

import "github.com/spf13/cobra"

func newNotImplementedCommand(use, short string) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		Run: func(cmd *cobra.Command, _ []string) {
			printNotImplemented(cmd)
		},
	}
}
