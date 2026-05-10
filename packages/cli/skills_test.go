package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func readSkillFile(t *testing.T, name string) string {
	t.Helper()

	path := filepath.Join("..", "..", "skills", "maestro-snap", name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

func requireContains(t *testing.T, fileName, content, want string) {
	t.Helper()

	if !strings.Contains(content, want) {
		t.Fatalf("%s missing %q", fileName, want)
	}
}

func requireNotContains(t *testing.T, fileName, content, forbidden string) {
	t.Helper()

	if strings.Contains(content, forbidden) {
		t.Fatalf("%s contains forbidden substring %q", fileName, forbidden)
	}
}

func TestMaestroSnapSkillDocsStayLanguageAgnostic(t *testing.T) {
	skill := readSkillFile(t, "SKILL.md")
	templates := readSkillFile(t, "OUTPUT-TEMPLATES.md")
	patterns := readSkillFile(t, "SCAN-PATTERNS.md")

	for _, want := range []string{
		"manifest and lockfiles first",
		"read the code directly",
		"proto files exist",
		"OpenAPI or Swagger files exist",
		"migration directories by name heuristics",
	} {
		requireContains(t, "SKILL.md", skill, want)
	}

	requireContains(t, "OUTPUT-TEMPLATES.md", templates, "🔒 (inferred)")

	for _, content := range []struct {
		name string
		text string
	}{
		{name: "SKILL.md", text: skill},
		{name: "OUTPUT-TEMPLATES.md", text: templates},
		{name: "SCAN-PATTERNS.md", text: patterns},
	} {
		for _, forbidden := range []string{
			"--include=\"*.kt\"",
			"Kotlin",
		} {
			requireNotContains(t, content.name, content.text, forbidden)
		}
	}
}
