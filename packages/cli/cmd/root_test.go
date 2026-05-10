package cmd

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func executeCommand(t *testing.T, args []string) string {
	t.Helper()

	root := NewRootCommand("dev")
	var stdout bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(io.Discard)
	root.SetArgs(args)

	if err := root.Execute(); err != nil {
		t.Fatalf("execute command: %v", err)
	}

	return stdout.String()
}

func TestRootHelpShowsTopLevelCommands(t *testing.T) {
	out := executeCommand(t, []string{"--help"})
	for _, want := range []string{"install", "list", "update"} {
		if !strings.Contains(out, want) {
			t.Fatalf("help output missing %q:\n%s", want, out)
		}
	}
}

func TestLeafHelpShowsExpectedUsage(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "install",
			args: []string{"install", "--help"},
			want: "install [skill-name]",
		},
		{
			name: "list",
			args: []string{"list", "--help"},
			want: "List maestro skills",
		},
		{
			name: "update",
			args: []string{"update", "--help"},
			want: "Update installed skills to latest versions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := executeCommand(t, tt.args)
			if !strings.Contains(out, tt.want) {
				t.Fatalf("help output missing %q:\n%s", tt.want, out)
			}
		})
	}
}

func TestRootWithoutArgsShowsHelp(t *testing.T) {
	out := executeCommand(t, nil)
	for _, want := range []string{"Usage:", "install", "list", "update"} {
		if !strings.Contains(out, want) {
			t.Fatalf("root output missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "not implemented") {
		t.Fatalf("root output still contains not implemented:\n%s", out)
	}
}
