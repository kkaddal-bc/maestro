package installer

import (
	"archive/tar"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/kkaddal-bc/maestro/packages/cli/internal/targets"
)

func TestInstallWritesAllSkillsToAllTargets(t *testing.T) {
	archive := buildTestArchive(t)
	home := t.TempDir()
	firstTarget := filepath.Join(home, ".maestro", "skills")
	secondTarget := filepath.Join(home, ".claude", "skills")

	result, err := Install(nil, bytes.NewReader(archive), []targets.Target{
		{Path: firstTarget, Required: true},
		{Path: secondTarget, Required: false},
	})
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if len(result.Installed) != 4 {
		t.Fatalf("len(result.Installed) = %d, want 4", len(result.Installed))
	}

	assertFileContents(t, filepath.Join(firstTarget, "maestro-snap", "SKILL.md"), "snap skill")
	assertFileContents(t, filepath.Join(firstTarget, "maestro-snap", "notes", "summary.txt"), "summary")
	assertFileContents(t, filepath.Join(secondTarget, "other-skill", "SKILL.md"), "other skill")
}

func TestInstallSelectedSkillOnly(t *testing.T) {
	archive := buildTestArchive(t)
	home := t.TempDir()
	target := filepath.Join(home, ".maestro", "skills")

	result, err := Install([]string{"maestro-snap"}, bytes.NewReader(archive), []targets.Target{{Path: target, Required: true}})
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if len(result.Installed) != 1 {
		t.Fatalf("len(result.Installed) = %d, want 1", len(result.Installed))
	}

	assertFileContents(t, filepath.Join(target, "maestro-snap", "SKILL.md"), "snap skill")
	if _, err := os.Stat(filepath.Join(target, "other-skill")); !os.IsNotExist(err) {
		t.Fatalf("other-skill exists unexpectedly: %v", err)
	}
}

func TestInstallUnknownSkillErrors(t *testing.T) {
	archive := buildTestArchive(t)
	target := filepath.Join(t.TempDir(), ".maestro", "skills")

	if _, err := Install([]string{"missing"}, bytes.NewReader(archive), []targets.Target{{Path: target, Required: true}}); err == nil {
		t.Fatal("Install() error = nil, want error")
	}
}

func TestUpdateOnlyOverwritesInstalledSkills(t *testing.T) {
	archive := buildUpdatedTestArchive(t)
	home := t.TempDir()
	maestroTarget := filepath.Join(home, ".maestro", "skills")
	claudeTarget := filepath.Join(home, ".claude", "skills")
	missingTarget := filepath.Join(home, ".agents", "skills")

	mustWriteFile(t, filepath.Join(maestroTarget, "maestro-snap", "SKILL.md"), "old snap")
	mustWriteFile(t, filepath.Join(maestroTarget, "maestro-snap", "stale.txt"), "stale")
	mustWriteFile(t, filepath.Join(claudeTarget, "maestro-snap", "SKILL.md"), "old snap")
	mustWriteFile(t, filepath.Join(claudeTarget, "other-skill", "SKILL.md"), "old other")

	result, err := Update([]string{"maestro-snap"}, bytes.NewReader(archive), []targets.Target{
		{Path: maestroTarget, Required: true},
		{Path: claudeTarget, Required: false},
		{Path: missingTarget, Required: false},
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if len(result.Updated) != 2 {
		t.Fatalf("len(result.Updated) = %d, want 2", len(result.Updated))
	}
	if result.UpToDate {
		t.Fatal("Update() reported up to date, want changes")
	}

	assertFileContents(t, filepath.Join(maestroTarget, "maestro-snap", "SKILL.md"), "new snap")
	assertFileContents(t, filepath.Join(claudeTarget, "maestro-snap", "SKILL.md"), "new snap")
	assertFileContents(t, filepath.Join(claudeTarget, "other-skill", "SKILL.md"), "old other")
	if _, err := os.Stat(filepath.Join(maestroTarget, "maestro-snap", "stale.txt")); !os.IsNotExist(err) {
		t.Fatalf("stale file remains unexpectedly: %v", err)
	}
	if _, err := os.Stat(filepath.Join(missingTarget, "maestro-snap")); !os.IsNotExist(err) {
		t.Fatalf("missing target was created unexpectedly: %v", err)
	}
}

func buildTestArchive(t *testing.T) []byte {
	t.Helper()

	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	files := map[string]string{
		"maestro-snap/SKILL.md":            "snap skill",
		"maestro-snap/notes/summary.txt":   "summary",
		"other-skill/SKILL.md":             "other skill",
		"other-skill/resources/README.txt": "resource note",
		"other-skill/resources/inner/info": "inner info",
		"maestro-snap/resources/image.txt": "image",
		"maestro-snap/resources/docs.txt":  "docs",
		"other-skill/resources/extra.txt":  "extra",
	}

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
		t.Fatalf("Close(): %v", err)
	}
	return buf.Bytes()
}

func buildUpdatedTestArchive(t *testing.T) []byte {
	t.Helper()

	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	files := map[string]string{
		"maestro-snap/SKILL.md": "new snap",
		"other-skill/SKILL.md":   "new other",
	}

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
		t.Fatalf("Close(): %v", err)
	}
	return buf.Bytes()
}

func mustWriteFile(t *testing.T, path, contents string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%s): %v", path, err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("WriteFile(%s): %v", path, err)
	}
}

func assertFileContents(t *testing.T, path, want string) {
	t.Helper()

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", path, err)
	}
	if string(got) != want {
		t.Fatalf("%s = %q, want %q", path, string(got), want)
	}
}
