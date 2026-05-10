package style

import (
	"strings"
	"testing"
)

func TestSuccessRendersWithANSI(t *testing.T) {
	got := Success.Render("installed")
	if !strings.Contains(got, "\x1b[") {
		t.Fatalf("Success.Render() = %q, want ANSI styling", got)
	}
	if !strings.Contains(got, "installed") {
		t.Fatalf("Success.Render() = %q, want content preserved", got)
	}
}

func TestSecondaryRendersWithANSI(t *testing.T) {
	got := Secondary.Render("secondary")
	if !strings.Contains(got, "\x1b[") {
		t.Fatalf("Secondary.Render() = %q, want ANSI styling", got)
	}
	if !strings.Contains(got, "secondary") {
		t.Fatalf("Secondary.Render() = %q, want content preserved", got)
	}
}
