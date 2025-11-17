package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/yugo-ibuki/dot-claude-sync/config"
)

// setupGitRepo creates a git repository with worktrees for testing
func setupGitRepo(t *testing.T, tmpDir string) (mainRepo string, worktrees []string) {
	t.Helper()

	// Create main repository
	mainRepo = filepath.Join(tmpDir, "main-repo")
	if err := os.MkdirAll(mainRepo, 0755); err != nil {
		t.Fatalf("Failed to create main repo: %v", err)
	}

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = mainRepo
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to init git repo: %v, output: %s", err, output)
	}

	// Configure git user (required for commits)
	configCmds := [][]string{
		{"config", "user.email", "test@example.com"},
		{"config", "user.name", "Test User"},
		{"config", "commit.gpgsign", "false"},
	}
	for _, args := range configCmds {
		cmd := exec.Command("git", args...)
		cmd.Dir = mainRepo
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("Failed to configure git: %v, output: %s", err, output)
		}
	}

	// Create initial commit
	readmeFile := filepath.Join(mainRepo, "README.md")
	if err := os.WriteFile(readmeFile, []byte("# Test Repo"), 0644); err != nil {
		t.Fatalf("Failed to create README: %v", err)
	}

	cmd = exec.Command("git", "add", "README.md")
	cmd.Dir = mainRepo
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to add README: %v, output: %s", err, output)
	}

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = mainRepo
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to commit: %v, output: %s", err, output)
	}

	// Create worktrees
	worktreeNames := []string{"feature-1", "feature-2"}
	worktrees = make([]string, len(worktreeNames))

	for i, name := range worktreeNames {
		worktreePath := filepath.Join(tmpDir, name)
		worktrees[i] = worktreePath

		// Create worktree
		cmd := exec.Command("git", "worktree", "add", worktreePath, "-b", name)
		cmd.Dir = mainRepo
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("Failed to create worktree %s: %v, output: %s", name, err, output)
		}
	}

	return mainRepo, worktrees
}

func TestGetWorktreePaths(t *testing.T) {
	tmpDir := t.TempDir()
	mainRepo, expectedWorktrees := setupGitRepo(t, tmpDir)

	// Test getting worktree paths
	paths, err := getWorktreePaths(mainRepo)
	if err != nil {
		t.Fatalf("getWorktreePaths failed: %v", err)
	}

	// Should include main repo + worktrees
	expectedCount := 1 + len(expectedWorktrees)
	if len(paths) != expectedCount {
		t.Errorf("Expected %d worktrees, got %d", expectedCount, len(paths))
	}

	// Verify main repo is included
	found := false
	for _, path := range paths {
		if path == mainRepo {
			found = true
			break
		}
	}
	if !found {
		t.Error("Main repository should be included in worktree list")
	}

	// Verify all worktrees are included
	for _, expected := range expectedWorktrees {
		found := false
		for _, path := range paths {
			if path == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Worktree %s not found in paths", expected)
		}
	}
}

func TestGetWorktreePathsNonGitRepo(t *testing.T) {
	tmpDir := t.TempDir()

	// Try to get worktree paths from non-git directory
	_, err := getWorktreePaths(tmpDir)
	if err == nil {
		t.Error("Expected error for non-git directory")
	}

	if !strings.Contains(err.Error(), "is this a git repository?") {
		t.Errorf("Error message should mention git repository: %v", err)
	}
}

