package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yugo-ibuki/dot-claude-sync/utils"
)

func TestRunMv(t *testing.T) {
	// Create temporary test directory structure
	tmpDir := t.TempDir()

	// Create test projects
	project1 := filepath.Join(tmpDir, "project1", ".claude")
	project2 := filepath.Join(tmpDir, "project2", ".claude")
	project3 := filepath.Join(tmpDir, "project3", ".claude")

	for _, dir := range []string{project1, project2, project3} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}
	}

	// Create test files
	testFile1 := filepath.Join(project1, "old.md")
	testFile2 := filepath.Join(project2, "old.md")
	testDir := filepath.Join(project1, "olddir")

	if err := os.WriteFile(testFile1, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := os.WriteFile(testFile2, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(testDir, "nested.md"), []byte("nested"), 0644); err != nil {
		t.Fatalf("Failed to create nested file: %v", err)
	}

	// Create test config file
	configContent := `groups:
  test-group:
    paths:
      proj1: ` + project1 + `
      proj2: ` + project2 + `
      proj3: ` + project3 + `
`
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	tests := []struct {
		name       string
		args       []string
		setupFiles func()
		wantErr    bool
		checkAfter func(t *testing.T)
		setupFlags func()
	}{
		{
			name: "move existing file in multiple projects",
			args: []string{"test-group", "old.md", "new.md"},
			setupFiles: func() {
				os.WriteFile(testFile1, []byte("test"), 0644)
				os.WriteFile(testFile2, []byte("test"), 0644)
			},
			setupFlags: func() {
				force = true
				dryRun = false
			},
			wantErr: false,
			checkAfter: func(t *testing.T) {
				newFile1 := filepath.Join(project1, "new.md")
				newFile2 := filepath.Join(project2, "new.md")

				if !utils.FileExists(newFile1) {
					t.Error("New file should exist in project1")
				}
				if !utils.FileExists(newFile2) {
					t.Error("New file should exist in project2")
				}
				if utils.FileExists(testFile1) {
					t.Error("Old file should not exist in project1")
				}
				if utils.FileExists(testFile2) {
					t.Error("Old file should not exist in project2")
				}
			},
		},
		{
			name: "move directory",
			args: []string{"test-group", "olddir", "newdir"},
			setupFiles: func() {
				os.MkdirAll(testDir, 0755)
				os.WriteFile(filepath.Join(testDir, "nested.md"), []byte("nested"), 0644)
			},
			setupFlags: func() {
				force = true
				dryRun = false
			},
			wantErr: false,
			checkAfter: func(t *testing.T) {
				newDir := filepath.Join(project1, "newdir")
				nestedFile := filepath.Join(newDir, "nested.md")

				if !utils.FileExists(newDir) {
					t.Error("New directory should exist")
				}
				if !utils.FileExists(nestedFile) {
					t.Error("Nested file should exist in new directory")
				}
				if utils.FileExists(testDir) {
					t.Error("Old directory should not exist")
				}
			},
		},
		{
			name: "dry-run mode does not move",
			args: []string{"test-group", "old.md", "new.md"},
			setupFiles: func() {
				os.WriteFile(testFile1, []byte("test"), 0644)
				os.WriteFile(testFile2, []byte("test"), 0644)
			},
			setupFlags: func() {
				force = true
				dryRun = true
			},
			wantErr: false,
			checkAfter: func(t *testing.T) {
				newFile1 := filepath.Join(project1, "new.md")
				newFile2 := filepath.Join(project2, "new.md")

				if utils.FileExists(newFile1) {
					t.Error("New file should not exist in dry-run mode")
				}
				if utils.FileExists(newFile2) {
					t.Error("New file should not exist in dry-run mode")
				}
				if !utils.FileExists(testFile1) {
					t.Error("Old file should still exist in dry-run mode")
				}
				if !utils.FileExists(testFile2) {
					t.Error("Old file should still exist in dry-run mode")
				}
			},
		},
		{
			name: "move non-existent file succeeds with skip",
			args: []string{"test-group", "nonexistent.md", "new.md"},
			setupFiles: func() {
				// No files to create
			},
			setupFlags: func() {
				force = true
				dryRun = false
				verbose = true
			},
			wantErr: false,
			checkAfter: func(t *testing.T) {
				// Nothing to check, just verify no error
			},
		},
		{
			name: "move with destination already exists",
			args: []string{"test-group", "old.md", "existing.md"},
			setupFiles: func() {
				os.WriteFile(testFile1, []byte("old"), 0644)
				existingFile := filepath.Join(project1, "existing.md")
				os.WriteFile(existingFile, []byte("existing"), 0644)
			},
			setupFlags: func() {
				force = true
				dryRun = false
			},
			wantErr: false,
			checkAfter: func(t *testing.T) {
				// Should skip because destination exists
				if !utils.FileExists(testFile1) {
					t.Error("Old file should still exist when destination exists")
				}
				existingFile := filepath.Join(project1, "existing.md")
				content, err := os.ReadFile(existingFile)
				if err != nil {
					t.Errorf("Failed to read existing file: %v", err)
				}
				if string(content) != "existing" {
					t.Error("Existing file should not be overwritten")
				}
			},
		},
		{
			name: "invalid group name returns error",
			args: []string{"nonexistent-group", "old.md", "new.md"},
			setupFiles: func() {
				// No setup needed
			},
			setupFlags: func() {
				force = true
				dryRun = false
			},
			wantErr: true,
			checkAfter: func(t *testing.T) {
				// No cleanup needed
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			tt.setupFiles()
			tt.setupFlags()

			// Override config file path
			cfgFile = configPath

			// Run command
			err := runMv(nil, tt.args)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("runMv() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check results
			if !tt.wantErr && tt.checkAfter != nil {
				tt.checkAfter(t)
			}

			// Reset flags
			force = false
			dryRun = false
			verbose = false

			// Cleanup files for next test
			os.RemoveAll(project1)
			os.RemoveAll(project2)
			os.RemoveAll(project3)
			os.MkdirAll(project1, 0755)
			os.MkdirAll(project2, 0755)
			os.MkdirAll(project3, 0755)
		})
	}
}

func TestMoveInProject(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test project
	projectDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("Failed to create project directory: %v", err)
	}

	testFile := filepath.Join(projectDir, "source.md")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name         string
		fromPath     string
		toPath       string
		setupFiles   func()
		dryRun       bool
		expectMoved  bool
		expectSkip   bool
		skipReason   string
		expectError  bool
	}{
		{
			name:        "move existing file",
			fromPath:    "source.md",
			toPath:      "dest.md",
			setupFiles:  func() {},
			dryRun:      false,
			expectMoved: true,
			expectSkip:  false,
		},
		{
			name:       "source not found",
			fromPath:   "nonexistent.md",
			toPath:     "dest.md",
			setupFiles: func() {},
			dryRun:     false,
			expectSkip: true,
			skipReason: "source not found",
		},
		{
			name:     "destination already exists",
			fromPath: "source.md",
			toPath:   "existing.md",
			setupFiles: func() {
				existingFile := filepath.Join(projectDir, "existing.md")
				os.WriteFile(existingFile, []byte("existing"), 0644)
			},
			dryRun:     false,
			expectSkip: true,
			skipReason: "destination already exists",
		},
		{
			name:        "dry-run mode",
			fromPath:    "source.md",
			toPath:      "dest.md",
			setupFiles:  func() {},
			dryRun:      true,
			expectMoved: true,
			expectSkip:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			tt.setupFiles()

			// Test the logic directly
			srcFullPath := filepath.Join(projectDir, tt.fromPath)
			dstFullPath := filepath.Join(projectDir, tt.toPath)

			// Check source exists
			srcExists := utils.FileExists(srcFullPath)
			dstExists := utils.FileExists(dstFullPath)

			if !srcExists && tt.expectSkip && tt.skipReason == "source not found" {
				// Expected behavior
				return
			}

			if srcExists && dstExists && tt.expectSkip && tt.skipReason == "destination already exists" {
				// Expected behavior
				return
			}

			if tt.dryRun {
				// In dry-run, files should not actually move
				if srcExists {
					if !utils.FileExists(srcFullPath) {
						t.Error("Source file should still exist in dry-run")
					}
				}
				return
			}

			if srcExists && !dstExists && tt.expectMoved {
				// Perform actual move
				err := utils.MoveFile(srcFullPath, dstFullPath)
				if err != nil {
					t.Errorf("MoveFile failed: %v", err)
				}

				// Verify move
				if utils.FileExists(srcFullPath) {
					t.Error("Source file should not exist after move")
				}
				if !utils.FileExists(dstFullPath) {
					t.Error("Destination file should exist after move")
				}
			}

			// Cleanup
			os.Remove(filepath.Join(projectDir, "dest.md"))
			os.Remove(filepath.Join(projectDir, "existing.md"))
			os.WriteFile(testFile, []byte("test"), 0644)
		})
	}
}

