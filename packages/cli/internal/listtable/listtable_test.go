package listtable

import (
	"bytes"
	"strings"
	"testing"

	"github.com/kkaddal-bc/maestro/packages/cli/internal/style"
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
	if len(lines) != 4 {
		t.Fatalf("rendered %d lines, want 4:\n%s", len(lines), got)
	}
	if wantSeparator := strings.Repeat("─", len(lines[0])); !strings.Contains(got, wantSeparator) {
		t.Fatalf("output missing separator %q:\n%s", wantSeparator, got)
	}
	if strings.Contains(got, "\t") {
		t.Fatalf("output unexpectedly contains tabs:\n%s", got)
	}
}

func TestStatusStyleUsesSuccessAndSecondary(t *testing.T) {
	installed := statusStyle("installed", style.Success, style.Secondary).Render("installed")
	if !strings.Contains(installed, "\x1b[") {
		t.Fatalf("installed status missing ANSI styling: %q", installed)
	}

	notInstalled := statusStyle("not installed", style.Success, style.Secondary).Render("not installed")
	if !strings.Contains(notInstalled, "\x1b[") {
		t.Fatalf("not installed status missing ANSI styling: %q", notInstalled)
	}
}

func TestTruncateDescriptionKeepsShortValuesIntact(t *testing.T) {
	renderer := NewRenderer()

	if got := renderer.truncateDescription("readable"); got != "readable" {
		t.Fatalf("truncateDescription() = %q, want %q", got, "readable")
	}
}
