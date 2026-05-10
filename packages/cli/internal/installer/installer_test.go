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
