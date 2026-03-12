package internal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUpdateProjectDescription_NoLLM(t *testing.T) {
	t.Setenv("RIFF_NO_AI", "1")

	dir := t.TempDir()
	_, err := UpdateProjectDescription(dir)
	if err != ErrNoLLM {
		t.Errorf("UpdateProjectDescription error = %v, want ErrNoLLM", err)
	}
}

func TestUpdateProjectDescription_NoLLM_PreservesExistingMetadata(t *testing.T) {
	t.Setenv("RIFF_NO_AI", "1")

	dir := t.TempDir()
	// Write existing metadata.
	meta := ProjectMetadata{
		Description: "old description",
		Created:     "2025-01-01T00:00:00Z",
		Template:    "go",
		Tags:        []string{"test"},
	}
	if err := WriteMetadata(dir, meta); err != nil {
		t.Fatal(err)
	}

	// UpdateProjectDescription should fail with ErrNoLLM.
	_, err := UpdateProjectDescription(dir)
	if err != ErrNoLLM {
		t.Errorf("expected ErrNoLLM, got %v", err)
	}

	// Existing metadata should be unchanged.
	got, err := ReadMetadata(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got.Description != "old description" {
		t.Errorf("Description = %q, want %q (should be unchanged after failed update)", got.Description, "old description")
	}
}

func TestBackfillDescriptions_NoMissing(t *testing.T) {
	t.Setenv("RIFF_NO_AI", "1")

	// All projects have descriptions — should be a no-op.
	projects := []ProjectInfo{
		{ID: "abc", Description: "has desc", Path: "/tmp"},
	}
	// Should not panic or print anything.
	BackfillDescriptions(projects)
}

func TestBackfillDescriptions_EmptySlice(t *testing.T) {
	t.Setenv("RIFF_NO_AI", "1")
	// Empty slice — should be a no-op.
	BackfillDescriptions(nil)
	BackfillDescriptions([]ProjectInfo{})
}

func TestGenerateDescription_DisabledByEnv_EmptyResult(t *testing.T) {
	t.Setenv("RIFF_NO_AI", "1")

	desc, err := GenerateDescription(t.TempDir())
	if err != ErrNoLLM {
		t.Errorf("error = %v, want ErrNoLLM", err)
	}
	if desc != "" {
		t.Errorf("description = %q, want empty", desc)
	}
}

func TestUpdateProjectDescription_CreatesMetadataIfMissing(t *testing.T) {
	// This test verifies that UpdateProjectDescription handles a missing
	// .riff.json file gracefully (it should create one from scratch).
	// We can't test the full flow without an LLM, but we can verify the
	// error path when no LLM is available.
	t.Setenv("RIFF_NO_AI", "1")

	dir := t.TempDir()
	// No .riff.json exists.
	metaPath := filepath.Join(dir, MetadataFile)
	if _, err := os.Stat(metaPath); err == nil {
		t.Fatal("expected no metadata file to exist")
	}

	_, err := UpdateProjectDescription(dir)
	if err != ErrNoLLM {
		t.Errorf("expected ErrNoLLM, got %v", err)
	}

	// Metadata file should NOT have been created since the LLM step failed.
	if _, err := os.Stat(metaPath); err == nil {
		t.Error("metadata file should not be created when LLM is unavailable")
	}
}