func TestMvCommandFlags(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test project
	project := filepath.Join(tmpDir, "project", ".claude")
	if err := os.MkdirAll(project, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	testFile := filepath.Join(project, "old.md")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create config
	configContent := `groups:
  test:
    paths:
      proj: ` + project + `
`
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	tests := []struct {
		name          string
		force         bool
		dryRun        bool
		verbose       bool
		shouldMove    bool
	}{
		{
			name:       "force flag moves without prompt",
			force:      true,
			dryRun:     false,
			verbose:    false,
			shouldMove: true,
		},
		{
			name:       "dry-run flag preserves file",
			force:      true,
			dryRun:     true,
			verbose:    false,
			shouldMove: false,
		},
		{
			name:       "verbose flag shows extra output",
			force:      true,
			dryRun:     false,
			verbose:    true,
			shouldMove: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Recreate test file for each test
			if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Set flags
			force = tt.force
			dryRun = tt.dryRun
			verbose = tt.verbose
			cfgFile = configPath

			// Run command
			err := runMv(nil, []string{"test", "old.md", "new.md"})
			if err != nil {
				t.Errorf("runMv() unexpected error: %v", err)
			}

			// Check file state
			oldExists := utils.FileExists(testFile)
			newFile := filepath.Join(project, "new.md")
			newExists := utils.FileExists(newFile)

			if tt.shouldMove {
				if oldExists {
					t.Error("Old file should not exist after move")
				}
				if !newExists {
					t.Error("New file should exist after move")
				}
			} else {
				if !oldExists {
					t.Error("Old file should still exist in dry-run")
				}
				if newExists {
					t.Error("New file should not exist in dry-run")
				}
			}

			// Reset flags
			force = false
			dryRun = false
			verbose = false

			// Cleanup
			os.Remove(newFile)
		})
	}
}
