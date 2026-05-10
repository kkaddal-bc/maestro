package installer

import (
	"archive/tar"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kkaddal-bc/maestro/packages/cli/internal/targets"
)

type Result struct {
	Installed []Installation
}

type Installation struct {
	Skill  string
	Target string
}

type archiveEntry struct {
	relPath string
	mode    fs.FileMode
	data    []byte
	isDir   bool
}

func Install(skillNames []string, archive io.Reader, installTargets []targets.Target) (Result, error) {
	bundles, err := readArchive(archive)
	if err != nil {
		return Result{}, err
	}

	selected := skillNames
	if len(selected) == 0 {
		selected = sortedKeys(bundles)
	}
	if len(selected) == 0 {
		return Result{}, fmt.Errorf("archive contains no skills")
	}

	for _, name := range selected {
		if _, ok := bundles[name]; !ok {
			return Result{}, fmt.Errorf("unknown skill %q", name)
		}
	}

	result := Result{}
	for _, target := range installTargets {
		for _, skill := range selected {
			if err := writeSkill(target.Path, skill, bundles[skill]); err != nil {
				return result, err
			}
			result.Installed = append(result.Installed, Installation{
				Skill:  skill,
				Target: target.Path,
			})
		}
	}

	return result, nil
}

func readArchive(archive io.Reader) (map[string][]archiveEntry, error) {
	reader := tar.NewReader(archive)
	bundles := map[string][]archiveEntry{}

	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read archive: %w", err)
		}
		if header == nil {
			continue
		}

		cleanName := path.Clean(header.Name)
		if cleanName == "." || cleanName == "/" {
			continue
		}

		parts := strings.Split(cleanName, "/")
		if len(parts) == 0 || parts[0] == "." || parts[0] == "" {
			continue
		}

		skillName := parts[0]
		relPath := strings.Join(parts[1:], "/")
		entry := archiveEntry{
			relPath: relPath,
			mode:    fs.FileMode(header.Mode),
			isDir:   header.FileInfo().IsDir() || header.Typeflag == tar.TypeDir,
		}

		if header.Typeflag == tar.TypeReg || header.Typeflag == tar.TypeRegA || header.Typeflag == 0 {
			data, err := io.ReadAll(reader)
			if err != nil {
				return nil, fmt.Errorf("read archive entry %s: %w", header.Name, err)
			}
			entry.data = data
		}

		bundles[skillName] = append(bundles[skillName], entry)
	}

	return bundles, nil
}

func writeSkill(targetRoot, skill string, entries []archiveEntry) error {
	skillRoot := filepath.Join(targetRoot, skill)
	if err := os.MkdirAll(skillRoot, 0o755); err != nil {
		return fmt.Errorf("create target %s: %w", skillRoot, err)
	}

	for _, entry := range entries {
		dst := skillRoot
		if entry.relPath != "" {
			dst = filepath.Join(skillRoot, filepath.FromSlash(entry.relPath))
		}
		if entry.isDir {
			if err := os.MkdirAll(dst, os.FileMode(entry.mode)); err != nil {
				return fmt.Errorf("create directory %s: %w", dst, err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return fmt.Errorf("create parent for %s: %w", dst, err)
		}
		if err := os.WriteFile(dst, entry.data, os.FileMode(entry.mode)); err != nil {
			return fmt.Errorf("write file %s: %w", dst, err)
		}
	}

	return nil
}

func sortedKeys(bundles map[string][]archiveEntry) []string {
	keys := make([]string, 0, len(bundles))
	for key := range bundles {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
