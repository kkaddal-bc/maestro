package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMaestroSnapSkillDocumentsConsumedGrpcExclusion(t *testing.T) {
	t.Helper()

	skillPath := filepath.Join("..", "..", "skills", "maestro-snap", "SKILL.md")
	templatePath := filepath.Join("..", "..", "skills", "maestro-snap", "OUTPUT-TEMPLATES.md")

	skill, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("read skill markdown: %v", err)
	}
	template, err := os.ReadFile(templatePath)
	if err != nil {
		t.Fatalf("read output template markdown: %v", err)
	}

	skillText := string(skill)
	templateText := string(template)

	for _, want := range []string{
		"After finding proto files, infer whether each gRPC service is provided or consumed",
		"Only services the current repo registers as server implementations should be documented in `api-contracts.md`.",
		"consumed gRPC services that were skipped",
	} {
		if !strings.Contains(skillText, want) {
			t.Fatalf("SKILL.md missing %q", want)
		}
	}

	if !strings.Contains(templateText, "Only include provided gRPC services; omit consumed services.") {
		t.Fatalf("OUTPUT-TEMPLATES.md missing consumed-service exclusion guidance")
	}
}
