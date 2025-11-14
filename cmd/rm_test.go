package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yugo-ibuki/dot-claude-sync/utils"
)

func TestRunRm(t *testing.T) {
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
	testFile1 := filepath.Join(project1, "test.md")
	testFile2 := filepath.Join(project2, "test.md")
	testDir := filepath.Join(project1, "testdir")

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
		name           string
		args           []string
		setupFiles     func()
		wantErr        bool
		checkAfter     func(t *testing.T)
		setupFlags     func()
	}{
		{
			name: "delete existing file from multiple projects",
			args: []string{"test-group", "test.md"},
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
				if utils.FileExists(testFile1) {
					t.Error("File should have been deleted from project1")
				}
				if utils.FileExists(testFile2) {
					t.Error("File should have been deleted from project2")
				}
			},
		},
		{
			name: "delete directory recursively",
			args: []string{"test-group", "testdir"},
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
				if utils.FileExists(testDir) {
					t.Error("Directory should have been deleted")
				}
			},
		},
		{
			name: "dry-run mode does not delete",
			args: []string{"test-group", "test.md"},
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
				if !utils.FileExists(testFile1) {
					t.Error("File should still exist in dry-run mode")
				}
				if !utils.FileExists(testFile2) {
					t.Error("File should still exist in dry-run mode")
				}
			},
		},
		{
			name: "delete non-existent file succeeds with skip",
			args: []string{"test-group", "nonexistent.md"},
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
			name: "invalid group name returns error",
			args: []string{"nonexistent-group", "test.md"},
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
			err := runRm(nil, tt.args)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("runRm() error = %v, wantErr %v", err, tt.wantErr)
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
		})
	}
}

func TestDeleteTarget(t *testing.T) {
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.txt")
	testDir := filepath.Join(tmpDir, "testdir")

	// Create test file
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create test directory
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	tests := []struct {
		name     string
		path     string
		exists   bool
		isDir    bool
	}{
		{
			name:   "existing file",
			path:   testFile,
			exists: true,
			isDir:  false,
		},
		{
			name:   "existing directory",
			path:   testDir,
			exists: true,
			isDir:  true,
		},
		{
			name:   "non-existent path",
			path:   filepath.Join(tmpDir, "nonexistent"),
			exists: false,
			isDir:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists := utils.FileExists(tt.path)
			if exists != tt.exists {
				t.Errorf("FileExists() = %v, want %v", exists, tt.exists)
			}

			if exists {
				isDir := utils.IsDirectory(tt.path)
				if isDir != tt.isDir {
					t.Errorf("IsDirectory() = %v, want %v", isDir, tt.isDir)
				}
			}
		})
	}
}

func TestRmCommandFlags(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test project
	project := filepath.Join(tmpDir, "project", ".claude")
	if err := os.MkdirAll(project, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	testFile := filepath.Join(project, "test.md")
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
		name       string
		force      bool
		dryRun     bool
		verbose    bool
		shouldExist bool
	}{
		{
			name:        "force flag deletes without prompt",
			force:       true,
			dryRun:      false,
			verbose:     false,
			shouldExist: false,
		},
		{
			name:        "dry-run flag preserves file",
			force:       true,
			dryRun:      true,
			verbose:     false,
			shouldExist: true,
		},
		{
			name:        "verbose flag shows extra output",
			force:       true,
			dryRun:      false,
			verbose:     true,
			shouldExist: false,
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
			err := runRm(nil, []string{"test", "test.md"})
			if err != nil {
				t.Errorf("runRm() unexpected error: %v", err)
			}

			// Check file existence
			exists := utils.FileExists(testFile)
			if exists != tt.shouldExist {
				t.Errorf("File exists = %v, want %v", exists, tt.shouldExist)
			}

			// Reset flags
			force = false
			dryRun = false
			verbose = false
		})
	}
}
