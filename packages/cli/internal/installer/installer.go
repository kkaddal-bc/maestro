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

type UpdateResult struct {
	Updated  []Installation
	UpToDate bool
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

func Update(skillNames []string, archive io.Reader, installTargets []targets.Target) (UpdateResult, error) {
	bundles, err := readArchive(archive)
	if err != nil {
		return UpdateResult{}, err
	}

	selected := skillNames
	if len(selected) == 0 {
		selected = sortedKeys(bundles)
	}

	for _, name := range selected {
		if _, ok := bundles[name]; !ok {
			return UpdateResult{}, fmt.Errorf("unknown skill %q", name)
		}
	}

	result := UpdateResult{UpToDate: true}
	for _, target := range installTargets {
		for _, skill := range selected {
			skillRoot := filepath.Join(target.Path, skill)
			exists, err := pathExists(skillRoot)
			if err != nil {
				return result, err
			}
			if !exists {
				continue
			}

			matches, err := skillTreeMatches(skillRoot, bundles[skill])
			if err != nil {
				return result, err
			}
			if matches {
				continue
			}

			if err := writeSkill(target.Path, skill, bundles[skill]); err != nil {
				return result, err
			}
			result.UpToDate = false
			result.Updated = append(result.Updated, Installation{
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
	if err := os.RemoveAll(skillRoot); err != nil {
		return fmt.Errorf("clear target %s: %w", skillRoot, err)
	}
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

func skillTreeMatches(skillRoot string, entries []archiveEntry) (bool, error) {
	expectedFiles, expectedDirs := expectedArchiveTree(entries)

	seenFiles := map[string]struct{}{}
	seenDirs := map[string]struct{}{
		".": {},
	}

	walkErr := filepath.WalkDir(skillRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(skillRoot, path)
		if err != nil {
			return err
		}
		rel = filepath.Clean(rel)
		if rel == "." {
			return nil
		}

		if d.IsDir() {
			seenDirs[rel] = struct{}{}
			if _, ok := expectedDirs[rel]; ok {
				return nil
			}
			if _, ok := expectedFiles[rel]; ok {
				return fmt.Errorf("expected file but found directory %s", rel)
			}
			return fmt.Errorf("unexpected directory %s", rel)
		}

		entry, ok := expectedFiles[rel]
		if !ok {
			return fmt.Errorf("unexpected file %s", rel)
		}
		seenFiles[rel] = struct{}{}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if string(data) != string(entry.data) {
			return fmt.Errorf("file %s does not match archive", rel)
		}

		info, err := d.Info()
		if err != nil {
			return err
		}
		if info.Mode().Perm() != entry.mode.Perm() {
			return fmt.Errorf("file %s mode does not match archive", rel)
		}
		return nil
	})

	if walkErr != nil {
		return false, walkErr
	}

	if len(seenFiles) != len(expectedFiles) {
		return false, nil
	}
	for rel := range expectedFiles {
		if _, ok := seenFiles[rel]; !ok {
			return false, nil
		}
	}
	if len(seenDirs) != len(expectedDirs) {
		return false, nil
	}
	for rel := range expectedDirs {
		if _, ok := seenDirs[rel]; !ok {
			return false, nil
		}
	}

	return true, nil
}

func expectedArchiveTree(entries []archiveEntry) (map[string]archiveEntry, map[string]struct{}) {
	expectedFiles := map[string]archiveEntry{}
	expectedDirs := map[string]struct{}{
		".": {},
	}

	for _, entry := range entries {
		cleanRel := filepath.Clean(filepath.FromSlash(entry.relPath))
		if entry.relPath == "" || cleanRel == "." {
			continue
		}

		addExpectedParentDirs(cleanRel, expectedDirs)
		if entry.isDir {
			expectedDirs[cleanRel] = struct{}{}
			continue
		}
		expectedFiles[cleanRel] = entry
	}

	return expectedFiles, expectedDirs
}

func addExpectedParentDirs(relPath string, expectedDirs map[string]struct{}) {
	for parent := filepath.Dir(relPath); parent != "."; parent = filepath.Dir(parent) {
		expectedDirs[parent] = struct{}{}
	}
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func sortedKeys(bundles map[string][]archiveEntry) []string {
	keys := make([]string, 0, len(bundles))
	for key := range bundles {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
