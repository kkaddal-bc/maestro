package releaseassets

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Manifest struct {
	Version string `json:"version"`
	Skills  []Skill `json:"skills"`
}

type Skill struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func WriteArtifacts(version, skillsDir, outputDir string) error {
	manifest, err := GenerateManifest(version, skillsDir)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(outputDir, "manifest.json"), manifest, 0o644); err != nil {
		return err
	}

	archive, err := BuildSkillsArchive(skillsDir)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(outputDir, "skills.tar.gz"), archive, 0o644)
}

func GenerateManifest(version, skillsDir string) ([]byte, error) {
	skills, err := discoverSkills(skillsDir)
	if err != nil {
		return nil, err
	}

	payload := Manifest{
		Version: version,
		Skills:  skills,
	}
	return json.MarshalIndent(payload, "", "  ")
}

func BuildSkillsArchive(skillsDir string) ([]byte, error) {
	skills, err := discoverSkillNames(skillsDir)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)

	for _, skill := range skills {
		root := filepath.Join(skillsDir, skill)
		if err := filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}

			rel, err := filepath.Rel(skillsDir, path)
			if err != nil {
				return err
			}
			name := filepath.ToSlash(rel)
			hdr, err := tar.FileInfoHeader(info, "")
			if err != nil {
				return err
			}
			hdr.Name = name
			if info.IsDir() && !strings.HasSuffix(hdr.Name, "/") {
				hdr.Name += "/"
			}
			if err := tw.WriteHeader(hdr); err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}

			file, err := os.Open(path)
			if err != nil {
				return err
			}
			_, copyErr := io.Copy(tw, file)
			closeErr := file.Close()
			if copyErr != nil {
				return copyErr
			}
			return closeErr
		}); err != nil {
			_ = tw.Close()
			_ = gz.Close()
			return nil, err
		}
	}

	if err := tw.Close(); err != nil {
		_ = gz.Close()
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func discoverSkills(skillsDir string) ([]Skill, error) {
	names, err := discoverSkillNames(skillsDir)
	if err != nil {
		return nil, err
	}

	skills := make([]Skill, 0, len(names))
	for _, name := range names {
		description, err := readDescription(filepath.Join(skillsDir, name, "SKILL.md"))
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", name, err)
		}
		skills = append(skills, Skill{Name: name, Description: description})
	}

	return skills, nil
}

func discoverSkillNames(skillsDir string) ([]string, error) {
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return nil, err
	}

	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			if _, err := os.Stat(filepath.Join(skillsDir, entry.Name(), "SKILL.md")); err == nil {
				names = append(names, entry.Name())
			} else if !os.IsNotExist(err) {
				return nil, err
			}
		}
	}
	sort.Strings(names)
	return names, nil
}

func readDescription(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) < 3 || strings.TrimSpace(lines[0]) != "---" {
		return "", fmt.Errorf("missing frontmatter")
	}

	end := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			end = i
			break
		}
	}
	if end == -1 {
		return "", fmt.Errorf("unterminated frontmatter")
	}

	for _, line := range lines[1:end] {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "description:") {
			continue
		}

		description := strings.TrimSpace(strings.TrimPrefix(trimmed, "description:"))
		description = strings.Trim(description, `"'`)
		if description == "" {
			return "", fmt.Errorf("missing description")
		}
		return description, nil
	}

	return "", fmt.Errorf("missing description")
}
