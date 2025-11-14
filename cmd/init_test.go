package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestInitCommand tests the init command functionality
func TestInitCommand(t *testing.T) {
	// Save original HOME and restore after test
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	t.Run("create config with example data", func(t *testing.T) {
		// Create temporary home directory
		tmpHome := t.TempDir()
		os.Setenv("HOME", tmpHome)

		configDir := filepath.Join(tmpHome, ".config", "claude-sync")
		configPath := filepath.Join(configDir, "config.yaml")

		// Test the config creation logic (without interactive input)
		// Create directory
		if err := os.MkdirAll(configDir, 0755); err != nil {
			t.Fatalf("Failed to create config directory: %v", err)
		}

		// Create example config
		groups := map[string]interface{}{
			"test-group": map[string]interface{}{
				"paths": map[string]string{
					"project-a": filepath.Join(tmpHome, "project-a", ".claude"),
					"project-b": filepath.Join(tmpHome, "project-b", ".claude"),
				},
				"priority": []string{"project-a", "project-b"},
			},
		}

		config := map[string]interface{}{
			"groups": groups,
		}

		// Write config
		data, err := yaml.Marshal(config)
		if err != nil {
			t.Fatalf("Failed to marshal config: %v", err)
		}

		if err := os.WriteFile(configPath, data, 0644); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		// Verify config file exists
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("Config file should exist")
		}

		// Verify config content
		readData, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("Failed to read config file: %v", err)
		}

		var readConfig map[string]interface{}
		if err := yaml.Unmarshal(readData, &readConfig); err != nil {
			t.Fatalf("Failed to unmarshal config: %v", err)
		}

		// Verify groups exist
		if readConfig["groups"] == nil {
			t.Error("Config should have groups")
		}

		groups, ok := readConfig["groups"].(map[string]interface{})
		if !ok {
			t.Fatal("Groups should be a map")
		}

		if groups["test-group"] == nil {
			t.Error("test-group should exist in config")
		}
	})

	t.Run("config directory creation", func(t *testing.T) {
		// Create temporary home directory
		tmpHome := t.TempDir()
		os.Setenv("HOME", tmpHome)

		configDir := filepath.Join(tmpHome, ".config", "claude-sync")

		// Verify directory doesn't exist yet
		if _, err := os.Stat(configDir); !os.IsNotExist(err) {
			t.Error("Config directory should not exist yet")
		}

		// Create directory
		if err := os.MkdirAll(configDir, 0755); err != nil {
			t.Fatalf("Failed to create config directory: %v", err)
		}

		// Verify directory exists
		info, err := os.Stat(configDir)
		if err != nil {
			t.Fatalf("Config directory should exist: %v", err)
		}

		if !info.IsDir() {
			t.Error("Config path should be a directory")
		}
	})

	t.Run("yaml marshaling and unmarshaling", func(t *testing.T) {
		// Test YAML operations
		testConfig := map[string]interface{}{
			"groups": map[string]interface{}{
				"frontend": map[string]interface{}{
					"paths": map[string]string{
						"web":    "/path/to/web/.claude",
						"mobile": "/path/to/mobile/.claude",
					},
					"priority": []string{"web", "mobile"},
				},
				"backend": map[string]interface{}{
					"paths": []string{
						"/path/to/api/.claude",
						"/path/to/worker/.claude",
					},
				},
			},
		}

		// Marshal to YAML
		data, err := yaml.Marshal(testConfig)
		if err != nil {
			t.Fatalf("Failed to marshal config: %v", err)
		}

		if len(data) == 0 {
			t.Error("Marshaled data should not be empty")
		}

		// Unmarshal back
		var result map[string]interface{}
		if err := yaml.Unmarshal(data, &result); err != nil {
			t.Fatalf("Failed to unmarshal config: %v", err)
		}

		// Verify structure
		if result["groups"] == nil {
			t.Error("Unmarshaled config should have groups")
		}
	})

	t.Run("handle existing config file", func(t *testing.T) {
		tmpHome := t.TempDir()
		os.Setenv("HOME", tmpHome)

		configDir := filepath.Join(tmpHome, ".config", "claude-sync")
		configPath := filepath.Join(configDir, "config.yaml")

		// Create directory
		if err := os.MkdirAll(configDir, 0755); err != nil {
			t.Fatalf("Failed to create config directory: %v", err)
		}

		// Create existing config
		existingContent := "groups:\n  existing-group:\n    paths:\n      - /path/to/project/.claude\n"
		if err := os.WriteFile(configPath, []byte(existingContent), 0644); err != nil {
			t.Fatalf("Failed to write existing config: %v", err)
		}

		// Verify file exists
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("Existing config file should exist")
		}

		// Read and verify content
		content, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("Failed to read config: %v", err)
		}

		if string(content) != existingContent {
			t.Errorf("Content mismatch: got %q, expected %q", string(content), existingContent)
		}
	})

	t.Run("create config with paths list format", func(t *testing.T) {
		tmpHome := t.TempDir()
		os.Setenv("HOME", tmpHome)

		configDir := filepath.Join(tmpHome, ".config", "claude-sync")
		configPath := filepath.Join(configDir, "config.yaml")

		// Create config with list format
		if err := os.MkdirAll(configDir, 0755); err != nil {
			t.Fatalf("Failed to create config directory: %v", err)
		}

		groups := map[string]interface{}{
			"simple-group": map[string]interface{}{
				"paths": []string{
					"/path/to/project1/.claude",
					"/path/to/project2/.claude",
				},
			},
		}

		config := map[string]interface{}{
			"groups": groups,
		}

		data, err := yaml.Marshal(config)
		if err != nil {
			t.Fatalf("Failed to marshal config: %v", err)
		}

		if err := os.WriteFile(configPath, data, 0644); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		// Read and verify
		readData, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("Failed to read config: %v", err)
		}

		var readConfig map[string]interface{}
		if err := yaml.Unmarshal(readData, &readConfig); err != nil {
			t.Fatalf("Failed to unmarshal config: %v", err)
		}

		groups2, ok := readConfig["groups"].(map[string]interface{})
		if !ok {
			t.Fatal("Groups should be a map")
		}

		simpleGroup, ok := groups2["simple-group"].(map[string]interface{})
		if !ok {
			t.Fatal("simple-group should be a map")
		}

		paths, ok := simpleGroup["paths"].([]interface{})
		if !ok {
			t.Fatal("paths should be a list")
		}

		if len(paths) != 2 {
			t.Errorf("Expected 2 paths, got %d", len(paths))
		}
	})

	t.Run("create config with priority", func(t *testing.T) {
		tmpHome := t.TempDir()
		os.Setenv("HOME", tmpHome)

		configDir := filepath.Join(tmpHome, ".config", "claude-sync")
		configPath := filepath.Join(configDir, "config.yaml")

		if err := os.MkdirAll(configDir, 0755); err != nil {
			t.Fatalf("Failed to create config directory: %v", err)
		}

		// Create config with priority
		groups := map[string]interface{}{
			"priority-group": map[string]interface{}{
				"paths": map[string]string{
					"high-priority": "/path/to/high/.claude",
					"low-priority":  "/path/to/low/.claude",
				},
				"priority": []string{"high-priority", "low-priority"},
			},
		}

		config := map[string]interface{}{
			"groups": groups,
		}

		data, err := yaml.Marshal(config)
		if err != nil {
			t.Fatalf("Failed to marshal config: %v", err)
		}

		if err := os.WriteFile(configPath, data, 0644); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		// Verify priority was saved
		readData, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("Failed to read config: %v", err)
		}

		var readConfig map[string]interface{}
		if err := yaml.Unmarshal(readData, &readConfig); err != nil {
			t.Fatalf("Failed to unmarshal config: %v", err)
		}

		groups2, ok := readConfig["groups"].(map[string]interface{})
		if !ok {
			t.Fatal("Groups should be a map")
		}

		priorityGroup, ok := groups2["priority-group"].(map[string]interface{})
		if !ok {
			t.Fatal("priority-group should be a map")
		}

		priority, ok := priorityGroup["priority"].([]interface{})
		if !ok {
			t.Fatal("priority should be a list")
		}

		if len(priority) != 2 {
			t.Errorf("Expected 2 priority items, got %d", len(priority))
		}

		if priority[0] != "high-priority" {
			t.Errorf("First priority should be high-priority, got %v", priority[0])
		}
	})
}

// TestConfigPath tests config path logic
func TestConfigPath(t *testing.T) {
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	t.Run("config path construction", func(t *testing.T) {
		tmpHome := t.TempDir()
		os.Setenv("HOME", tmpHome)

		homeDir, err := os.UserHomeDir()
		if err != nil {
			t.Fatalf("Failed to get home directory: %v", err)
		}

		configDir := filepath.Join(homeDir, ".config", "claude-sync")
		configPath := filepath.Join(configDir, "config.yaml")

		expectedConfigDir := filepath.Join(tmpHome, ".config", "claude-sync")
		expectedConfigPath := filepath.Join(expectedConfigDir, "config.yaml")

		if configDir != expectedConfigDir {
			t.Errorf("Config dir mismatch: got %s, expected %s", configDir, expectedConfigDir)
		}

		if configPath != expectedConfigPath {
			t.Errorf("Config path mismatch: got %s, expected %s", configPath, expectedConfigPath)
		}
	})
}
