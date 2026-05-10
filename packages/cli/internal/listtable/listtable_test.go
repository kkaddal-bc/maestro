package listtable

import (
	"bytes"
	"strings"
	"testing"
)

func TestRenderProducesAlignedPlainTextAndTruncatesDescriptions(t *testing.T) {
	var out bytes.Buffer

	renderer := NewRenderer()
	err := renderer.Render(&out, []string{"SKILL", "DESCRIPTION", "STATUS"}, []Row{
		{
			Name:        "maestro-snap",
			Description: strings.Repeat("a", 80),
			Status:      "installed",
		},
		{
			Name:        "other-skill",
			Description: "Short description",
			Status:      "not installed",
		},
	})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	got := out.String()
	if strings.Contains(got, "\x1b[") {
		t.Fatalf("output unexpectedly contains ANSI escapes:\n%s", got)
	}

	truncated := strings.Repeat("a", 59) + "…"
	for _, want := range []string{
		"SKILL",
		"DESCRIPTION",
		"STATUS",
		"maestro-snap",
		truncated,
		"other-skill",
		"installed",
		"not installed",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("output missing %q:\n%s", want, got)
		}
	}

	lines := strings.Split(strings.TrimSpace(got), "\n")
	if len(lines) != 3 {
		t.Fatalf("rendered %d lines, want 3:\n%s", len(lines), got)
	}
	if strings.Contains(got, "\t") {
		t.Fatalf("output unexpectedly contains tabs:\n%s", got)
	}
}

func TestTruncateDescriptionKeepsShortValuesIntact(t *testing.T) {
	renderer := NewRenderer()

	if got := renderer.truncateDescription("readable"); got != "readable" {
		t.Fatalf("truncateDescription() = %q, want %q", got, "readable")
	}
}
