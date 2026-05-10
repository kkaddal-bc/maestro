package releaseassets

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMaestroSnapDocsReferenceDependenciesMd(t *testing.T) {
	assertMaestroSnapDocs(t, []docCheck{
		{
			name:    "skill outputs",
			path:    filepath.Join("skills", "maestro-snap", "SKILL.md"),
			want:    ".maestro/maestro-interface/dependencies.md",
		},
		{
			name:    "skill scope",
			path:    filepath.Join("skills", "maestro-snap", "SKILL.md"),
			want:    "`dependencies.md` is v1 documentation for consumed gRPC services only.",
		},
		{
			name:    "template path",
			path:    filepath.Join("skills", "maestro-snap", "OUTPUT-TEMPLATES.md"),
			want:    "## `.maestro/maestro-interface/dependencies.md`",
		},
		{
			name:    "template scope",
			path:    filepath.Join("skills", "maestro-snap", "OUTPUT-TEMPLATES.md"),
			want:    "`dependencies.md` captures consumed gRPC services only in v1.",
		},
		{
			name:    "agents path",
			path:    "AGENTS.md",
			want:    ".maestro/maestro-interface/dependencies.md",
		},
	})
}

func TestMaestroSnapDocsReferenceMaestroJson(t *testing.T) {
	assertMaestroSnapDocs(t, []docCheck{
		{
			name:    "skill outputs",
			path:    filepath.Join("skills", "maestro-snap", "SKILL.md"),
			want:    ".maestro/maestro-interface/maestro.json",
		},
		{
			name:    "skill schema",
			path:    filepath.Join("skills", "maestro-snap", "SKILL.md"),
			want:    "`maestro.json` is the machine-readable snapshot for registry consumption.",
		},
		{
			name:    "template path",
			path:    filepath.Join("skills", "maestro-snap", "OUTPUT-TEMPLATES.md"),
			want:    "## `.maestro/maestro-interface/maestro.json`",
		},
		{
			name:    "template schema",
			path:    filepath.Join("skills", "maestro-snap", "OUTPUT-TEMPLATES.md"),
			want:    "\"schema_version\": \"1\"",
		},
		{
			name:    "agents path",
			path:    "AGENTS.md",
			want:    ".maestro/maestro-interface/maestro.json",
		},
	})
}

type docCheck struct {
	name string
	path string
	want string
}

func assertMaestroSnapDocs(t *testing.T, checks []docCheck) {
	t.Helper()

	root := filepath.Join("..", "..", "..", "..")

	for _, check := range checks {
		content := mustReadFile(t, filepath.Join(root, check.path))
		if !strings.Contains(content, check.want) {
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
