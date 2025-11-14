package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yugo-ibuki/dot-claude-sync/config"
	"github.com/yugo-ibuki/dot-claude-sync/syncer"
	"github.com/yugo-ibuki/dot-claude-sync/utils"
)

// TestPushWorkflow tests the complete push workflow
func TestPushWorkflow(t *testing.T) {
	// Create temporary directory for test projects
	tmpDir := t.TempDir()

	// Create test project structure
	project1 := filepath.Join(tmpDir, "project1", ".claude")
	project2 := filepath.Join(tmpDir, "project2", ".claude")
	project3 := filepath.Join(tmpDir, "project3", ".claude")

	for _, dir := range []string{project1, project2, project3} {
		if err := os.MkdirAll(filepath.Join(dir, "prompts"), 0755); err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}
	}

	t.Run("collect files from multiple projects", func(t *testing.T) {
		// Create test files in different projects
		files := map[string]string{
			filepath.Join(project1, "prompts", "coding.md"):  "coding prompt from project1",
			filepath.Join(project1, "config.json"):           `{"setting": "value1"}`,
			filepath.Join(project2, "prompts", "coding.md"):  "coding prompt from project2",
			filepath.Join(project2, "prompts", "testing.md"): "testing prompt from project2",
			filepath.Join(project3, "prompts", "review.md"):  "review prompt from project3",
		}

		for path, content := range files {
			if err := os.WriteFile(path, []byte(content), 0644); err != nil {
				t.Fatalf("Failed to create test file %s: %v", path, err)
			}
		}

		// Create config
		projects := []config.ProjectPath{
			{Alias: "project1", Path: project1, Priority: 1},
			{Alias: "project2", Path: project2, Priority: 2},
			{Alias: "project3", Path: project3, Priority: 3},
		}

		// Collect files
		collected, err := syncer.CollectFiles(projects)
		if err != nil {
			t.Fatalf("CollectFiles failed: %v", err)
		}

		// Verify collection
		if len(collected) != 5 {
			t.Errorf("Expected 5 files, got %d", len(collected))
		}

		// Verify each file has correct project and priority
		fileMap := make(map[string]syncer.FileInfo)
		for _, file := range collected {
			fileMap[file.RelPath] = file
		}

		// Check specific files
		if file, ok := fileMap["prompts/coding.md"]; ok {
			// Should have 2 instances (from project1 and project2)
			count := 0
			for _, f := range collected {
				if f.RelPath == "prompts/coding.md" {
					count++
				}
			}
			if count != 2 {
				t.Errorf("Expected 2 instances of prompts/coding.md, got %d", count)
			}
			_ = file // Use the variable to avoid unused error
		}

		if file, ok := fileMap["config.json"]; !ok {
			t.Error("config.json should be collected from project1")
		} else {
			if file.Project != "project1" {
				t.Errorf("config.json should be from project1, got %s", file.Project)
			}
			if file.Priority != 1 {
				t.Errorf("config.json should have priority 1, got %d", file.Priority)
			}
		}
	})

	t.Run("resolve conflicts based on priority", func(t *testing.T) {
		// Create test files with conflicts
		files := []syncer.FileInfo{
			{
				RelPath:  "prompts/coding.md",
				AbsPath:  filepath.Join(project1, "prompts", "coding.md"),
				Project:  "project1",
				Priority: 1,
			},
			{
				RelPath:  "prompts/coding.md",
				AbsPath:  filepath.Join(project2, "prompts", "coding.md"),
				Project:  "project2",
				Priority: 2,
			},
			{
				RelPath:  "config.json",
				AbsPath:  filepath.Join(project1, "config.json"),
				Project:  "project1",
				Priority: 1,
			},
		}

		resolved, conflicts, err := syncer.ResolveConflicts(files)
		if err != nil {
			t.Fatalf("ResolveConflicts failed: %v", err)
		}

		// Should have 2 unique files after resolution
		if len(resolved) != 2 {
			t.Errorf("Expected 2 resolved files, got %d", len(resolved))
		}

		// Should have 1 conflict (prompts/coding.md)
		if len(conflicts) != 1 {
			t.Errorf("Expected 1 conflict, got %d", len(conflicts))
		}

		if len(conflicts) > 0 {
			conflict := conflicts[0]
			if conflict.RelPath != "prompts/coding.md" {
				t.Errorf("Expected conflict for prompts/coding.md, got %s", conflict.RelPath)
			}

			// Winner should be from project1 (priority 1)
			if conflict.Resolved.Project != "project1" {
				t.Errorf("Expected winner from project1, got %s", conflict.Resolved.Project)
			}
		}
	})

	t.Run("sync files to all projects", func(t *testing.T) {
		// Create source file
		sourceFile := filepath.Join(project1, "prompts", "new.md")
		sourceContent := "new prompt content"
		if err := os.WriteFile(sourceFile, []byte(sourceContent), 0644); err != nil {
			t.Fatalf("Failed to create source file: %v", err)
		}

		// Verify source file was created correctly
		if !utils.FileExists(sourceFile) {
			t.Fatalf("Source file does not exist: %s", sourceFile)
		}

		verifyContent, err := os.ReadFile(sourceFile)
		if err != nil {
			t.Fatalf("Failed to read source file: %v", err)
		}
		if string(verifyContent) != sourceContent {
			t.Fatalf("Source file content mismatch: got %q, expected %q", string(verifyContent), sourceContent)
		}

		// Create resolved files
		resolved := []syncer.ResolvedFile{
			{
				RelPath:  "prompts/new.md",
				AbsPath:  sourceFile,
				Source:   "project1",
				Priority: 1,
			},
		}

		projects := []config.ProjectPath{
			{Alias: "project1", Path: project1, Priority: 1},
			{Alias: "project2", Path: project2, Priority: 2},
			{Alias: "project3", Path: project3, Priority: 3},
		}

		// Sync files
		results, err := syncer.SyncFiles(resolved, projects, false, true)
		if err != nil {
			t.Fatalf("SyncFiles failed: %v", err)
		}

		// Verify results and check for errors
		if len(results) != 3 {
			t.Errorf("Expected 3 results, got %d", len(results))
		}

		for i, result := range results {
			if len(result.Errors) > 0 {
				t.Errorf("Project %s had errors:", result.Project)
				for _, err := range result.Errors {
					t.Errorf("  - %v", err)
				}
			}
			t.Logf("Result[%d] %s: New=%d, Overwritten=%d, Failed=%d",
				i, result.Project, result.NewFiles, result.Overwritten, result.Failed)
		}

		// Verify files exist in all projects
		for i, proj := range projects {
			destFile := filepath.Join(proj.Path, "prompts", "new.md")
			if !utils.FileExists(destFile) {
				t.Errorf("File should exist in %s", proj.Alias)
				continue
			}

			// Verify content
			content, err := os.ReadFile(destFile)
			if err != nil {
				t.Errorf("Failed to read file from %s: %v", proj.Alias, err)
				continue
			}

			if string(content) != sourceContent {
				t.Errorf("Content mismatch in %s (result[%d]): got %q, expected %q",
					proj.Alias, i, string(content), sourceContent)
			}
		}
	})

	t.Run("handle overwriting existing files", func(t *testing.T) {
		// Create existing file in project2
		existingFile := filepath.Join(project2, "prompts", "existing.md")
		oldContent := "old content"
		if err := os.WriteFile(existingFile, []byte(oldContent), 0644); err != nil {
			t.Fatalf("Failed to create existing file: %v", err)
		}

		// Create new source file in project1
		sourceFile := filepath.Join(project1, "prompts", "existing.md")
		newContent := "new content from project1"
		if err := os.WriteFile(sourceFile, []byte(newContent), 0644); err != nil {
			t.Fatalf("Failed to create source file: %v", err)
		}

		// Sync files
		resolved := []syncer.ResolvedFile{
			{
				RelPath:  "prompts/existing.md",
				AbsPath:  sourceFile,
				Source:   "project1",
				Priority: 1,
			},
		}

		projects := []config.ProjectPath{
			{Alias: "project1", Path: project1, Priority: 1},
			{Alias: "project2", Path: project2, Priority: 2},
		}

		_, err := syncer.SyncFiles(resolved, projects, false, false)
		if err != nil {
			t.Fatalf("SyncFiles failed: %v", err)
		}

		// Verify file in project2 is overwritten with new content
		content, err := os.ReadFile(existingFile)
		if err != nil {
			t.Fatalf("Failed to read overwritten file: %v", err)
		}

		if string(content) != newContent {
			t.Errorf("File should be overwritten with new content: got %q, expected %q",
				string(content), newContent)
		}
	})

	t.Run("dry-run mode does not modify files", func(t *testing.T) {
		// Create source file
		sourceFile := filepath.Join(project1, "prompts", "dryrun.md")
		sourceContent := "dry run test"
		if err := os.WriteFile(sourceFile, []byte(sourceContent), 0644); err != nil {
			t.Fatalf("Failed to create source file: %v", err)
		}

		// Prepare resolved files
		resolved := []syncer.ResolvedFile{
			{
				RelPath:  "prompts/dryrun.md",
				AbsPath:  sourceFile,
				Source:   "project1",
				Priority: 1,
			},
		}

		projects := []config.ProjectPath{
			{Alias: "project1", Path: project1, Priority: 1},
			{Alias: "project2", Path: project2, Priority: 2},
		}

		// Run in dry-run mode
		_, err := syncer.SyncFiles(resolved, projects, true, false)
		if err != nil {
			t.Fatalf("SyncFiles in dry-run mode failed: %v", err)
		}

		// Verify file does NOT exist in project2 (dry-run should not copy)
		destFile := filepath.Join(project2, "prompts", "dryrun.md")
		if utils.FileExists(destFile) {
			t.Error("File should not exist in project2 in dry-run mode")
		}
	})

	t.Run("copy directory structure", func(t *testing.T) {
		// Create nested directory structure
		nestedDir := filepath.Join(project1, "prompts", "nested", "deep")
		if err := os.MkdirAll(nestedDir, 0755); err != nil {
			t.Fatalf("Failed to create nested directory: %v", err)
		}

		// Create file in nested directory
		nestedFile := filepath.Join(nestedDir, "file.md")
		nestedContent := "nested content"
		if err := os.WriteFile(nestedFile, []byte(nestedContent), 0644); err != nil {
			t.Fatalf("Failed to create nested file: %v", err)
		}

		// Collect and sync
		projects := []config.ProjectPath{
			{Alias: "project1", Path: project1, Priority: 1},
			{Alias: "project2", Path: project2, Priority: 2},
		}

		collected, err := syncer.CollectFiles(projects)
		if err != nil {
			t.Fatalf("CollectFiles failed: %v", err)
		}

		// Find the nested file
		var nestedFileInfo *syncer.FileInfo
		for i, file := range collected {
			if file.RelPath == "prompts/nested/deep/file.md" {
				nestedFileInfo = &collected[i]
				break
			}
		}

		if nestedFileInfo == nil {
			t.Fatal("Nested file should be collected")
		}

		// Resolve and sync
		resolved := []syncer.ResolvedFile{
			{
				RelPath:  nestedFileInfo.RelPath,
				AbsPath:  nestedFileInfo.AbsPath,
				Source:   nestedFileInfo.Project,
				Priority: nestedFileInfo.Priority,
			},
		}

		_, err = syncer.SyncFiles(resolved, projects, false, false)
		if err != nil {
			t.Fatalf("SyncFiles failed: %v", err)
		}

		// Verify file exists in project2 with correct structure
		destFile := filepath.Join(project2, "prompts", "nested", "deep", "file.md")
		if !utils.FileExists(destFile) {
			t.Error("Nested file should exist in project2")
		}

		// Verify content
		content, err := os.ReadFile(destFile)
		if err != nil {
			t.Fatalf("Failed to read nested file: %v", err)
		}

		if string(content) != nestedContent {
			t.Errorf("Content mismatch: got %q, expected %q",
				string(content), nestedContent)
		}
	})
}

// TestPushErrorCases tests error handling in push workflow
func TestPushErrorCases(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("no files to collect", func(t *testing.T) {
		// Create empty project directories
		project1 := filepath.Join(tmpDir, "empty1", ".claude")
		if err := os.MkdirAll(project1, 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}

		projects := []config.ProjectPath{
			{Alias: "empty1", Path: project1, Priority: 1},
		}

		// Should return error when no files found
		_, err := syncer.CollectFiles(projects)
		if err == nil {
			t.Error("Expected error when no files to collect, got nil")
		}
	})

	t.Run("handle missing project directory", func(t *testing.T) {
		nonExistentDir := filepath.Join(tmpDir, "nonexistent", ".claude")

		projects := []config.ProjectPath{
			{Alias: "nonexistent", Path: nonExistentDir, Priority: 1},
		}

		// Should handle gracefully (warning but continue)
		_, err := syncer.CollectFiles(projects)
		if err == nil {
			t.Error("Expected error for non-existent directory")
		}
	})
}
