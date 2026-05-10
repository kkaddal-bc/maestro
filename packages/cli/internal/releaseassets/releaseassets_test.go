package releaseassets

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateManifestUsesSkillDirectoriesAndDescriptions(t *testing.T) {
	skillsDir := t.TempDir()
	writeSkill(t, skillsDir, "zeta-tool", "Zeta description", nil)
	writeSkill(t, skillsDir, "alpha-tool", "Alpha description", map[string]string{
		"resources/note.txt": "resource",
	})

	got, err := GenerateManifest("v1.2.3", skillsDir)
	if err != nil {
		t.Fatalf("GenerateManifest() error = %v", err)
	}

	var manifest Manifest
	if err := json.Unmarshal(got, &manifest); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if manifest.Version != "v1.2.3" {
		t.Fatalf("Version = %q", manifest.Version)
	}
	if len(manifest.Skills) != 2 {
		t.Fatalf("len(Skills) = %d, want 2", len(manifest.Skills))
	}
	if manifest.Skills[0].Name != "alpha-tool" || manifest.Skills[0].Description != "Alpha description" {
		t.Fatalf("first skill = %#v", manifest.Skills[0])
	}
	if manifest.Skills[1].Name != "zeta-tool" || manifest.Skills[1].Description != "Zeta description" {
		t.Fatalf("second skill = %#v", manifest.Skills[1])
	}
}

func TestBuildSkillsArchivePreservesSkillTrees(t *testing.T) {
	skillsDir := t.TempDir()
	writeSkill(t, skillsDir, "alpha-tool", "Alpha description", map[string]string{
		"resources/note.txt": "resource",
	})
	writeSkill(t, skillsDir, "beta-tool", "Beta description", map[string]string{
		"docs/guide.md":    "guide",
		"docs/extra/readme": "readme",
	})

	archive, err := BuildSkillsArchive(skillsDir)
	if err != nil {
		t.Fatalf("BuildSkillsArchive() error = %v", err)
	}

	got := make(map[string]string)
	gzr, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		t.Fatalf("gzip.NewReader() error = %v", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("tar.Next() error = %v", err)
		}
		if hdr.FileInfo().IsDir() {
			continue
		}
		data, err := io.ReadAll(tr)
		if err != nil {
			t.Fatalf("ReadAll() error = %v", err)
		}
		got[hdr.Name] = string(data)
	}

	for _, want := range []string{
		"alpha-tool/SKILL.md",
		"alpha-tool/resources/note.txt",
		"beta-tool/SKILL.md",
		"beta-tool/docs/guide.md",
		"beta-tool/docs/extra/readme",
	} {
		if _, ok := got[want]; !ok {
			t.Fatalf("archive missing %q; entries=%v", want, got)
		}
	}
}

func TestWriteArtifactsWritesManifestAndArchive(t *testing.T) {
	skillsDir := t.TempDir()
	outDir := t.TempDir()
	writeSkill(t, skillsDir, "alpha-tool", "Alpha description", nil)

	if err := WriteArtifacts("v9.9.9", skillsDir, outDir); err != nil {
		t.Fatalf("WriteArtifacts() error = %v", err)
	}

	manifest, err := os.ReadFile(filepath.Join(outDir, "manifest.json"))
	if err != nil {
		t.Fatalf("ReadFile(manifest.json) error = %v", err)
	}
	if !bytes.Contains(manifest, []byte(`"version": "v9.9.9"`)) {
		t.Fatalf("manifest = %s", string(manifest))
	}

	if _, err := os.Stat(filepath.Join(outDir, "skills.tar.gz")); err != nil {
		t.Fatalf("skills.tar.gz missing: %v", err)
	}
}

func writeSkill(t *testing.T, root, name, description string, extraFiles map[string]string) {
	t.Helper()

	skillDir := filepath.Join(root, name)
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(%s): %v", skillDir, err)
	}

	skillMetadata := "---\nname: " + name + "\ndescription: " + description + "\n---\n"
	skillMetadataPath := filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(skillMetadataPath, []byte(skillMetadata), 0o644); err != nil {
		t.Fatalf("WriteFile(%s): %v", skillMetadataPath, err)
	}

	for rel, contents := range extraFiles {
		path := filepath.Join(skillDir, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("MkdirAll(%s): %v", path, err)
		}
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatalf("WriteFile(%s): %v", path, err)
		}
	}
}
