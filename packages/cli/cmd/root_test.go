package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootHelpShowsVerbFirstSubcommands(t *testing.T) {
	root := NewRootCommand("dev")
	var stdout, stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs([]string{"--help"})

	if err := root.Execute(); err != nil {
		t.Fatalf("execute help: %v", err)
	}

	out := stdout.String()
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
			root := NewRootCommand("dev")
			var stdout, stderr bytes.Buffer
			root.SetOut(&stdout)
			root.SetErr(&stderr)
			root.SetArgs(tt.args)

			if err := root.Execute(); err != nil {
				t.Fatalf("execute help: %v", err)
			}

			out := stdout.String()
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
		{"install", "skills"},
		{"list"},
		{"list", "skills"},
		{"update"},
		{"update", "skills"},
	}

	for _, args := range tests {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			root := NewRootCommand("dev")
			var stdout, stderr bytes.Buffer
			root.SetOut(&stdout)
			root.SetErr(&stderr)
			root.SetArgs(args)

			if err := root.Execute(); err != nil {
				t.Fatalf("execute command: %v", err)
			}

			if got := strings.TrimSpace(stdout.String()); got != "not implemented" {
				t.Fatalf("unexpected output %q", got)
			}
		})
	}
}
