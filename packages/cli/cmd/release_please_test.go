package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

const (
	releasePleaseWorkflowPath = ".github/workflows/release-please.yml"
	releasePleaseConfigPath   = "release-please-config.json"
	releasePleaseManifestPath = ".release-please-manifest.json"

	releasePleasePackagePath = "packages/cli"
	releasePleaseVersion     = "0.1.0"
)

type releasePleaseConfigFile struct {
	Packages map[string]releasePleasePackageConfig `json:"packages"`
}

type releasePleasePackageConfig struct {
	ReleaseType   string `json:"release-type"`
	PackageName   string `json:"package-name"`
	ChangelogPath string `json:"changelog-path"`
}

func TestReleasePleaseWorkflowAndConfigFilesMatchIssue33(t *testing.T) {
	root := repoRoot(t)

	workflow := mustReadFile(t, filepath.Join(root, releasePleaseWorkflowPath))
	assertContainsAll(t, workflow,
		"name: release-please",
		"branches:\n      - main",
		"contents: write",
		"pull-requests: write",
		"googleapis/release-please-action@v4",
		"config-file: release-please-config.json",
		"manifest-file: .release-please-manifest.json",
	)

	config := mustReadJSON[releasePleaseConfigFile](t, filepath.Join(root, releasePleaseConfigPath))

	pkg, ok := config.Packages[releasePleasePackagePath]
	if !ok {
		t.Fatalf("release-please config missing %s entry", releasePleasePackagePath)
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

	manifest := mustReadJSON[map[string]string](t, filepath.Join(root, releasePleaseManifestPath))
	if got := manifest[releasePleasePackagePath]; got != releasePleaseVersion {
		t.Fatalf("manifest version = %q, want %s", got, releasePleaseVersion)
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

func mustReadJSON[T any](t *testing.T, path string) T {
	t.Helper()

	var decoded T
	if err := json.Unmarshal([]byte(mustReadFile(t, path)), &decoded); err != nil {
		t.Fatalf("unmarshal %s: %v", path, err)
	}
	return decoded
}

func assertContainsAll(t *testing.T, content string, wants ...string) {
	t.Helper()

	for _, want := range wants {
		if !strings.Contains(content, want) {
			t.Fatalf("content missing %q:\n%s", want, content)
		}
	}
}
