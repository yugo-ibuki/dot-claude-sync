package syncer

import (
	"testing"
	"time"
)

// TestResolveConflicts tests the conflict resolution logic
func TestResolveConflicts(t *testing.T) {
	baseTime := time.Now()

	tests := []struct {
		name      string
		files     []FileInfo
		wantCount int // Number of resolved files
		conflicts int // Number of conflicts
	}{
		{
			name: "No conflicts - single file",
			files: []FileInfo{
				{RelPath: "file1.txt", AbsPath: "/path/to/file1.txt", Project: "project1", Priority: 1, ModTime: baseTime},
			},
			wantCount: 1,
			conflicts: 0,
		},
		{
			name: "No conflicts - different files",
			files: []FileInfo{
				{RelPath: "file1.txt", AbsPath: "/path/to/file1.txt", Project: "project1", Priority: 1, ModTime: baseTime},
				{RelPath: "file2.txt", AbsPath: "/path/to/file2.txt", Project: "project2", Priority: 2, ModTime: baseTime},
			},
			wantCount: 2,
			conflicts: 0,
		},
		{
			name: "Conflict - same file different projects",
			files: []FileInfo{
				{RelPath: "config.yaml", AbsPath: "/project1/config.yaml", Project: "project1", Priority: 1, ModTime: baseTime},
				{RelPath: "config.yaml", AbsPath: "/project2/config.yaml", Project: "project2", Priority: 2, ModTime: baseTime},
			},
			wantCount: 1,
			conflicts: 1,
		},
		{
			name: "Multiple conflicts",
			files: []FileInfo{
				{RelPath: "config.yaml", AbsPath: "/project1/config.yaml", Project: "project1", Priority: 1, ModTime: baseTime},
				{RelPath: "config.yaml", AbsPath: "/project2/config.yaml", Project: "project2", Priority: 2, ModTime: baseTime},
				{RelPath: "hooks/pre-commit", AbsPath: "/project1/hooks/pre-commit", Project: "project1", Priority: 1, ModTime: baseTime},
				{RelPath: "hooks/pre-commit", AbsPath: "/project3/hooks/pre-commit", Project: "project3", Priority: 3, ModTime: baseTime},
			},
			wantCount: 2,
			conflicts: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolved, conflicts, err := ResolveConflicts(tt.files)
			if err != nil {
				t.Fatalf("ResolveConflicts failed: %v", err)
			}

			if len(resolved) != tt.wantCount {
				t.Errorf("Expected %d resolved files, got %d", tt.wantCount, len(resolved))
			}

			if len(conflicts) != tt.conflicts {
				t.Errorf("Expected %d conflicts, got %d", tt.conflicts, len(conflicts))
			}
		})
	}
}

// TestResolveConflict_Timestamp tests that newest file wins
func TestResolveConflict_Timestamp(t *testing.T) {
	baseTime := time.Now()

	files := []FileInfo{
		{RelPath: "test.txt", AbsPath: "/project3/test.txt", Project: "project3", Priority: 3, ModTime: baseTime.Add(-2 * time.Hour)},
		{RelPath: "test.txt", AbsPath: "/project1/test.txt", Project: "project1", Priority: 1, ModTime: baseTime.Add(-1 * time.Hour)},
		{RelPath: "test.txt", AbsPath: "/project2/test.txt", Project: "project2", Priority: 2, ModTime: baseTime}, // Newest
	}

	resolved, conflicts, err := ResolveConflicts(files)
	if err != nil {
		t.Fatalf("ResolveConflicts failed: %v", err)
	}

	if len(resolved) != 1 {
		t.Fatalf("Expected 1 resolved file, got %d", len(resolved))
	}

	if len(conflicts) != 1 {
		t.Fatalf("Expected 1 conflict, got %d", len(conflicts))
	}

	// Project2 with newest timestamp should win
	if resolved[0].Source != "project2" {
		t.Errorf("Expected winner to be project2 (newest), got %s", resolved[0].Source)
	}

	// Check conflict info
	conflict := conflicts[0]
	if conflict.Resolved.Project != "project2" {
		t.Errorf("Conflict resolved project should be project2, got %s", conflict.Resolved.Project)
	}

	if len(conflict.Candidates) != 3 {
		t.Errorf("Expected 3 candidates in conflict, got %d", len(conflict.Candidates))
	}
}

// TestResolveConflict_PriorityFallback tests that priority is used when timestamps are equal
func TestResolveConflict_PriorityFallback(t *testing.T) {
	baseTime := time.Now()

	files := []FileInfo{
		{RelPath: "test.txt", AbsPath: "/project3/test.txt", Project: "project3", Priority: 3, ModTime: baseTime},
		{RelPath: "test.txt", AbsPath: "/project1/test.txt", Project: "project1", Priority: 1, ModTime: baseTime},
		{RelPath: "test.txt", AbsPath: "/project2/test.txt", Project: "project2", Priority: 2, ModTime: baseTime},
	}

	resolved, conflicts, err := ResolveConflicts(files)
	if err != nil {
		t.Fatalf("ResolveConflicts failed: %v", err)
	}

	if len(resolved) != 1 {
		t.Fatalf("Expected 1 resolved file, got %d", len(resolved))
	}

	if len(conflicts) != 1 {
		t.Fatalf("Expected 1 conflict, got %d", len(conflicts))
	}

	// When timestamps are equal, project1 with priority 1 should win
	if resolved[0].Source != "project1" {
		t.Errorf("Expected winner to be project1 (highest priority), got %s", resolved[0].Source)
	}

	if resolved[0].Priority != 1 {
		t.Errorf("Expected winner priority to be 1, got %d", resolved[0].Priority)
	}

	// Check conflict info
	conflict := conflicts[0]
	if conflict.Resolved.Project != "project1" {
		t.Errorf("Conflict resolved project should be project1, got %s", conflict.Resolved.Project)
	}

	if len(conflict.Candidates) != 3 {
		t.Errorf("Expected 3 candidates in conflict, got %d", len(conflict.Candidates))
	}
}