func TestRunDetect(t *testing.T) {
	tmpDir := t.TempDir()
	mainRepo, worktrees := setupGitRepo(t, tmpDir)

	// Create .claude directories in some worktrees
	claudeDir1 := filepath.Join(worktrees[0], ".claude")
	claudeDir2 := filepath.Join(worktrees[1], ".claude")

	for _, dir := range []string{claudeDir1, claudeDir2} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create .claude directory: %v", err)
		}
	}

	// Create config file
	configPath := filepath.Join(tmpDir, "config.yaml")

	tests := []struct {
		name         string
		args         []string
		setupFlags   func()
		setupConfig  func()
		wantErr      bool
		checkResults func(t *testing.T)
	}{
		{
			name: "detect .claude directories in worktrees",
			args: []string{mainRepo},
			setupFlags: func() {
				groupName = "test-group"
				force = true
				dryRun = false
			},
			setupConfig: func() {
				// No existing config
				os.Remove(configPath)
			},
			wantErr: false,
			checkResults: func(t *testing.T) {
				// Verify config was created
				if _, err := os.Stat(configPath); os.IsNotExist(err) {
					t.Error("Config file should be created")
					return
				}

				// Load and verify config
				cfg, err := config.Load(configPath)
				if err != nil {
					t.Fatalf("Failed to load config: %v", err)
				}

				group, err := cfg.GetGroup("test-group")
				if err != nil {
					t.Fatalf("Group 'test-group' should exist: %v", err)
				}

				projects, err := group.GetProjectPaths()
				if err != nil {
					t.Fatalf("Failed to get project paths: %v", err)
				}

				// Should have 2 .claude directories
				if len(projects) != 2 {
					t.Errorf("Expected 2 projects, got %d", len(projects))
				}
			},
		},
		{
			name: "dry-run mode does not modify config",
			args: []string{mainRepo},
			setupFlags: func() {
				groupName = "dry-run-group"
				force = true
				dryRun = true
			},
			setupConfig: func() {
				os.Remove(configPath)
			},
			wantErr: false,
			checkResults: func(t *testing.T) {
				// Config should NOT be created in dry-run mode
				if _, err := os.Stat(configPath); !os.IsNotExist(err) {
					t.Error("Config file should not be created in dry-run mode")
				}
			},
		},
		{
			name: "add to existing group",
			args: []string{mainRepo},
			setupFlags: func() {
				groupName = "existing-group"
				force = true
				dryRun = false
			},
			setupConfig: func() {
				// Create config with existing group
				cfg := &config.Config{
					Groups: map[string]*config.Group{
						"existing-group": {
							Paths: []interface{}{"/existing/path/.claude"},
						},
					},
				}
				data, _ := yaml.Marshal(cfg)
				os.WriteFile(configPath, data, 0644)
			},
			wantErr: false,
			checkResults: func(t *testing.T) {
				cfg, err := config.Load(configPath)
				if err != nil {
					t.Fatalf("Failed to load config: %v", err)
				}

				group, err := cfg.GetGroup("existing-group")
				if err != nil {
					t.Fatalf("Group should exist: %v", err)
				}

				projects, err := group.GetProjectPaths()
				if err != nil {
					t.Fatalf("Failed to get project paths: %v", err)
				}

				// Should have original path + 2 new paths
				if len(projects) < 3 {
					t.Errorf("Expected at least 3 projects (1 existing + 2 detected), got %d", len(projects))
				}
			},
		},
		{
			name: "error when directory does not exist",
			args: []string{filepath.Join(tmpDir, "nonexistent")},
			setupFlags: func() {
				groupName = "test-group"
				force = true
			},
			setupConfig:  func() {},
			wantErr:      true,
			checkResults: func(t *testing.T) {},
		},
		{
			name: "no .claude directories found",
			args: []string{mainRepo},
			setupFlags: func() {
				groupName = "empty-group"
				force = true
			},
			setupConfig: func() {
				// Remove .claude directories
				os.RemoveAll(claudeDir1)
				os.RemoveAll(claudeDir2)
			},
			wantErr: false,
			checkResults: func(t *testing.T) {
				// Should complete without error but not create/modify config
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			tt.setupFlags()
			tt.setupConfig()
			cfgFile = configPath

			// Recreate .claude directories if they were removed in previous test
			for _, dir := range []string{claudeDir1, claudeDir2} {
				os.MkdirAll(dir, 0755)
			}

			// Run command
			err := runDetect(nil, tt.args)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("runDetect() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Check results
			if !tt.wantErr && tt.checkResults != nil {
				tt.checkResults(t)
			}

			// Reset flags
			groupName = ""
			force = false
			dryRun = false
			verbose = false
		})
	}
}

