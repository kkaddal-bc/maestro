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
			data, err := os.ReadFile(filepath.Join(root, tt.path))
			if err != nil {
				t.Fatalf("ReadFile(%s): %v", tt.path, err)
			}
			content := string(data)
			for _, want := range tt.want {
				if !strings.Contains(content, want) {
					t.Fatalf("%s missing %q:\n%s", tt.path, want, content)
				}
			}
			for _, deny := range tt.deny {
				if strings.Contains(content, deny) {
					t.Fatalf("%s unexpectedly contains %q:\n%s", tt.path, deny, content)
				}
			}
		})
	}
}

func TestOldWorkflowFilesAreRemoved(t *testing.T) {
	root := repoRoot(t)

	for _, path := range []string{
		".github/workflows/release-please.yml",
	} {
		_, err := os.Stat(filepath.Join(root, path))
		if !os.IsNotExist(err) {
			t.Fatalf("expected %s to be removed, got err=%v", path, err)
		}
	}
}
