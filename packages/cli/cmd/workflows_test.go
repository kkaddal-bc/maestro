package cmd

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func repoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
}

func TestWorkflowFilesUseRenamedNames(t *testing.T) {
	root := repoRoot(t)

	tests := []struct {
		path string
		want []string
		deny []string
	}{
		{
			path: ".github/workflows/release.yml",
			want: []string{
				"name: release",
				"googleapis/release-please-action@v4",
				"branches:",
				"- main",
			},
			deny: []string{
				"softprops/action-gh-release@v2",
			},
		},
		{
			path: ".github/workflows/brew-publish.yml",
			want: []string{
				"name: brew-publish",
				"softprops/action-gh-release@v2",
				"tags:",
				"- \"v*\"",
			},
			deny: []string{
				"googleapis/release-please-action@v4",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			content := readTestFile(t, filepath.Join(root, tt.path))
			assertContainsAll(t, tt.path, content, tt.want...)
			assertExcludesAll(t, tt.path, content, tt.deny...)
		})
	}
}

func TestOldWorkflowFilesAreRemoved(t *testing.T) {
	root := repoRoot(t)

	for _, path := range []string{
		".github/workflows/release-please.yml",
	} {
		assertFileRemoved(t, filepath.Join(root, path))
	}
}

func assertContainsAll(t *testing.T, path, content string, wants ...string) {
	t.Helper()

	for _, want := range wants {
		if !strings.Contains(content, want) {
			t.Fatalf("%s missing %q:\n%s", path, want, content)
		}
	}
}

func assertExcludesAll(t *testing.T, path, content string, denies ...string) {
	t.Helper()

	for _, deny := range denies {
		if strings.Contains(content, deny) {
			t.Fatalf("%s unexpectedly contains %q:\n%s", path, deny, content)
		}
	}
}

func assertFileRemoved(t *testing.T, path string) {
	t.Helper()

	_, err := os.Stat(path)
	if err == nil || !os.IsNotExist(err) {
		t.Fatalf("expected %s to be removed, got err=%v", path, err)
	}
}

func readTestFile(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", path, err)
	}
	return string(data)
}
