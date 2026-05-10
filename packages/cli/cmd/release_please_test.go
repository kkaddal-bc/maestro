package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestReleasePleaseWorkflowAndConfigFilesMatchIssue33(t *testing.T) {
	root := repoRoot(t)

	workflow := mustReadFile(t, filepath.Join(root, ".github", "workflows", "release-please.yml"))
	for _, want := range []string{
		"name: release-please",
		"branches:\n      - main",
		"contents: write",
		"pull-requests: write",
		"googleapis/release-please-action@v4",
		"config-file: release-please-config.json",
		"manifest-file: .release-please-manifest.json",
	} {
		if !strings.Contains(workflow, want) {
			t.Fatalf("workflow missing %q:\n%s", want, workflow)
		}
	}

	config := mustReadFile(t, filepath.Join(root, "release-please-config.json"))
	var releasePleaseConfig struct {
		Packages map[string]struct {
			ReleaseType  string `json:"release-type"`
			PackageName  string `json:"package-name"`
			ChangelogPath string `json:"changelog-path"`
		} `json:"packages"`
	}
	if err := json.Unmarshal([]byte(config), &releasePleaseConfig); err != nil {
		t.Fatalf("unmarshal release-please-config.json: %v", err)
	}

	pkg, ok := releasePleaseConfig.Packages["packages/cli"]
	if !ok {
		t.Fatal("release-please config missing packages/cli entry")
	}
	if pkg.ReleaseType != "simple" {
		t.Fatalf("release type = %q, want simple", pkg.ReleaseType)
	}
	if pkg.PackageName != "maestro" {
		t.Fatalf("package name = %q, want maestro", pkg.PackageName)
	}
	if pkg.ChangelogPath != "CHANGELOG.md" {
		t.Fatalf("changelog path = %q, want CHANGELOG.md", pkg.ChangelogPath)
	}

	manifest := mustReadFile(t, filepath.Join(root, ".release-please-manifest.json"))
	var releasePleaseManifest map[string]string
	if err := json.Unmarshal([]byte(manifest), &releasePleaseManifest); err != nil {
		t.Fatalf("unmarshal .release-please-manifest.json: %v", err)
	}
	if got := releasePleaseManifest["packages/cli"]; got != "0.1.0" {
		t.Fatalf("manifest version = %q, want 0.1.0", got)
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
}

func mustReadFile(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", path, err)
	}
	return string(data)
}
