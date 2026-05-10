package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMaestroSnapSkillDocumentsConsumedGrpcExclusion(t *testing.T) {
	t.Run("skill markdown", func(t *testing.T) {
		assertContainsAll(t, readMarkdown(t, filepath.Join("..", "..", "skills", "maestro-snap", "SKILL.md")), []string{
			"After finding proto files, infer whether each gRPC service is provided or consumed",
			"Only services the current repo registers as server implementations should be documented in `api-contracts.md`.",
			"consumed gRPC services that were skipped",
		})
	})

	t.Run("output templates", func(t *testing.T) {
		assertContainsAll(t, readMarkdown(t, filepath.Join("..", "..", "skills", "maestro-snap", "OUTPUT-TEMPLATES.md")), []string{
			"Only include provided gRPC services; omit consumed services.",
		})
	})
}

func readMarkdown(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read markdown %s: %v", path, err)
	}

	return string(content)
}

func assertContainsAll(t *testing.T, text string, wants []string) {
	t.Helper()

	for _, want := range wants {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q", want)
		}
	}
}
