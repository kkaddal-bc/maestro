package skilldocs_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMaestroSnapDocsAreManifestFirstAndLanguageAgnostic(t *testing.T) {
	root := repoRoot(t)

	skill := readFile(t, filepath.Join(root, "skills", "maestro-snap", "SKILL.md"))
	templates := readFile(t, filepath.Join(root, "skills", "maestro-snap", "OUTPUT-TEMPLATES.md"))

	for _, want := range []string{
		"go.mod",
		"package.json",
		"Cargo.toml",
		"build.gradle.kts",
		"find . -name \"*.proto\"",
		"openapi.yml",
		"swagger.yaml",
		"Proto files found",
		"OpenAPI/Swagger spec found",
		"AGENTS.md",
	} {
		if !strings.Contains(skill, want) {
			t.Fatalf("SKILL.md missing %q", want)
		}
	}

	for _, forbidden := range []string{
		"--include=\"*.kt\"",
		"@RestController",
		"@Controller",
		"SCAN-PATTERNS.md",
		"Kotlin",
		"Spring Boot",
		"Exposed",
		"data class",
	} {
		if strings.Contains(skill, forbidden) {
			t.Fatalf("SKILL.md still contains %q", forbidden)
		}
	}

	if !strings.Contains(templates, "🔒 (inferred)") {
		t.Fatalf("OUTPUT-TEMPLATES.md missing inferred auth marker")
	}

	for _, forbidden := range []string{
		"data class",
		"@field:",
		"ResponseEntity<",
		"Kotlin",
		"Spring Boot",
	} {
		if strings.Contains(templates, forbidden) {
			t.Fatalf("OUTPUT-TEMPLATES.md still contains %q", forbidden)
		}
	}

	if _, err := os.Stat(filepath.Join(root, "skills", "maestro-snap", "SCAN-PATTERNS.md")); !os.IsNotExist(err) {
		t.Fatalf("SCAN-PATTERNS.md should not exist")
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not locate repo root")
		}
		dir = parent
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", path, err)
	}
	return string(data)
}
