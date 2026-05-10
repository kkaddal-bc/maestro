package targets

import (
	"os"
	"path/filepath"
)

type Target struct {
	Path     string
	Required bool
}

func Detect() []Target {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		home = os.Getenv("HOME")
	}
	return detect(home)
}

func detect(home string) []Target {
	targets := []Target{
		{Path: filepath.Join(home, ".maestro", "skills"), Required: true},
	}

	for _, parent := range []string{".claude", ".agents"} {
		parentPath := filepath.Join(home, parent)
		info, err := os.Stat(parentPath)
		if err != nil || !info.IsDir() {
			continue
		}
		targets = append(targets, Target{
			Path:     filepath.Join(parentPath, "skills"),
			Required: false,
		})
	}

	return targets
}
