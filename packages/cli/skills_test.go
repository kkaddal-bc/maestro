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

func TestMaestroSnapSkillDocsStayLanguageAgnostic(t *testing.T) {
	skill := readSkillFile(t, "SKILL.md")
	templates := readSkillFile(t, "OUTPUT-TEMPLATES.md")
	patterns := readSkillFile(t, "SCAN-PATTERNS.md")

	for _, want := range []string{
		"manifest and lockfiles first",
		"LLM-native code reading",
		"proto files exist",
		"OpenAPI or Swagger files exist",
		"migration directories by name heuristics",
	} {
		if !strings.Contains(skill, want) {
			t.Fatalf("SKILL.md missing %q", want)
		}
	}

	if !strings.Contains(templates, "🔒 (inferred)") {
		t.Fatalf("OUTPUT-TEMPLATES.md missing inferred auth marker")
	}

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
			if strings.Contains(content.text, forbidden) {
				t.Fatalf("%s contains forbidden substring %q", content.name, forbidden)
			}
		}
	}
}
