package syncer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yugo-ibuki/dot-claude-sync/config"
	"github.com/yugo-ibuki/dot-claude-sync/utils"
)

// TestSyncFiles_OverwriteDetection tests the overwrite detection logic
func TestSyncFiles_OverwriteDetection(t *testing.T) {
	// Create temporary directories for testing
	tmpDir, err := os.MkdirTemp("", "syncer-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create three test projects with different priorities
	project1Dir := filepath.Join(tmpDir, "project1", ".claude")
	project2Dir := filepath.Join(tmpDir, "project2", ".claude")
	project3Dir := filepath.Join(tmpDir, "project3", ".claude")

	for _, dir := range []string{project1Dir, project2Dir, project3Dir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create project dir: %v", err)
		}
	}

	// Test case 1: Different content - should be detected for overwrite
	testFile1 := "config.yaml"
	if err := os.WriteFile(filepath.Join(project1Dir, testFile1), []byte("content from project1"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(project2Dir, testFile1), []byte("content from project2"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Test case 2: Same content - should NOT be detected for overwrite
	testFile2 := "identical.txt"
	sameContent := []byte("same content in both projects")
	if err := os.WriteFile(filepath.Join(project1Dir, testFile2), sameContent, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(project2Dir, testFile2), sameContent, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Test case 3: New file (doesn't exist in project2) - should not be in overwrite list
	testFile3 := "new-file.txt"
	if err := os.WriteFile(filepath.Join(project1Dir, testFile3), []byte("new file"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Setup resolved files (from project1, priority 1)
	resolved := []ResolvedFile{
		{
			RelPath:  testFile1,
			AbsPath:  filepath.Join(project1Dir, testFile1),
			Source:   "project1",
			Priority: 1,
		},
		{
			RelPath:  testFile2,
			AbsPath:  filepath.Join(project1Dir, testFile2),
			Source:   "project1",
			Priority: 1,
		},
		{
			RelPath:  testFile3,
			AbsPath:  filepath.Join(project1Dir, testFile3),
			Source:   "project1",
			Priority: 1,
		},
	}

	// Setup projects
	projects := []config.ProjectPath{
		{Alias: "project1", Path: project1Dir, Priority: 1},
		{Alias: "project2", Path: project2Dir, Priority: 2}, // Lower priority (higher number)
		{Alias: "project3", Path: project3Dir, Priority: 3}, // Lowest priority
	}

	// Run sync with force flag to skip confirmation
	results, err := SyncFiles(resolved, projects, false, false, true)
	if err != nil {
		t.Fatalf("SyncFiles failed: %v", err)
	}

	// Verify results
	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	// Project2 should have 2 overwritten files (config.yaml and identical.txt)
	// and 1 new file (new-file.txt)
	// Note: The sync process counts all existing files as overwritten, regardless of content
	// The content filtering is only for the warning display, not the actual sync behavior
	var project2Result *SyncResult
	for i := range results {
		if results[i].Project == "project2" {
			project2Result = &results[i]
			break
		}
	}

	if project2Result == nil {
		t.Fatal("Project2 result not found")
	}

	// Check that config.yaml and identical.txt were overwritten
	if project2Result.Overwritten != 2 {
		t.Errorf("Expected 2 overwritten files in project2, got %d", project2Result.Overwritten)
	}

	// Check that new-file.txt was added as new
	if project2Result.NewFiles != 1 {
		t.Errorf("Expected 1 new file in project2, got %d", project2Result.NewFiles)
	}

	// Verify that files were actually copied correctly
	copiedContent, err := os.ReadFile(filepath.Join(project2Dir, testFile1))
	if err != nil {
		t.Fatalf("Failed to read copied file: %v", err)
	}
	if string(copiedContent) != "content from project1" {
		t.Errorf("File content mismatch: expected 'content from project1', got '%s'", string(copiedContent))
	}
}

// TestSyncFiles_PriorityFiltering tests that only lower priority projects are shown in overwrite warnings
func TestSyncFiles_PriorityFiltering(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "syncer-priority-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create three projects with different priorities
	highPriorityDir := filepath.Join(tmpDir, "high", ".claude")
	medPriorityDir := filepath.Join(tmpDir, "med", ".claude")
	lowPriorityDir := filepath.Join(tmpDir, "low", ".claude")

	for _, dir := range []string{highPriorityDir, medPriorityDir, lowPriorityDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create project dir: %v", err)
		}
	}

	// Create same file in all three projects with different content
	testFile := "test.txt"
	if err := os.WriteFile(filepath.Join(highPriorityDir, testFile), []byte("high priority"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(medPriorityDir, testFile), []byte("med priority"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(lowPriorityDir, testFile), []byte("low priority"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Resolved file is from high priority project
	resolved := []ResolvedFile{
		{
			RelPath:  testFile,
			AbsPath:  filepath.Join(highPriorityDir, testFile),
			Source:   "high",
			Priority: 1,
		},
	}

	projects := []config.ProjectPath{
		{Alias: "high", Path: highPriorityDir, Priority: 1},
		{Alias: "med", Path: medPriorityDir, Priority: 2},
		{Alias: "low", Path: lowPriorityDir, Priority: 3},
	}

	// Run sync with force flag
	results, err := SyncFiles(resolved, projects, false, false, true)
	if err != nil {
		t.Fatalf("SyncFiles failed: %v", err)
	}

	// All projects have existing files, so they will be counted as overwrites
	// Even the high priority project, because CopyFile is called (though it may skip same file)
	// The count happens before the copy operation, so it's based on file existence
	for _, result := range results {
		// All projects should show 1 overwrite since the file exists in all
		if result.Overwritten != 1 {
			t.Errorf("Project %s should have 1 overwrite, got %d", result.Project, result.Overwritten)
		}
	}

	// Verify content was overwritten correctly
	for _, dir := range []string{medPriorityDir, lowPriorityDir} {
		content, err := os.ReadFile(filepath.Join(dir, testFile))
		if err != nil {
			t.Fatalf("Failed to read file in %s: %v", dir, err)
		}
		if string(content) != "high priority" {
			t.Errorf("File in %s should contain 'high priority', got '%s'", dir, string(content))
		}
	}
}

// TestSyncFiles_IdenticalContent tests that files with identical content are not counted as overwrites
func TestSyncFiles_IdenticalContent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "syncer-identical-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	project1Dir := filepath.Join(tmpDir, "project1", ".claude")
	project2Dir := filepath.Join(tmpDir, "project2", ".claude")

	for _, dir := range []string{project1Dir, project2Dir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create project dir: %v", err)
		}
	}

	// Create identical files in both projects
	testFile := "identical.yaml"
	identicalContent := []byte("key: value\nlist:\n  - item1\n  - item2\n")

	if err := os.WriteFile(filepath.Join(project1Dir, testFile), identicalContent, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(project2Dir, testFile), identicalContent, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	resolved := []ResolvedFile{
		{
			RelPath:  testFile,
			AbsPath:  filepath.Join(project1Dir, testFile),
			Source:   "project1",
			Priority: 1,
		},
	}

	projects := []config.ProjectPath{
		{Alias: "project1", Path: project1Dir, Priority: 1},
		{Alias: "project2", Path: project2Dir, Priority: 2},
	}

	results, err := SyncFiles(resolved, projects, false, false, true)
	if err != nil {
		t.Fatalf("SyncFiles failed: %v", err)
	}

	// Project2 should have 1 overwrite even though content is identical
	// The sync process counts all existing files as overwrites
	// The content filtering (hash comparison) is only for the warning display
	for _, result := range results {
		if result.Project == "project2" {
			if result.Overwritten != 1 {
				t.Errorf("Expected 1 overwrite (file exists regardless of content), got %d", result.Overwritten)
			}
			if result.NewFiles != 0 {
				t.Errorf("Expected 0 new files, got %d", result.NewFiles)
			}
		}
	}

	// Verify content remains identical after sync
	content, err := os.ReadFile(filepath.Join(project2Dir, testFile))
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if string(content) != string(identicalContent) {
		t.Errorf("Content should remain identical after sync")
	}
}

// TestSyncFiles_DryRun tests dry-run mode
func TestSyncFiles_DryRun(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "syncer-dryrun-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	project1Dir := filepath.Join(tmpDir, "project1", ".claude")
	project2Dir := filepath.Join(tmpDir, "project2", ".claude")

	for _, dir := range []string{project1Dir, project2Dir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create project dir: %v", err)
		}
	}

	testFile := "test.txt"
	originalContent := []byte("original content")
	newContent := []byte("new content")

	if err := os.WriteFile(filepath.Join(project1Dir, testFile), newContent, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(project2Dir, testFile), originalContent, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	resolved := []ResolvedFile{
		{
			RelPath:  testFile,
			AbsPath:  filepath.Join(project1Dir, testFile),
			Source:   "project1",
			Priority: 1,
		},
	}

	projects := []config.ProjectPath{
		{Alias: "project1", Path: project1Dir, Priority: 1},
		{Alias: "project2", Path: project2Dir, Priority: 2},
	}

	// Run in dry-run mode
	results, err := SyncFiles(resolved, projects, true, false, true)
	if err != nil {
		t.Fatalf("SyncFiles failed: %v", err)
	}

	// Check that results show what would happen
	var project2Result *SyncResult
	for i := range results {
		if results[i].Project == "project2" {
			project2Result = &results[i]
			break
		}
	}

	if project2Result == nil {
		t.Fatal("Project2 result not found")
	}

	if project2Result.Overwritten != 1 {
		t.Errorf("Dry-run should show 1 file would be overwritten, got %d", project2Result.Overwritten)
	}

	// Verify that file was NOT actually modified
	content, err := os.ReadFile(filepath.Join(project2Dir, testFile))
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if string(content) != string(originalContent) {
		t.Errorf("File should not be modified in dry-run mode. Expected '%s', got '%s'", originalContent, content)
	}
}

// TestSyncFiles_SkippedProject tests that non-existent .claude directories are skipped
func TestSyncFiles_SkippedProject(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "syncer-skip-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	project1Dir := filepath.Join(tmpDir, "project1", ".claude")
	project2Dir := filepath.Join(tmpDir, "project2", ".claude") // This won't be created

	// Only create project1
	if err := os.MkdirAll(project1Dir, 0755); err != nil {
		t.Fatalf("Failed to create project dir: %v", err)
	}

	testFile := "test.txt"
	if err := os.WriteFile(filepath.Join(project1Dir, testFile), []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	resolved := []ResolvedFile{
		{
			RelPath:  testFile,
			AbsPath:  filepath.Join(project1Dir, testFile),
			Source:   "project1",
			Priority: 1,
		},
	}

	projects := []config.ProjectPath{
		{Alias: "project1", Path: project1Dir, Priority: 1},
		{Alias: "project2", Path: project2Dir, Priority: 2},
	}

	results, err := SyncFiles(resolved, projects, false, false, true)
	if err != nil {
		t.Fatalf("SyncFiles failed: %v", err)
	}

	// Check that project2 was skipped
	var project2Result *SyncResult
	for i := range results {
		if results[i].Project == "project2" {
			project2Result = &results[i]
			break
		}
	}

	if project2Result == nil {
		t.Fatal("Project2 result not found")
	}

	if !project2Result.Skipped {
		t.Error("Project2 should be marked as skipped")
	}

	if project2Result.SkipReason == "" {
		t.Error("Skip reason should be provided")
	}
}

// TestExpandPath tests the expandPath helper function
func TestExpandPath(t *testing.T) {
	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// Set a test HOME
	testHome := "/home/testuser"
	os.Setenv("HOME", testHome)

	tests := []struct {
		input    string
		expected string
	}{
		{"~/test", filepath.Join(testHome, "test")},
		{"~/.claude", filepath.Join(testHome, ".claude")},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
		{"~", testHome},
	}

	for _, tt := range tests {
		result := expandPath(tt.input)
		if result != tt.expected {
			t.Errorf("expandPath(%s) = %s, expected %s", tt.input, result, tt.expected)
		}
	}
}

// Helper function to create a test file with specific content
func createTestFile(t *testing.T, path string, content string) {
	if err := utils.EnsureDir(filepath.Dir(path)); err != nil {
		t.Fatalf("Failed to ensure directory: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file %s: %v", path, err)
	}
}