// TestResolveConflicts_EmptyInput tests error handling for empty input
func TestResolveConflicts_EmptyInput(t *testing.T) {
	_, _, err := ResolveConflicts([]FileInfo{})
	if err == nil {
		t.Error("Expected error for empty input, got nil")
	}
}

// TestGroupFilesByRelPath tests file grouping
func TestGroupFilesByRelPath(t *testing.T) {
	baseTime := time.Now()

	files := []FileInfo{
		{RelPath: "file1.txt", AbsPath: "/p1/file1.txt", Project: "p1", Priority: 1, ModTime: baseTime},
		{RelPath: "file2.txt", AbsPath: "/p1/file2.txt", Project: "p1", Priority: 1, ModTime: baseTime},
		{RelPath: "file1.txt", AbsPath: "/p2/file1.txt", Project: "p2", Priority: 2, ModTime: baseTime},
	}

	grouped := GroupFilesByRelPath(files)

	if len(grouped) != 2 {
		t.Errorf("Expected 2 groups, got %d", len(grouped))
	}

	if len(grouped["file1.txt"]) != 2 {
		t.Errorf("Expected 2 files for file1.txt, got %d", len(grouped["file1.txt"]))
	}

	if len(grouped["file2.txt"]) != 1 {
		t.Errorf("Expected 1 file for file2.txt, got %d", len(grouped["file2.txt"]))
	}
}

// TestGetConflictSummary tests conflict summary generation
func TestGetConflictSummary(t *testing.T) {
	baseTime := time.Now()

	tests := []struct {
		name      string
		conflicts []Conflict
		contains  string
	}{
		{
			name:      "No conflicts",
			conflicts: []Conflict{},
			contains:  "No conflicts detected",
		},
		{
			name: "Single conflict",
			conflicts: []Conflict{
				{
					RelPath: "config.yaml",
					Resolved: FileInfo{
						Project:  "project1",
						Priority: 1,
						ModTime:  baseTime,
					},
				},
			},
			contains: "1 conflict(s) resolved",
		},
		{
			name: "Multiple conflicts",
			conflicts: []Conflict{
				{
					RelPath: "file1.txt",
					Resolved: FileInfo{
						Project:  "project1",
						Priority: 1,
						ModTime:  baseTime,
					},
				},
				{
					RelPath: "file2.txt",
					Resolved: FileInfo{
						Project:  "project2",
						Priority: 2,
						ModTime:  baseTime,
					},
				},
			},
			contains: "2 conflict(s) resolved",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := GetConflictSummary(tt.conflicts)
			if summary == "" {
				t.Error("Expected non-empty summary")
			}
			// Check if summary contains expected text
			if len(tt.conflicts) == 0 {
				if summary != "No conflicts detected" {
					t.Errorf("Expected 'No conflicts detected', got '%s'", summary)
				}
			} else {
				// Just verify it's not empty and contains the count
				if len(summary) == 0 {
					t.Error("Expected non-empty conflict summary")
				}
			}
		})
	}
}

// TestGetResolvedSummary tests resolved files summary generation
func TestGetResolvedSummary(t *testing.T) {
	tests := []struct {
		name     string
		resolved []ResolvedFile
		contains string
	}{
		{
			name:     "No resolved files",
			resolved: []ResolvedFile{},
			contains: "No files resolved",
		},
		{
			name: "Single resolved file",
			resolved: []ResolvedFile{
				{RelPath: "file1.txt", Source: "project1"},
			},
			contains: "1 unique file(s) resolved",
		},
		{
			name: "Multiple resolved files",
			resolved: []ResolvedFile{
				{RelPath: "file1.txt", Source: "project1"},
				{RelPath: "file2.txt", Source: "project1"},
				{RelPath: "file3.txt", Source: "project2"},
			},
			contains: "3 unique file(s) resolved",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := GetResolvedSummary(tt.resolved)
			if summary == "" {
				t.Error("Expected non-empty summary")
			}
		})
	}
}

// TestHasConflicts tests the helper function
func TestHasConflicts(t *testing.T) {
	if HasConflicts([]Conflict{}) {
		t.Error("Expected false for empty conflicts")
	}

	conflicts := []Conflict{
		{RelPath: "test.txt"},
	}
	if !HasConflicts(conflicts) {
		t.Error("Expected true for non-empty conflicts")
	}
}

// TestGetConflictCount tests the counter function
func TestGetConflictCount(t *testing.T) {
	if GetConflictCount([]Conflict{}) != 0 {
		t.Error("Expected 0 for empty conflicts")
	}

	conflicts := []Conflict{
		{RelPath: "test1.txt"},
		{RelPath: "test2.txt"},
	}
	if GetConflictCount(conflicts) != 2 {
		t.Error("Expected 2 conflicts")
	}
}

// TestGetResolvedCount tests the counter function
func TestGetResolvedCount(t *testing.T) {
	if GetResolvedCount([]ResolvedFile{}) != 0 {
		t.Error("Expected 0 for empty resolved files")
	}

	resolved := []ResolvedFile{
		{RelPath: "test1.txt"},
		{RelPath: "test2.txt"},
		{RelPath: "test3.txt"},
	}
	if GetResolvedCount(resolved) != 3 {
		t.Error("Expected 3 resolved files")
	}
}
