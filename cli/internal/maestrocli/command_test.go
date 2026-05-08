package maestrocli_test

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/kkaddal-bc/maestro/cli/internal/maestrocli"
)

func runSnapshotCommand(t *testing.T, args ...string) (maestrocli.SnapshotConfig, string) {
	t.Helper()

	cwd := t.TempDir()
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(originalWD)
	})
	if err := os.Chdir(cwd); err != nil {
		t.Fatal(err)
	}

	var got maestrocli.SnapshotConfig
	cmd := maestrocli.NewRootCmd(func(ctx context.Context, cfg maestrocli.SnapshotConfig) error {
		got = cfg
		return nil
	})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs(append([]string{"snapshot"}, args...))

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute snapshot: %v", err)
	}

	return got, cwd
}

func TestSnapshotCommandDefaults(t *testing.T) {
	got, absCwd := runSnapshotCommand(t)

	if got.Agent != maestrocli.DefaultAgent {
		t.Fatalf("default agent = %q, want %q", got.Agent, maestrocli.DefaultAgent)
	}
	if got.Path != absCwd {
		t.Fatalf("default path = %q, want %q", got.Path, absCwd)
	}
}

func TestSnapshotCommandExplicitFlags(t *testing.T) {
	got, absCwd := runSnapshotCommand(t, "--agent", "claude", "--path", ".")

	if got.Agent != "claude" {
		t.Fatalf("explicit agent = %q, want %q", got.Agent, "claude")
	}
	if got.Path != absCwd {
		t.Fatalf("explicit path = %q, want %q", got.Path, absCwd)
	}
}
