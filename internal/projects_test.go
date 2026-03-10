package internal

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// setupTestRiffHome creates a temporary RIFF_HOME and re-initialises paths.
// Cleanup is automatic via t.TempDir.
func setupTestRiffHome(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("RIFF_HOME", dir)
	t.Setenv("RIFF_NO_AI", "1")
	InitPaths()
	return dir
}

func TestGenerateID_DefaultLength(t *testing.T) {
	id := GenerateID(0)
	if len(id) != 7 {
		t.Errorf("GenerateID(0) returned length %d, want 7", len(id))
	}
}

func TestGenerateID_CustomLength(t *testing.T) {
	for _, length := range []int{1, 5, 10, 20} {
		id := GenerateID(length)
		if len(id) != length {
			t.Errorf("GenerateID(%d) returned length %d", length, len(id))
		}
	}
}

func TestGenerateID_NegativeLength(t *testing.T) {
	id := GenerateID(-3)
	if len(id) != 7 {
		t.Errorf("GenerateID(-3) returned length %d, want 7", len(id))
	}
}

func TestGenerateID_Charset(t *testing.T) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	allowed := make(map[byte]bool)
	for i := range charset {
		allowed[charset[i]] = true
	}

	// Generate a long ID to get good charset coverage.
	id := GenerateID(1000)
	for i := range id {
		if !allowed[id[i]] {
			t.Errorf("GenerateID produced invalid character %q at position %d", id[i], i)
		}
	}
}

func TestGenerateID_Uniqueness(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		id := GenerateID(7)
		if seen[id] {
			t.Fatalf("GenerateID produced duplicate %q on iteration %d", id, i)
		}
		seen[id] = true
	}
}

func TestFormatAge(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name string
		iso  string
		want string
	}{
		{"empty", "", "unknown"},
		{"invalid", "not-a-date", "unknown"},
		{"just now", now.Add(-5 * time.Second).Format(time.RFC3339), "just now"},
		{"1 minute ago", now.Add(-90 * time.Second).Format(time.RFC3339), "1 minute ago"},
		{"minutes ago", now.Add(-15 * time.Minute).Format(time.RFC3339), "15 minutes ago"},
		{"1 hour ago", now.Add(-90 * time.Minute).Format(time.RFC3339), "1 hour ago"},
		{"hours ago", now.Add(-5 * time.Hour).Format(time.RFC3339), "5 hours ago"},
		{"1 day ago", now.Add(-30 * time.Hour).Format(time.RFC3339), "1 day ago"},
		{"days ago", now.Add(-10 * 24 * time.Hour).Format(time.RFC3339), "10 days ago"},
		{"future", now.Add(1 * time.Hour).Format(time.RFC3339), "just now"},
		{"old date", "2020-01-15T10:00:00Z", "Jan 15, 2020"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatAge(tt.iso)
			if got != tt.want {
				t.Errorf("FormatAge(%q) = %q, want %q", tt.iso, got, tt.want)
			}
		})
	}
}

func TestReadWriteMetadata_Roundtrip(t *testing.T) {
	dir := t.TempDir()

	meta := ProjectMetadata{
		Description: "test project",
		Created:     "2025-01-01T00:00:00Z",
		Template:    "bun",
		Tags:        []string{"web", "test"},
	}

	if err := WriteMetadata(dir, meta); err != nil {
		t.Fatalf("WriteMetadata: %v", err)
	}

	got, err := ReadMetadata(dir)
	if err != nil {
		t.Fatalf("ReadMetadata: %v", err)
	}

	if got.Description != meta.Description {
		t.Errorf("Description = %q, want %q", got.Description, meta.Description)
	}
	if got.Template != meta.Template {
		t.Errorf("Template = %q, want %q", got.Template, meta.Template)
	}
	if got.Created != meta.Created {
		t.Errorf("Created = %q, want %q", got.Created, meta.Created)
	}
	if len(got.Tags) != len(meta.Tags) {
		t.Errorf("Tags length = %d, want %d", len(got.Tags), len(meta.Tags))
	}
}

func TestReadMetadata_MissingFile(t *testing.T) {
	dir := t.TempDir()
	_, err := ReadMetadata(dir)
	if err == nil {
		t.Error("ReadMetadata on missing file should return error")
	}
}

func TestReadMetadata_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, MetadataFile), []byte("{invalid"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := ReadMetadata(dir)
	if err == nil {
		t.Error("ReadMetadata on invalid JSON should return error")
	}
}

func TestGetProjects_Empty(t *testing.T) {
	setupTestRiffHome(t)
	if err := EnsureRiffDir(); err != nil {
		t.Fatal(err)
	}

	projects, err := GetProjects()
	if err != nil {
		t.Fatalf("GetProjects: %v", err)
	}
	if len(projects) != 0 {
		t.Errorf("expected 0 projects, got %d", len(projects))
	}
}

