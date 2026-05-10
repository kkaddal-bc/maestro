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

func TestRootHelpShowsVerbFirstSubcommands(t *testing.T) {
	out := executeCommand(t, []string{"--help"})
	for _, want := range []string{"install", "list", "update"} {
		if !strings.Contains(out, want) {
			t.Fatalf("help output missing %q:\n%s", want, out)
		}
	}
}

func TestSkillsHelpShowsExpectedUsage(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "install skills",
			args: []string{"install", "skills", "--help"},
			want: "skills [skill-name]",
		},
		{
			name: "list skills",
			args: []string{"list", "skills", "--help"},
			want: "maestro list skills",
		},
		{
			name: "update skills",
			args: []string{"update", "skills", "--help"},
			want: "maestro update skills",
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

func TestCommandsWithoutArgsPrintNotImplemented(t *testing.T) {
	tests := [][]string{
		{},
		{"install"},
		{"list"},
	}

	for _, args := range tests {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			if got := strings.TrimSpace(executeCommand(t, args)); got != "not implemented" {
				t.Fatalf("unexpected output %q", got)
			}
		})
	}
}
