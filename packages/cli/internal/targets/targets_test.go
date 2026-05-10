package targets

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectReturnsAlwaysAndConditionalTargets(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	if err := os.MkdirAll(filepath.Join(home, ".claude"), 0o755); err != nil {
		t.Fatalf("mkdir .claude: %v", err)
	}

	got := detect(home)
	if len(got) != 2 {
		t.Fatalf("len(detect()) = %d, want 2", len(got))
	}

	wantMaestro := filepath.Join(home, ".maestro", "skills")
	if got[0].Path != wantMaestro || !got[0].Required {
		t.Fatalf("maestro target = %#v", got[0])
	}

	wantClaude := filepath.Join(home, ".claude", "skills")
	if got[1].Path != wantClaude || got[1].Required {
		t.Fatalf("claude target = %#v", got[1])
	}
}

func TestDetectSkipsMissingOptionalTargets(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	if err := os.MkdirAll(filepath.Join(home, ".agents"), 0o755); err != nil {
		t.Fatalf("mkdir .agents: %v", err)
	}

	got := detect(home)
	if len(got) != 2 {
		t.Fatalf("len(detect()) = %d, want 2", len(got))
	}

	wantMaestro := filepath.Join(home, ".maestro", "skills")
	wantAgents := filepath.Join(home, ".agents", "skills")
	if got[0].Path != wantMaestro {
		t.Fatalf("maestro target = %#v", got[0])
	}
	if got[1].Path != wantAgents {
		t.Fatalf("agents target = %#v", got[1])
	}
}
