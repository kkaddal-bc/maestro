package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kkaddal-bc/maestro/packages/cli/internal/manifest"
)

func TestListSkillsCommandShowsInstalledStatusPerActiveTarget(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	if err := os.MkdirAll(filepath.Join(home, ".claude"), 0o755); err != nil {
		t.Fatalf("mkdir .claude: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(home, ".maestro", "skills", "maestro-snap"), 0o755); err != nil {
		t.Fatalf("mkdir maestro skill: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(home, ".claude", "skills", "other-skill"), 0o755); err != nil {
		t.Fatalf("mkdir claude skill: %v", err)
	}

	oldFetcher := newSkillsFetcher
	oldTargets := detectInstallTargets
	t.Cleanup(func() {
		newSkillsFetcher = oldFetcher
		detectInstallTargets = oldTargets
	})

	newSkillsFetcher = func() skillsFetcher {
		return fakeSkillsFetcher{
			manifest: &manifest.Manifest{
				Version: "v1.2.3",
				Skills: []manifest.SkillEntry{
					{Name: "maestro-snap", Description: "Capture"},
					{Name: "other-skill", Description: "Other"},
				},
			},
		}
	}

	cmd := newListCommand()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(io.Discard)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	out := stdout.String()
	for _, want := range []string{
		"SKILL",
		"DESCRIPTION",
		"~/.maestro/skills/",
		"~/.claude/skills/",
		"maestro-snap",
		"other-skill",
		"installed",
		"not installed",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "~/.agents/skills/") {
		t.Fatalf("output unexpectedly includes missing target:\n%s", out)
	}
}

func TestListSkillsCommandSucceedsWhenNothingIsInstalled(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	oldFetcher := newSkillsFetcher
	t.Cleanup(func() {
		newSkillsFetcher = oldFetcher
	})

	newSkillsFetcher = func() skillsFetcher {
		return fakeSkillsFetcher{
			manifest: &manifest.Manifest{
				Version: "v1.2.3",
				Skills: []manifest.SkillEntry{
					{Name: "maestro-snap", Description: "Capture"},
				},
			},
		}
	}

	cmd := newListCommand()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(io.Discard)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if got := stdout.String(); !strings.Contains(got, "not installed") {
		t.Fatalf("output missing installation state:\n%s", got)
	}
}