func TestAddPathsToGroup(t *testing.T) {
	tests := []struct {
		name          string
		initialGroup  *config.Group
		pathsToAdd    []string
		expectedCount int
		wantErr       bool
	}{
		{
			name:          "add to new group",
			initialGroup:  nil,
			pathsToAdd:    []string{"/path/1/.claude", "/path/2/.claude"},
			expectedCount: 2,
			wantErr:       false,
		},
		{
			name: "add to existing group with slice paths",
			initialGroup: &config.Group{
				Paths: []interface{}{"/existing/.claude"},
			},
			pathsToAdd:    []string{"/new/1/.claude", "/new/2/.claude"},
			expectedCount: 3,
			wantErr:       false,
		},
		{
			name: "add to existing group with map paths",
			initialGroup: &config.Group{
				Paths: map[string]interface{}{
					"existing": "/existing/.claude",
				},
			},
			pathsToAdd:    []string{"/new/1/.claude"},
			expectedCount: 2,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Groups: make(map[string]*config.Group),
			}

			groupName := "test-group"
			if tt.initialGroup != nil {
				cfg.Groups[groupName] = tt.initialGroup
			}

			err := addPathsToGroup(cfg, groupName, tt.pathsToAdd)

			if (err != nil) != tt.wantErr {
				t.Errorf("addPathsToGroup() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				group := cfg.Groups[groupName]
				if group == nil {
					t.Fatal("Group should exist after adding paths")
				}

				// Get paths based on type
				var pathCount int
				switch paths := group.Paths.(type) {
				case []interface{}:
					pathCount = len(paths)
				case []string:
					pathCount = len(paths)
				case map[string]interface{}:
					pathCount = len(paths)
				}

				if pathCount != tt.expectedCount {
					t.Errorf("Expected %d paths, got %d", tt.expectedCount, pathCount)
				}
			}
		})
	}
}

func TestPluralize(t *testing.T) {
	tests := []struct {
		count int
		want  string
	}{
		{count: 0, want: "ies"},
		{count: 1, want: "y"},
		{count: 2, want: "ies"},
		{count: 10, want: "ies"},
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.count+'0')), func(t *testing.T) {
			got := pluralize(tt.count)
			if got != tt.want {
				t.Errorf("pluralize(%d) = %q, want %q", tt.count, got, tt.want)
			}
		})
	}
}

func TestSaveConfig(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Groups: map[string]*config.Group{
			"test-group": {
				Paths: []interface{}{
					"/path/1/.claude",
					"/path/2/.claude",
				},
			},
		},
	}

	configPath := filepath.Join(tmpDir, "test-config.yaml")
	cfgFile = configPath

	err := saveConfig(cfg)
	if err != nil {
		t.Fatalf("saveConfig() failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file should be created")
	}

	// Verify content can be loaded back
	loadedCfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	if len(loadedCfg.Groups) != 1 {
		t.Errorf("Expected 1 group, got %d", len(loadedCfg.Groups))
	}

	if _, exists := loadedCfg.Groups["test-group"]; !exists {
		t.Error("test-group should exist in loaded config")
	}
}

func TestLoadOrCreateConfig(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("load existing config", func(t *testing.T) {
		// Create existing config
		configContent := `groups:
  test-group:
    paths:
      - /path/.claude
`
		configPath := filepath.Join(tmpDir, "existing.yaml")
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to create config: %v", err)
		}

		cfgFile = configPath
		cfg, err := loadOrCreateConfig()
		if err != nil {
			t.Fatalf("loadOrCreateConfig() failed: %v", err)
		}

		if len(cfg.Groups) != 1 {
			t.Errorf("Expected 1 group, got %d", len(cfg.Groups))
		}
	})

	t.Run("create new config when not found", func(t *testing.T) {
		configPath := filepath.Join(tmpDir, "nonexistent.yaml")
		cfgFile = configPath

		cfg, err := loadOrCreateConfig()
		if err != nil {
			t.Fatalf("loadOrCreateConfig() should create new config: %v", err)
		}

		if cfg == nil {
			t.Fatal("Config should not be nil")
		}

		if cfg.Groups == nil {
			t.Error("Groups should be initialized")
		}

		if len(cfg.Groups) != 0 {
			t.Errorf("New config should have 0 groups, got %d", len(cfg.Groups))
		}
	})
}
