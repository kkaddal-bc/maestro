package releaseassets

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMaestroSnapDocsReferenceDependenciesMd(t *testing.T) {
	root := filepath.Join("..", "..", "..", "..")

	skillDoc := mustReadFile(t, filepath.Join(root, "skills", "maestro-snap", "SKILL.md"))
	templateDoc := mustReadFile(t, filepath.Join(root, "skills", "maestro-snap", "OUTPUT-TEMPLATES.md"))
	agentsDoc := mustReadFile(t, filepath.Join(root, "AGENTS.md"))

	checks := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "skill outputs",
			content: skillDoc,
			want:    ".maestro/maestro-interface/dependencies.md",
		},
		{
			name:    "skill scope",
			content: skillDoc,
			want:    "`dependencies.md` is v1 documentation for consumed gRPC services only.",
		},
		{
			name:    "template path",
			content: templateDoc,
			want:    "## `.maestro/maestro-interface/dependencies.md`",
		},
		{
			name:    "template scope",
			content: templateDoc,
			want:    "`dependencies.md` captures consumed gRPC services only in v1.",
		},
		{
			name:    "agents path",
			content: agentsDoc,
			want:    ".maestro/maestro-interface/dependencies.md",
		},
	}

	for _, check := range checks {
		if !strings.Contains(check.content, check.want) {
			t.Fatalf("%s missing %q", check.name, check.want)
		}
	}
}

func mustReadFile(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", path, err)
	}
	return string(data)
}
