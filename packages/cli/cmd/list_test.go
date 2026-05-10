package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kkaddal-bc/maestro/packages/cli/internal/manifest"
	"github.com/kkaddal-bc/maestro/packages/cli/internal/targets"
)

func TestListCommandShowsInstalledStatusPerActiveTarget(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	if err := os.MkdirAll(filepath.Join(home, ".claude"), 0o755); err != nil {
		t.Fatalf("mkdir .claude: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(home, ".maestro", "skills", "maestro-snap"), 0o755); err != nil {
		t.Fatalf("mkdir maestro skill: %v", err)
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
					{Name: "maestro-snap", Description: strings.Repeat("Capture all the things. ", 4)},
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
		"STATUS",
		"───",
		"maestro-snap",
		"other-skill",
		"installed",
		"not installed",
		"…",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "\x1b[") {
		t.Fatalf("output unexpectedly contains ANSI escapes:\n%s", out)
	}
	if strings.Contains(out, "\t") {
		t.Fatalf("output unexpectedly contains tabs:\n%s", out)
	}
}

func TestListCommandSucceedsWhenNothingIsInstalled(t *testing.T) {
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

func TestBuildListRowsMarksInstalledIfSkillExistsOnAnyTarget(t *testing.T) {
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, ".claude", "skills", "maestro-snap"), 0o755); err != nil {
		t.Fatalf("mkdir skill: %v", err)
	}

	rows := buildListRows(&manifest.Manifest{
		Skills: []manifest.SkillEntry{
			{Name: "maestro-snap", Description: "Capture"},
			{Name: "other-skill", Description: "Other"},
		},
	}, []targets.Target{
		{Path: filepath.Join(home, ".missing", "skills"), Required: true},
		{Path: filepath.Join(home, ".claude", "skills"), Required: false},
	})

	if len(rows) != 2 {
		t.Fatalf("rows len = %d, want 2", len(rows))
	}
	if rows[0].Status != "installed" {
		t.Fatalf("maestro-snap status = %q, want installed", rows[0].Status)
	}
	if rows[1].Status != "not installed" {
		t.Fatalf("other-skill status = %q, want not installed", rows[1].Status)
	}
}
