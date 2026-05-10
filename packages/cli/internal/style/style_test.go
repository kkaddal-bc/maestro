package style

import (
	"strings"
	"testing"
)

func TestAccentRendersWithANSI(t *testing.T) {
	got := Accent.Render("installed")
	if !strings.Contains(got, "\x1b[") {
		t.Fatalf("Accent.Render() = %q, want ANSI styling", got)
	}
	if !strings.Contains(got, "installed") {
		t.Fatalf("Accent.Render() = %q, want content preserved", got)
	}
}
