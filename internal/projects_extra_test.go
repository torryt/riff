package internal

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGetProjects_SortedNewestFirst(t *testing.T) {
	setupTestRiffHome(t)
	if err := EnsureRiffDir(); err != nil {
		t.Fatal(err)
	}

	// Create three projects with different timestamps.
	projects := []struct {
		id      string
		created string
	}{
		{"oldest1", "2025-01-01T00:00:00Z"},
		{"newest1", "2025-03-01T00:00:00Z"},
		{"middle1", "2025-02-01T00:00:00Z"},
	}

	for _, p := range projects {
		dir := filepath.Join(ProjectsDir, p.id)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		meta := ProjectMetadata{
			Description: "project " + p.id,
			Created:     p.created,
		}
		if err := WriteMetadata(dir, meta); err != nil {
			t.Fatal(err)
		}
	}

	got, err := GetProjects()
	if err != nil {
		t.Fatalf("GetProjects: %v", err)
	}

	if len(got) != 3 {
		t.Fatalf("expected 3 projects, got %d", len(got))
	}

	// Should be sorted newest first.
	if got[0].ID != "newest1" {
		t.Errorf("first project = %q, want newest1", got[0].ID)
	}
	if got[1].ID != "middle1" {
		t.Errorf("second project = %q, want middle1", got[1].ID)
	}
	if got[2].ID != "oldest1" {
		t.Errorf("third project = %q, want oldest1", got[2].ID)
	}
}

func TestGetProjects_MixedTimestampsAndMissing(t *testing.T) {
	setupTestRiffHome(t)
	if err := EnsureRiffDir(); err != nil {
		t.Fatal(err)
	}

	// Create a project with a timestamp.
	dir1 := filepath.Join(ProjectsDir, "withtime")
	if err := os.MkdirAll(dir1, 0755); err != nil {
		t.Fatal(err)
	}
	meta1 := ProjectMetadata{Created: "2025-06-01T00:00:00Z"}
	if err := WriteMetadata(dir1, meta1); err != nil {
		t.Fatal(err)
	}

	// Create a project without metadata (no timestamp).
	dir2 := filepath.Join(ProjectsDir, "notime1")
	if err := os.MkdirAll(dir2, 0755); err != nil {
		t.Fatal(err)
	}

	got, err := GetProjects()
	if err != nil {
		t.Fatalf("GetProjects: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(got))
	}

	// Project with timestamp should come first; one without goes to end.
	if got[0].ID != "withtime" {
		t.Errorf("first project = %q, want withtime (has timestamp)", got[0].ID)
	}
	if got[1].ID != "notime1" {
		t.Errorf("second project = %q, want notime1 (no timestamp)", got[1].ID)
	}
}

func TestWriteMetadata_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	meta := ProjectMetadata{
		Description: "test",
		Created:     time.Now().Format(time.RFC3339),
		Template:    "go",
		Tags:        []string{},
	}

	if err := WriteMetadata(dir, meta); err != nil {
		t.Fatalf("WriteMetadata: %v", err)
	}

	// Verify file exists.
	metaPath := filepath.Join(dir, MetadataFile)
	if _, err := os.Stat(metaPath); err != nil {
		t.Fatalf("metadata file not created: %v", err)
	}
}

func TestWriteMetadata_OverwritesExisting(t *testing.T) {
	dir := t.TempDir()
	meta1 := ProjectMetadata{Description: "first"}
	meta2 := ProjectMetadata{Description: "second"}

	if err := WriteMetadata(dir, meta1); err != nil {
		t.Fatal(err)
	}
	if err := WriteMetadata(dir, meta2); err != nil {
		t.Fatal(err)
	}

	got, err := ReadMetadata(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got.Description != "second" {
		t.Errorf("Description = %q, want %q after overwrite", got.Description, "second")
	}
}

func TestEnsureRiffDir_Idempotent(t *testing.T) {
	setupTestRiffHome(t)

	// Call twice — should not error.
	if err := EnsureRiffDir(); err != nil {
		t.Fatalf("first EnsureRiffDir: %v", err)
	}
	if err := EnsureRiffDir(); err != nil {
		t.Fatalf("second EnsureRiffDir: %v", err)
	}

	info, err := os.Stat(ProjectsDir)
	if err != nil {
		t.Fatalf("ProjectsDir missing: %v", err)
	}
	if !info.IsDir() {
		t.Error("ProjectsDir is not a directory")
	}
}

func TestGetProjects_PackageJSON_InvalidJSON(t *testing.T) {
	setupTestRiffHome(t)
	if err := EnsureRiffDir(); err != nil {
		t.Fatal(err)
	}

	// Create a project with an invalid package.json and no .riff.json.
	dir := filepath.Join(ProjectsDir, "badpkg1")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte("{broken"), 0644); err != nil {
		t.Fatal(err)
	}

	projects, err := GetProjects()
	if err != nil {
		t.Fatalf("GetProjects: %v", err)
	}

	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}
	// Description should be empty since package.json is invalid.
	if projects[0].Description != "" {
		t.Errorf("Description = %q, want empty for invalid package.json", projects[0].Description)
	}
}

func TestWriteCdPath_Overwrite(t *testing.T) {
	setupTestRiffHome(t)
	if err := os.MkdirAll(RiffDir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := WriteCdPath("/first/path"); err != nil {
		t.Fatal(err)
	}
	if err := WriteCdPath("/second/path"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(CdPathFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "/second/path" {
		t.Errorf("CdPath = %q, want %q after overwrite", string(data), "/second/path")
	}
}

func TestMetadata_TagsNilVsEmpty(t *testing.T) {
	dir := t.TempDir()

	// Write with nil tags.
	meta := ProjectMetadata{
		Description: "test",
		Created:     "2025-01-01T00:00:00Z",
		Tags:        nil,
	}
	if err := WriteMetadata(dir, meta); err != nil {
		t.Fatal(err)
	}

	got, err := ReadMetadata(dir)
	if err != nil {
		t.Fatal(err)
	}
	// nil marshals as JSON null, which unmarshals as nil slice.
	if got.Tags != nil {
		t.Errorf("Tags = %v, want nil", got.Tags)
	}
}
