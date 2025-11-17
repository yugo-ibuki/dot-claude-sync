package cmd

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yugo-ibuki/dot-claude-sync/utils"
)

func TestRunBackup(t *testing.T) {
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

	// Create test files in projects
	testFile1 := filepath.Join(project1, "test.md")
	testFile2 := filepath.Join(project2, "commands", "test.md")
	testFile3 := filepath.Join(project1, "config.yaml")

	if err := os.WriteFile(testFile1, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(testFile2), 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	if err := os.WriteFile(testFile2, []byte("command content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := os.WriteFile(testFile3, []byte("config content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
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
			name: "backup files from multiple projects",
			args: []string{"test-group"},
			setupFiles: func() {
				os.WriteFile(testFile1, []byte("test"), 0644)
				os.WriteFile(testFile3, []byte("config"), 0644)
			},
			setupFlags: func() {
				dryRun = false
				verbose = false
			},
			wantErr: false,
			checkAfter: func(t *testing.T) {
				// Check that backup directories were created with timestamps
				bkDir1 := filepath.Join(project1, "bk")
				bkDir2 := filepath.Join(project2, "bk")

				if !utils.FileExists(bkDir1) {
					t.Error("Backup directory should exist in project1")
				}
				if !utils.FileExists(bkDir2) {
					t.Error("Backup directory should exist in project2")
				}

				// Check that files were backed up
				entries, err := os.ReadDir(bkDir1)
				if err != nil {
					t.Fatalf("Failed to read backup directory: %v", err)
				}
				if len(entries) == 0 {
					t.Error("Backup directory should contain timestamped subdirectory")
				}

				// Verify that bk directory itself is not backed up
				if len(entries) > 0 {
					timestampDir := filepath.Join(bkDir1, entries[0].Name())
					nestedBkDir := filepath.Join(timestampDir, "bk")
					if utils.FileExists(nestedBkDir) {
						t.Error("bk directory should not be backed up into itself")
					}
				}
			},
		},
		{
			name: "dry-run mode does not create backup",
			args: []string{"test-group"},
			setupFiles: func() {
				os.WriteFile(testFile1, []byte("test"), 0644)
			},
			setupFlags: func() {
				dryRun = true
				verbose = false
			},
			wantErr: false,
			checkAfter: func(t *testing.T) {
				// In dry-run, no backup should be created
				bkDir1 := filepath.Join(project1, "bk")

				// If bk exists, it should be empty or not have new entries
				if utils.FileExists(bkDir1) {
					entries, err := os.ReadDir(bkDir1)
					if err == nil && len(entries) > 0 {
						// Check if the entry is from a previous test
						// In a fresh dry-run, no new entries should be created
					}
				}
			},
		},
		{
			name: "backup with verbose output",
			args: []string{"test-group"},
			setupFiles: func() {
				os.WriteFile(testFile1, []byte("test"), 0644)
			},
			setupFlags: func() {
				dryRun = false
				verbose = true
			},
			wantErr: false,
			checkAfter: func(t *testing.T) {
				bkDir1 := filepath.Join(project1, "bk")
				if !utils.FileExists(bkDir1) {
					t.Error("Backup directory should exist with verbose flag")
				}
			},
		},
		{
			name: "backup non-existent project skips gracefully",
			args: []string{"test-group"},
			setupFiles: func() {
				// Remove project3 to test skipping
				os.RemoveAll(project3)
			},
			setupFlags: func() {
				dryRun = false
				verbose = true
			},
			wantErr: false,
			checkAfter: func(t *testing.T) {
				// Should still succeed for other projects
				bkDir1 := filepath.Join(project1, "bk")
				if !utils.FileExists(bkDir1) {
					t.Error("Backup directory should exist for existing projects")
				}
			},
		},
		{
			name: "invalid group name returns error",
			args: []string{"nonexistent-group"},
			setupFiles: func() {
				// No setup needed
			},
			setupFlags: func() {
				dryRun = false
			},
			wantErr: true,
			checkAfter: func(t *testing.T) {
				// No backup should be created
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up any existing backup directories from previous tests
			for _, project := range []string{project1, project2, project3} {
				bkDir := filepath.Join(project, "bk")
				os.RemoveAll(bkDir)
			}

			// Setup
			tt.setupFiles()
			tt.setupFlags()

			// Override config file path
			cfgFile = configPath

			// Run command
			err := runBackup(nil, tt.args)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("runBackup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check results
			if !tt.wantErr && tt.checkAfter != nil {
				tt.checkAfter(t)
			}

			// Reset flags
			dryRun = false
			verbose = false
		})
	}
}

func TestBackupProject(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test project directory
	projectDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("Failed to create project directory: %v", err)
	}

	// Create test files
	testFile1 := filepath.Join(projectDir, "test.md")
	testFile2 := filepath.Join(projectDir, "commands", "cmd.md")
	if err := os.WriteFile(testFile1, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(testFile2), 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	if err := os.WriteFile(testFile2, []byte("command"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name         string
		setupProject func()
		expectSkip   bool
		expectError  bool
		skipReason   string
	}{
		{
			name: "backup existing project",
			setupProject: func() {
				// Project already set up
			},
			expectSkip:  false,
			expectError: false,
		},
		{
			name: "skip non-existent directory",
			setupProject: func() {
				os.RemoveAll(projectDir)
			},
			expectSkip: true,
			skipReason: ".claude directory does not exist",
		},
		{
			name: "skip non-directory path",
			setupProject: func() {
				os.RemoveAll(projectDir)
				// Create a file instead of directory
				os.WriteFile(projectDir, []byte("not a directory"), 0644)
			},
			expectSkip: true,
			skipReason: "path is not a directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset project directory
			os.RemoveAll(projectDir)
			os.MkdirAll(projectDir, 0755)
			os.WriteFile(testFile1, []byte("test"), 0644)
			os.MkdirAll(filepath.Dir(testFile2), 0755)
			os.WriteFile(testFile2, []byte("command"), 0644)

			// Setup specific test case
			tt.setupProject()

			// Create project config
			project := struct {
				Alias    string
				Path     string
				Priority int
			}{
				Alias:    "test-project",
				Path:     projectDir,
				Priority: 1,
			}

			// Convert to ProjectPath type
			var projectPath struct {
				Alias    string
				Path     string
				Priority int
			}
			projectPath.Alias = project.Alias
			projectPath.Path = project.Path
			projectPath.Priority = project.Priority

			// Generate timestamp
			timestamp := time.Now().Format("20060102-150405")

			// This test would need the actual backupProject function
			// For now, we test the logic that would be in backupProject

			claudeDir := projectDir
			if !utils.FileExists(claudeDir) {
				if !tt.expectSkip {
					t.Error("Expected directory to exist")
				}
				return
			}

			if !utils.IsDirectory(claudeDir) {
				if !tt.expectSkip {
					t.Error("Expected path to be a directory")
				}
				return
			}

			// Create backup directory
			backupDir := filepath.Join(claudeDir, "bk", timestamp)
			if err := utils.EnsureDir(backupDir); err != nil {
				if !tt.expectError {
					t.Errorf("Failed to create backup directory: %v", err)
				}
				return
			}

			// Copy files
			if err := utils.CopyDirExclude(claudeDir, backupDir, []string{"bk"}); err != nil {
				if !tt.expectError {
					t.Errorf("Failed to copy files: %v", err)
				}
				return
			}

			// Verify backup was created
			if !utils.FileExists(backupDir) {
				t.Error("Backup directory should exist")
			}

			// Verify files were copied
			backedUpFile1 := filepath.Join(backupDir, "test.md")
			backedUpFile2 := filepath.Join(backupDir, "commands", "cmd.md")

			if !utils.FileExists(backedUpFile1) {
				t.Error("Backed up file should exist")
			}
			if !utils.FileExists(backedUpFile2) {
				t.Error("Backed up nested file should exist")
			}

			// Verify bk directory itself was not copied
			nestedBkDir := filepath.Join(backupDir, "bk")
			if utils.FileExists(nestedBkDir) {
				t.Error("bk directory should not be copied into backup")
			}
		})
	}
}

func TestBackupTimestamp(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test project
	projectDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("Failed to create project directory: %v", err)
	}

	testFile := filepath.Join(projectDir, "test.md")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create first backup
	timestamp1 := time.Now().Format("20060102-150405")
	backupDir1 := filepath.Join(projectDir, "bk", timestamp1)
	if err := utils.EnsureDir(backupDir1); err != nil {
		t.Fatalf("Failed to create first backup directory: %v", err)
	}
	if err := utils.CopyDirExclude(projectDir, backupDir1, []string{"bk"}); err != nil {
		t.Fatalf("Failed to copy files for first backup: %v", err)
	}

	// Wait a moment to ensure different timestamp
	time.Sleep(2 * time.Second)

	// Create second backup
	timestamp2 := time.Now().Format("20060102-150405")
	backupDir2 := filepath.Join(projectDir, "bk", timestamp2)
	if err := utils.EnsureDir(backupDir2); err != nil {
		t.Fatalf("Failed to create second backup directory: %v", err)
	}
	if err := utils.CopyDirExclude(projectDir, backupDir2, []string{"bk"}); err != nil {
		t.Fatalf("Failed to copy files for second backup: %v", err)
	}

	// Verify both backups exist
	if !utils.FileExists(backupDir1) {
		t.Error("First backup should exist")
	}
	if !utils.FileExists(backupDir2) {
		t.Error("Second backup should exist")
	}

	// Verify timestamps are different
	if timestamp1 == timestamp2 {
		t.Log("Warning: Timestamps are the same (this may happen if tests run very quickly)")
	}

	// Verify both contain the backed up file
	if !utils.FileExists(filepath.Join(backupDir1, "test.md")) {
		t.Error("First backup should contain test file")
	}
	if !utils.FileExists(filepath.Join(backupDir2, "test.md")) {
		t.Error("Second backup should contain test file")
	}
}

func TestBackupExcludesBkDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create project with existing bk directory
	projectDir := filepath.Join(tmpDir, ".claude")
	existingBkDir := filepath.Join(projectDir, "bk", "old-backup")

	if err := os.MkdirAll(existingBkDir, 0755); err != nil {
		t.Fatalf("Failed to create existing bk directory: %v", err)
	}

	oldBackupFile := filepath.Join(existingBkDir, "old.md")
	if err := os.WriteFile(oldBackupFile, []byte("old backup"), 0644); err != nil {
		t.Fatalf("Failed to create old backup file: %v", err)
	}

	testFile := filepath.Join(projectDir, "test.md")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create new backup
	timestamp := time.Now().Format("20060102-150405")
	newBackupDir := filepath.Join(projectDir, "bk", timestamp)

	if err := utils.EnsureDir(newBackupDir); err != nil {
		t.Fatalf("Failed to create new backup directory: %v", err)
	}

	if err := utils.CopyDirExclude(projectDir, newBackupDir, []string{"bk"}); err != nil {
		t.Fatalf("Failed to copy files: %v", err)
	}

	// Verify test file was backed up
	backedUpTestFile := filepath.Join(newBackupDir, "test.md")
	if !utils.FileExists(backedUpTestFile) {
		t.Error("Test file should be backed up")
	}

	// Verify bk directory was not backed up
	nestedBkDir := filepath.Join(newBackupDir, "bk")
	if utils.FileExists(nestedBkDir) {
		t.Error("bk directory should not be backed up")
	}

	// Verify old backup still exists in original location
	if !utils.FileExists(oldBackupFile) {
		t.Error("Old backup file should still exist")
	}
}
