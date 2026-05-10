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
	knownTargets := Known(home)

	detected := make([]Target, 0, len(knownTargets))
	for _, target := range knownTargets {
		if target.Required || targetExists(target.Path) {
			detected = append(detected, target)
			continue
		}
	}

	return detected
}

func Known(home string) []Target {
	return []Target{
		{Path: filepath.Join(home, ".maestro", "skills"), Required: true},
		{Path: filepath.Join(home, ".claude", "skills"), Required: false},
		{Path: filepath.Join(home, ".agents", "skills"), Required: false},
	}
}

func targetExists(targetPath string) bool {
	parentDir := filepath.Dir(targetPath)
	info, err := os.Stat(parentDir)
	return err == nil && info.IsDir()
}
