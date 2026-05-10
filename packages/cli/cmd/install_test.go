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

	"github.com/kkaddal-bc/maestro/packages/cli/internal/installer"
	"github.com/kkaddal-bc/maestro/packages/cli/internal/manifest"
	"github.com/kkaddal-bc/maestro/packages/cli/internal/targets"
)

type fakeSkillsFetcher struct {
	manifest *manifest.Manifest
	archive  []byte
}

func (f fakeSkillsFetcher) FetchManifest() (*manifest.Manifest, error) {
	return f.manifest, nil
}

func (f fakeSkillsFetcher) FetchSkillsArchive(version string) (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(f.archive)), nil
}

func TestInstallCommandInstallsSelectedSkillAndPrintsSummary(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := os.MkdirAll(filepath.Join(home, ".claude"), 0o755); err != nil {
		t.Fatalf("mkdir .claude: %v", err)
	}

	oldFetcher := newSkillsFetcher
	oldTargets := detectInstallTargets
	oldInstaller := runInstaller
	t.Cleanup(func() {
		newSkillsFetcher = oldFetcher
		detectInstallTargets = oldTargets
		runInstaller = oldInstaller
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
			archive: gzipArchive(t),
		}
	}

	var gotSkills []string
	runInstaller = func(skillNames []string, archive io.Reader, installTargets []targets.Target) (installer.Result, error) {
		gotSkills = append([]string(nil), skillNames...)
		if _, err := io.ReadAll(archive); err != nil {
			t.Fatalf("ReadAll archive: %v", err)
		}

		result := installer.Result{}
		for _, target := range installTargets {
			result.Installed = append(result.Installed, installer.Installation{
				Skill:  skillNames[0],
				Target: target.Path,
			})
		}
		return result, nil
	}

	cmd := newInstallCommand()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"maestro-snap"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if len(gotSkills) != 1 || gotSkills[0] != "maestro-snap" {
		t.Fatalf("selected skills = %#v", gotSkills)
	}

	out := stdout.String()
	for _, want := range []string{
		"✓ installed maestro-snap → ~/.maestro/skills/",
		"✓ installed maestro-snap → ~/.claude/skills/",
		"- skipped ~/.agents/skills/ (not found)",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q:\n%s", want, out)
		}
	}
}

func TestInstallCommandInstallsAllSkillsByDefault(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := os.MkdirAll(filepath.Join(home, ".claude"), 0o755); err != nil {
		t.Fatalf("mkdir .claude: %v", err)
	}

	oldFetcher := newSkillsFetcher
	oldTargets := detectInstallTargets
	oldInstaller := runInstaller
	t.Cleanup(func() {
		newSkillsFetcher = oldFetcher
		detectInstallTargets = oldTargets
		runInstaller = oldInstaller
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
			archive: gzipArchive(t),
		}
	}

	var gotSkills []string
	runInstaller = func(skillNames []string, archive io.Reader, installTargets []targets.Target) (installer.Result, error) {
		gotSkills = append([]string(nil), skillNames...)
		if _, err := io.ReadAll(archive); err != nil {
			t.Fatalf("ReadAll archive: %v", err)
		}
		return installer.Result{}, nil
	}

	cmd := newInstallCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if len(gotSkills) != 2 || gotSkills[0] != "maestro-snap" || gotSkills[1] != "other-skill" {
		t.Fatalf("selected skills = %#v", gotSkills)
	}
}

func TestInstallCommandRejectsUnknownSkill(t *testing.T) {
	oldFetcher := newSkillsFetcher
	t.Cleanup(func() {
		newSkillsFetcher = oldFetcher
	})

	newSkillsFetcher = func() skillsFetcher {
		return fakeSkillsFetcher{
			manifest: &manifest.Manifest{
				Version: "v1.2.3",
				Skills:  []manifest.SkillEntry{{Name: "maestro-snap", Description: "Capture"}},
			},
		}
	}

	cmd := newInstallCommand()
	cmd.SetArgs([]string{"unknown"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	if err := cmd.Execute(); err == nil {
		t.Fatal("Execute() error = nil, want error")
	}
}

func gzipArchive(t *testing.T) []byte {
	t.Helper()

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	if err := tw.Close(); err != nil {
		t.Fatalf("tar close: %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("gzip close: %v", err)
	}
	return buf.Bytes()
}