func TestGetProjects_NoDir(t *testing.T) {
	setupTestRiffHome(t)
	// Don't call EnsureRiffDir — projects dir doesn't exist.

	projects, err := GetProjects()
	if err != nil {
		t.Fatalf("GetProjects: %v", err)
	}
	if projects != nil {
		t.Errorf("expected nil, got %v", projects)
	}
}

func TestGetProjects_WithMetadata(t *testing.T) {
	setupTestRiffHome(t)
	if err := EnsureRiffDir(); err != nil {
		t.Fatal(err)
	}

	// Create two projects with metadata.
	for _, id := range []string{"aaa1111", "zzz9999"} {
		dir := filepath.Join(ProjectsDir, id)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		meta := ProjectMetadata{
			Description: "project " + id,
			Created:     "2025-01-01T00:00:00Z",
			Template:    "go",
		}
		if err := WriteMetadata(dir, meta); err != nil {
			t.Fatal(err)
		}
	}

	projects, err := GetProjects()
	if err != nil {
		t.Fatalf("GetProjects: %v", err)
	}

	if len(projects) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(projects))
	}

	// Should be sorted by ID.
	if projects[0].ID != "aaa1111" {
		t.Errorf("first project ID = %q, want %q", projects[0].ID, "aaa1111")
	}
	if projects[1].ID != "zzz9999" {
		t.Errorf("second project ID = %q, want %q", projects[1].ID, "zzz9999")
	}

	// Check metadata was read.
	if projects[0].Description != "project aaa1111" {
		t.Errorf("Description = %q, want %q", projects[0].Description, "project aaa1111")
	}
	if projects[0].Template != "go" {
		t.Errorf("Template = %q, want %q", projects[0].Template, "go")
	}
}

func TestGetProjects_PackageJSONFallback(t *testing.T) {
	setupTestRiffHome(t)
	if err := EnsureRiffDir(); err != nil {
		t.Fatal(err)
	}

	// Create a project with only package.json (no .riff.json).
	dir := filepath.Join(ProjectsDir, "pkg0001")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	pkgJSON := `{"name": "test", "description": "from package.json"}`
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgJSON), 0644); err != nil {
		t.Fatal(err)
	}

	projects, err := GetProjects()
	if err != nil {
		t.Fatalf("GetProjects: %v", err)
	}

	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}
	if projects[0].Description != "from package.json" {
		t.Errorf("Description = %q, want %q", projects[0].Description, "from package.json")
	}
}

func TestGetProjects_SkipsHiddenEntries(t *testing.T) {
	setupTestRiffHome(t)
	if err := EnsureRiffDir(); err != nil {
		t.Fatal(err)
	}

	// Create a hidden dir and a normal dir.
	if err := os.MkdirAll(filepath.Join(ProjectsDir, ".hidden"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(ProjectsDir, "visible"), 0755); err != nil {
		t.Fatal(err)
	}

	projects, err := GetProjects()
	if err != nil {
		t.Fatalf("GetProjects: %v", err)
	}

	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}
	if projects[0].ID != "visible" {
		t.Errorf("ID = %q, want %q", projects[0].ID, "visible")
	}
}

func TestGetProjects_SkipsFiles(t *testing.T) {
	setupTestRiffHome(t)
	if err := EnsureRiffDir(); err != nil {
		t.Fatal(err)
	}

	// Create a file (not a directory) in projects dir.
	if err := os.WriteFile(filepath.Join(ProjectsDir, "notadir"), []byte("hi"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(ProjectsDir, "realdir"), 0755); err != nil {
		t.Fatal(err)
	}

	projects, err := GetProjects()
	if err != nil {
		t.Fatalf("GetProjects: %v", err)
	}

	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}
	if projects[0].ID != "realdir" {
		t.Errorf("ID = %q, want %q", projects[0].ID, "realdir")
	}
}

func TestWriteCdPath(t *testing.T) {
	setupTestRiffHome(t)
	if err := os.MkdirAll(RiffDir, 0755); err != nil {
		t.Fatal(err)
	}

	path := "/some/project/path"
	if err := WriteCdPath(path); err != nil {
		t.Fatalf("WriteCdPath: %v", err)
	}

	data, err := os.ReadFile(CdPathFile)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(data) != path {
		t.Errorf("CdPath = %q, want %q", string(data), path)
	}
}

func TestEnsureRiffDir(t *testing.T) {
	setupTestRiffHome(t)

	if err := EnsureRiffDir(); err != nil {
		t.Fatalf("EnsureRiffDir: %v", err)
	}

	info, err := os.Stat(ProjectsDir)
	if err != nil {
		t.Fatalf("ProjectsDir doesn't exist: %v", err)
	}
	if !info.IsDir() {
		t.Error("ProjectsDir is not a directory")
	}
}
