package cmd

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kkaddal-bc/maestro/packages/cli/internal/manifest"
)

type countingUpdateFetcher struct {
	manifest      *manifest.Manifest
	archive       []byte
	manifestCalls int
	archiveCalls  int
}

func stubUpdateSkillsFetcher(t *testing.T, fetcher updateSkillsFetcher) {
	t.Helper()

	oldFetcher := newUpdateSkillsFetcher
	t.Cleanup(func() {
		newUpdateSkillsFetcher = oldFetcher
	})
	newUpdateSkillsFetcher = func() updateSkillsFetcher {
		return fetcher
	}
}

func (f *countingUpdateFetcher) FetchManifest() (*manifest.Manifest, error) {
	f.manifestCalls++
	return f.manifest, nil
}

func (f *countingUpdateFetcher) FetchSkillsArchive(version string) (io.ReadCloser, error) {
	f.archiveCalls++
	return io.NopCloser(bytes.NewReader(f.archive)), nil
}

func TestUpdateCommandUpdatesInstalledSkillsAndFetchesFreshManifestEachRun(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := os.MkdirAll(filepath.Join(home, ".claude"), 0o755); err != nil {
		t.Fatalf("mkdir .claude: %v", err)
	}

	maestroTarget := filepath.Join(home, ".maestro", "skills")
	claudeTarget := filepath.Join(home, ".claude", "skills")
	mustWriteUpdateFile(t, filepath.Join(maestroTarget, "maestro-snap", "SKILL.md"), "old snap")
	mustWriteUpdateFile(t, filepath.Join(maestroTarget, "maestro-snap", "stale.txt"), "stale")
	mustWriteUpdateFile(t, filepath.Join(claudeTarget, "maestro-snap", "SKILL.md"), "old snap")

	fetcher := &countingUpdateFetcher{
		manifest: &manifest.Manifest{
			Version: "v1.2.3",
			Skills: []manifest.SkillEntry{
				{Name: "maestro-snap", Description: "Capture"},
				{Name: "other-skill", Description: "Other"},
			},
		},
		archive: gzipUpdateArchive(t, map[string]string{
			"maestro-snap/SKILL.md": "new snap",
			"other-skill/SKILL.md":  "new other",
		}),
	}

	stubUpdateSkillsFetcher(t, fetcher)

	firstOutput := executeUpdateCommand(t)
	for _, want := range []string{
		"✓ updated maestro-snap → ~/.maestro/skills/",
		"✓ updated maestro-snap → ~/.claude/skills/",
	} {
		if !strings.Contains(firstOutput, want) {
			t.Fatalf("first output missing %q:\n%s", want, firstOutput)
		}
	}
	if strings.Contains(firstOutput, "up to date") {
		t.Fatalf("first output reported up to date unexpectedly:\n%s", firstOutput)
	}

	assertUpdateFileContents(t, filepath.Join(maestroTarget, "maestro-snap", "SKILL.md"), "new snap")
	assertUpdateFileContents(t, filepath.Join(claudeTarget, "maestro-snap", "SKILL.md"), "new snap")
	if _, err := os.Stat(filepath.Join(maestroTarget, "maestro-snap", "stale.txt")); !os.IsNotExist(err) {
		t.Fatalf("stale file remains unexpectedly: %v", err)
	}

	secondOutput := executeUpdateCommand(t)
	if !strings.Contains(secondOutput, "skills are already up to date") {
		t.Fatalf("second output missing up-to-date message:\n%s", secondOutput)
	}

	if fetcher.manifestCalls != 2 {
		t.Fatalf("manifestCalls = %d, want 2", fetcher.manifestCalls)
	}
	if fetcher.archiveCalls != 2 {
		t.Fatalf("archiveCalls = %d, want 2", fetcher.archiveCalls)
	}
}

