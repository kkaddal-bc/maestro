package manifest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

type Manifest struct {
	Version string       `json:"version"`
	Skills  []SkillEntry `json:"skills"`
}

type SkillEntry struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func Parse(r io.Reader) (*Manifest, error) {
	var raw map[string]json.RawMessage
	dec := json.NewDecoder(r)
	if err := dec.Decode(&raw); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}

	if _, ok := raw["version"]; !ok {
		return nil, errors.New("manifest missing version")
	}
	if _, ok := raw["skills"]; !ok {
		return nil, errors.New("manifest missing skills")
	}

	var m Manifest
	if err := json.Unmarshal(raw["version"], &m.Version); err != nil {
		return nil, fmt.Errorf("parse manifest version: %w", err)
	}
	if err := json.Unmarshal(raw["skills"], &m.Skills); err != nil {
		return nil, fmt.Errorf("parse manifest skills: %w", err)
	}

	if m.Version == "" {
		return nil, errors.New("manifest missing version")
	}
	for _, skill := range m.Skills {
		if skill.Name == "" {
			return nil, errors.New("manifest skill missing name")
		}
		if skill.Description == "" {
			return nil, fmt.Errorf("manifest skill %q missing description", skill.Name)
		}
	}
	return &m, nil
}
