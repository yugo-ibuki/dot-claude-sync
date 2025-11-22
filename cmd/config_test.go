package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yugo-ibuki/dot-claude-sync/config"
	"gopkg.in/yaml.v3"
)

func TestRunConfigAddProject(t *testing.T) {
	// Save original global variables
	origCfgFile := cfgFile
	origDryRun := dryRun
	defer func() {
		cfgFile = origCfgFile
		dryRun = origDryRun
	}()

	t.Run("argument mode adds project successfully", func(t *testing.T) {
		// Setup temp config file
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")

		// Create initial config
		initialConfig := map[string]interface{}{
			"groups": map[string]interface{}{
				"test-group": map[string]interface{}{
					"paths": map[string]interface{}{
						"proj1": filepath.Join(tmpDir, "proj1", ".claude"),
					},
				},
			},
		}
		data, _ := yaml.Marshal(initialConfig)
		if err := os.WriteFile(configPath, data, 0600); err != nil {
			t.Fatalf("Failed to create test config: %v", err)
		}

		// Set global config file path
		cfgFile = configPath
		dryRun = false

		// Run command with arguments
		args := []string{"test-group", "proj2", filepath.Join(tmpDir, "proj2", ".claude")}
		if err := runConfigAddProject(nil, args); err != nil {
			t.Errorf("runConfigAddProject failed: %v", err)
		}

		// Verify project was added
		cfg, err := config.Load(configPath)
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		group, err := cfg.GetGroup("test-group")
		if err != nil {
			t.Fatalf("Failed to get group: %v", err)
		}

		projects, err := group.GetProjectPaths()
		if err != nil {
			t.Fatalf("Failed to get project paths: %v", err)
		}

		if len(projects) != 2 {
			t.Errorf("Expected 2 projects, got %d", len(projects))
		}

		// Check proj2 exists
		found := false
		for _, proj := range projects {
			if proj.Alias == "proj2" {
				found = true
				expectedPath := filepath.Join(tmpDir, "proj2", ".claude")
				if proj.Path != expectedPath {
					t.Errorf("Expected path %s, got %s", expectedPath, proj.Path)
				}
			}
		}
		if !found {
			t.Error("proj2 not found in projects")
		}
	})

	t.Run("argument mode with dry-run does not modify config", func(t *testing.T) {
		// Setup temp config file
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")

		// Create initial config
		initialConfig := map[string]interface{}{
			"groups": map[string]interface{}{
				"test-group": map[string]interface{}{
					"paths": map[string]interface{}{
						"proj1": filepath.Join(tmpDir, "proj1", ".claude"),
					},
				},
			},
		}
		data, _ := yaml.Marshal(initialConfig)
		if err := os.WriteFile(configPath, data, 0600); err != nil {
			t.Fatalf("Failed to create test config: %v", err)
		}

		// Set global config file path
		cfgFile = configPath
		dryRun = true

		// Run command with arguments
		args := []string{"test-group", "proj2", filepath.Join(tmpDir, "proj2", ".claude")}
		if err := runConfigAddProject(nil, args); err != nil {
			t.Errorf("runConfigAddProject failed: %v", err)
		}

		// Verify project was NOT added
		cfg, err := config.Load(configPath)
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		group, err := cfg.GetGroup("test-group")
		if err != nil {
			t.Fatalf("Failed to get group: %v", err)
		}

		projects, err := group.GetProjectPaths()
		if err != nil {
			t.Fatalf("Failed to get project paths: %v", err)
		}

		if len(projects) != 1 {
			t.Errorf("Expected 1 project (no change), got %d", len(projects))
		}
	})

	t.Run("argument mode with invalid argument count returns error", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")

		// Create minimal config
		initialConfig := map[string]interface{}{
			"groups": map[string]interface{}{},
		}
		data, _ := yaml.Marshal(initialConfig)
		if err := os.WriteFile(configPath, data, 0600); err != nil {
			t.Fatalf("Failed to create test config: %v", err)
		}

		cfgFile = configPath
		dryRun = false

		// Test with 1 argument (should error)
		args := []string{"test-group"}
		if err := runConfigAddProject(nil, args); err == nil {
			t.Error("Expected error with 1 argument, got nil")
		}

		// Test with 2 arguments (should error)
		args = []string{"test-group", "alias"}
		if err := runConfigAddProject(nil, args); err == nil {
			t.Error("Expected error with 2 arguments, got nil")
		}
	})
}

