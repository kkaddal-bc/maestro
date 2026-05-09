package maestrocli

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/spf13/cobra"
)

const DefaultAgent = "claude"

type SnapshotConfig struct {
	Agent string
	Path  string
}

type SnapshotHandler func(context.Context, SnapshotConfig) error

func NewRootCmd(handler SnapshotHandler) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "maestro",
		Short:         "Maestro CLI",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.AddCommand(NewSnapshotCmd(handler))
	return cmd
}

func NewSnapshotCmd(handler SnapshotHandler) *cobra.Command {
	if handler == nil {
		handler = func(context.Context, SnapshotConfig) error { return nil }
	}

	var agent string
	var path string

	cmd := &cobra.Command{
		Use:           "snapshot",
		Short:         "Create or update the maestro interface",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			resolvedPath, err := resolvePath(path)
			if err != nil {
				return err
			}

			return handler(cmd.Context(), SnapshotConfig{
				Agent: agent,
				Path:  resolvedPath,
			})
		},
	}

	cmd.Flags().StringVar(&agent, "agent", DefaultAgent, "AI provider to use")
	cmd.Flags().StringVar(&path, "path", "", "Repository root to snapshot")

	return cmd
}

func Execute(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	cmd := NewRootCmd(nil)
	cmd.SetArgs(args)
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	return cmd.ExecuteContext(ctx)
}

func resolvePath(path string) (string, error) {
	if path == "" {
		path = "."
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve snapshot path: %w", err)
	}

	return abs, nil
}