func TestUpdateCommandWithSkillFlagUpdatesOnlyRequestedSkill(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := os.MkdirAll(filepath.Join(home, ".claude"), 0o755); err != nil {
		t.Fatalf("mkdir .claude: %v", err)
	}

	maestroTarget := filepath.Join(home, ".maestro", "skills")
	claudeTarget := filepath.Join(home, ".claude", "skills")
	mustWriteUpdateFile(t, filepath.Join(maestroTarget, "maestro-snap", "SKILL.md"), "old snap")
	mustWriteUpdateFile(t, filepath.Join(maestroTarget, "other-skill", "SKILL.md"), "old other")
	mustWriteUpdateFile(t, filepath.Join(claudeTarget, "maestro-snap", "SKILL.md"), "old snap")
	mustWriteUpdateFile(t, filepath.Join(claudeTarget, "other-skill", "SKILL.md"), "old other")

	fetcher := &countingUpdateFetcher{
		manifest: &manifest.Manifest{
			Version: "v1.2.3",
			Skills: []manifest.SkillEntry{
				{Name: "maestro-snap", Description: "Capture"},
				{Name: "other-skill", Description: "Other"},
			},
		},
		archive: gzipUpdateArchive(t, map[string]string{
			"maestro-snap/SKILL.md": "new snap",
			"other-skill/SKILL.md":  "new other",
		}),
	}

	stubUpdateSkillsFetcher(t, fetcher)

	cmd := newUpdateCommand()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--skill", "maestro-snap"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	assertUpdateFileContents(t, filepath.Join(maestroTarget, "maestro-snap", "SKILL.md"), "new snap")
	assertUpdateFileContents(t, filepath.Join(claudeTarget, "maestro-snap", "SKILL.md"), "new snap")
	assertUpdateFileContents(t, filepath.Join(maestroTarget, "other-skill", "SKILL.md"), "old other")
	assertUpdateFileContents(t, filepath.Join(claudeTarget, "other-skill", "SKILL.md"), "old other")
	if strings.Contains(stdout.String(), "other-skill") {
		t.Fatalf("output unexpectedly mentions other-skill:\n%s", stdout.String())
	}
}

func TestUpdateCommandRejectsUnknownOrUninstalledSkill(t *testing.T) {
	tests := []struct {
		name            string
		skill           string
		wantErrContains string
	}{
		{
			name:            "unknown skill",
			skill:           "unknown",
			wantErrContains: `unknown skill "unknown"`,
		},
		{
			name:            "uninstalled skill",
			skill:           "maestro-snap",
			wantErrContains: `skill "maestro-snap" is not installed on any active target`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)

			stubUpdateSkillsFetcher(t, &countingUpdateFetcher{
				manifest: &manifest.Manifest{
					Version: "v1.2.3",
					Skills: []manifest.SkillEntry{
						{Name: "maestro-snap", Description: "Capture"},
					},
				},
			})

			cmd := newUpdateCommand()
			cmd.SetOut(io.Discard)
			cmd.SetErr(io.Discard)
			cmd.SetArgs([]string{"--skill", tt.skill})

			err := cmd.Execute()
			if err == nil {
				t.Fatal("Execute() error = nil, want error")
			}
			if !strings.Contains(err.Error(), tt.wantErrContains) {
				t.Fatalf("Execute() error = %v, want substring %q", err, tt.wantErrContains)
			}
		})
	}
}

func executeUpdateCommand(t *testing.T) string {
	t.Helper()

	cmd := newUpdateCommand()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	return stdout.String()
}

func mustWriteUpdateFile(t *testing.T, path, contents string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%s): %v", path, err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("WriteFile(%s): %v", path, err)
	}
}

func gzipUpdateArchive(t *testing.T, files map[string]string) []byte {
	t.Helper()

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for name, contents := range files {
		if err := tw.WriteHeader(&tar.Header{
			Name:     name,
			Mode:     0o644,
			Size:     int64(len(contents)),
			Typeflag: tar.TypeReg,
		}); err != nil {
			t.Fatalf("WriteHeader(%q): %v", name, err)
		}
		if _, err := io.WriteString(tw, contents); err != nil {
			t.Fatalf("WriteString(%q): %v", name, err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("tar close: %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("gzip close: %v", err)
	}
	return buf.Bytes()
}

func assertUpdateFileContents(t *testing.T, path, want string) {
	t.Helper()

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", path, err)
	}
	if string(got) != want {
		t.Fatalf("%s = %q, want %q", path, string(got), want)
	}
}