func TestRunConfigAddGroup(t *testing.T) {
	// Save original global variables
	origCfgFile := cfgFile
	defer func() {
		cfgFile = origCfgFile
	}()

	t.Run("add new group successfully", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")

		// Create initial config
		initialConfig := map[string]interface{}{
			"groups": map[string]interface{}{},
		}
		data, _ := yaml.Marshal(initialConfig)
		if err := os.WriteFile(configPath, data, 0600); err != nil {
			t.Fatalf("Failed to create test config: %v", err)
		}

		cfgFile = configPath

		args := []string{"new-group"}
		if err := runConfigAddGroup(nil, args); err != nil {
			t.Errorf("runConfigAddGroup failed: %v", err)
		}

		// Verify group was added
		cfg, err := config.Load(configPath)
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		if _, err := cfg.GetGroup("new-group"); err != nil {
			t.Error("new-group should exist")
		}
	})
}

func TestRunConfigRemoveGroup(t *testing.T) {
	// Save original global variables
	origCfgFile := cfgFile
	origForce := force
	defer func() {
		cfgFile = origCfgFile
		force = origForce
	}()

	t.Run("remove group successfully with force flag", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")

		// Create initial config with a group
		initialConfig := map[string]interface{}{
			"groups": map[string]interface{}{
				"test-group": map[string]interface{}{
					"paths": map[string]interface{}{
						"proj1": "/path/to/.claude",
					},
				},
			},
		}
		data, _ := yaml.Marshal(initialConfig)
		if err := os.WriteFile(configPath, data, 0600); err != nil {
			t.Fatalf("Failed to create test config: %v", err)
		}

		cfgFile = configPath
		force = true

		args := []string{"test-group"}
		if err := runConfigRemoveGroup(nil, args); err != nil {
			t.Errorf("runConfigRemoveGroup failed: %v", err)
		}

		// Verify group was removed
		cfg, err := config.Load(configPath)
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		if _, err := cfg.GetGroup("test-group"); err == nil {
			t.Error("test-group should have been removed")
		}
	})
}

func TestRunConfigRemoveProject(t *testing.T) {
	// Save original global variables
	origCfgFile := cfgFile
	defer func() {
		cfgFile = origCfgFile
	}()

	t.Run("remove project successfully", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")

		// Create initial config with projects
		initialConfig := map[string]interface{}{
			"groups": map[string]interface{}{
				"test-group": map[string]interface{}{
					"paths": map[string]interface{}{
						"proj1": "/path/to/proj1/.claude",
						"proj2": "/path/to/proj2/.claude",
					},
				},
			},
		}
		data, _ := yaml.Marshal(initialConfig)
		if err := os.WriteFile(configPath, data, 0600); err != nil {
			t.Fatalf("Failed to create test config: %v", err)
		}

		cfgFile = configPath

		args := []string{"test-group", "proj1"}
		if err := runConfigRemoveProject(nil, args); err != nil {
			t.Errorf("runConfigRemoveProject failed: %v", err)
		}

		// Verify project was removed
		cfg, err := config.Load(configPath)
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		group, err := cfg.GetGroup("test-group")
		if err != nil {
			t.Fatalf("Failed to get group: %v", err)
		}

		projects, err := group.GetProjectPaths()
		if err != nil {
			t.Fatalf("Failed to get project paths: %v", err)
		}

		if len(projects) != 1 {
			t.Errorf("Expected 1 project, got %d", len(projects))
		}

		if projects[0].Alias == "proj1" {
			t.Error("proj1 should have been removed")
		}
	})
}
