package config

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestLoadConfig tests the Load function with various config formats
func TestLoadConfig(t *testing.T) {
	// Save original HOME and restore after test
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	t.Run("load valid config with map paths", func(t *testing.T) {
		tmpHome := t.TempDir()
		os.Setenv("HOME", tmpHome)

		configDir := filepath.Join(tmpHome, ".config", "claude-sync")
		configPath := filepath.Join(configDir, "config.yaml")

		// Create config directory
		if err := os.MkdirAll(configDir, 0755); err != nil {
			t.Fatalf("Failed to create config directory: %v", err)
		}

		// Create valid config
		configContent := `
groups:
  test-group:
    paths:
      project-a: /path/to/project-a/.claude
      project-b: /path/to/project-b/.claude
    priority:
      - project-a
      - project-b
`
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write config: %v", err)
		}

		// Load config
		config, err := Load("")
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		if config.Groups == nil {
			t.Error("Groups should not be nil")
		}

		if _, ok := config.Groups["test-group"]; !ok {
			t.Error("test-group should exist")
		}
	})

	t.Run("load valid config with list paths", func(t *testing.T) {
		tmpHome := t.TempDir()
		os.Setenv("HOME", tmpHome)

		configDir := filepath.Join(tmpHome, ".config", "claude-sync")
		configPath := filepath.Join(configDir, "config.yaml")

		if err := os.MkdirAll(configDir, 0755); err != nil {
			t.Fatalf("Failed to create config directory: %v", err)
		}

		configContent := `
groups:
  simple-group:
    paths:
      - /path/to/project1/.claude
      - /path/to/project2/.claude
`
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write config: %v", err)
		}

		config, err := Load("")
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		if _, ok := config.Groups["simple-group"]; !ok {
			t.Error("simple-group should exist")
		}
	})

	t.Run("load config with explicit path", func(t *testing.T) {
		tmpDir := t.TempDir()
		customConfigPath := filepath.Join(tmpDir, "custom-config.yaml")

		configContent := `
groups:
  custom-group:
    paths:
      - /path/to/custom/.claude
`
		if err := os.WriteFile(customConfigPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write config: %v", err)
		}

		config, err := Load(customConfigPath)
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		if _, ok := config.Groups["custom-group"]; !ok {
			t.Error("custom-group should exist")
		}
	})

	t.Run("error on non-existent config file", func(t *testing.T) {
		tmpHome := t.TempDir()
		os.Setenv("HOME", tmpHome)

		_, err := Load("")
		if err == nil {
			t.Error("Should error when config file doesn't exist")
		}
	})

	t.Run("error on invalid YAML", func(t *testing.T) {
		tmpHome := t.TempDir()
		os.Setenv("HOME", tmpHome)

		configDir := filepath.Join(tmpHome, ".config", "claude-sync")
		configPath := filepath.Join(configDir, "config.yaml")

		if err := os.MkdirAll(configDir, 0755); err != nil {
			t.Fatalf("Failed to create config directory: %v", err)
		}

		// Write invalid YAML
		invalidYAML := "groups:\n  invalid: [unclosed bracket"
		if err := os.WriteFile(configPath, []byte(invalidYAML), 0644); err != nil {
			t.Fatalf("Failed to write config: %v", err)
		}

		_, err := Load("")
		if err == nil {
			t.Error("Should error on invalid YAML")
		}
	})

	t.Run("error on non-existent explicit path", func(t *testing.T) {
		_, err := Load("/non/existent/path.yaml")
		if err == nil {
			t.Error("Should error when explicit path doesn't exist")
		}
	})
}

// TestGetGroup tests the GetGroup method
func TestGetGroup(t *testing.T) {
	config := &Config{
		Groups: map[string]*Group{
			"existing-group": {
				Paths: map[string]interface{}{
					"project-a": "/path/to/a/.claude",
				},
			},
		},
	}

	t.Run("get existing group", func(t *testing.T) {
		group, err := config.GetGroup("existing-group")
		if err != nil {
			t.Errorf("Should not error: %v", err)
		}
		if group == nil {
			t.Error("Group should not be nil")
		}
	})

	t.Run("error on non-existent group", func(t *testing.T) {
		_, err := config.GetGroup("non-existent")
		if err == nil {
			t.Error("Should error when group doesn't exist")
		}
	})
}

// TestListGroups tests the ListGroups method
func TestListGroups(t *testing.T) {
	config := &Config{
		Groups: map[string]*Group{
			"group-a": {},
			"group-b": {},
			"group-c": {},
		},
	}

	groups := config.ListGroups()
	if len(groups) != 3 {
		t.Errorf("Expected 3 groups, got %d", len(groups))
	}

	// Check all groups are present
	groupMap := make(map[string]bool)
	for _, name := range groups {
		groupMap[name] = true
	}

	expectedGroups := []string{"group-a", "group-b", "group-c"}
	for _, expected := range expectedGroups {
		if !groupMap[expected] {
			t.Errorf("Expected group %s not found", expected)
		}
	}
}

// TestGetProjectPaths tests the GetProjectPaths method with various formats
func TestGetProjectPaths(t *testing.T) {
	t.Run("map paths with explicit priority", func(t *testing.T) {
		group := &Group{
			Paths: map[string]interface{}{
				"project-a": "/path/to/a/.claude",
				"project-b": "/path/to/b/.claude",
				"project-c": "/path/to/c/.claude",
			},
			Priority: []string{"project-b", "project-a", "project-c"},
		}

		projects, err := group.GetProjectPaths()
		if err != nil {
			t.Fatalf("Failed to get project paths: %v", err)
		}

		if len(projects) != 3 {
			t.Fatalf("Expected 3 projects, got %d", len(projects))
		}

		// Verify priorities are assigned correctly
		priorityMap := make(map[string]int)
		for _, proj := range projects {
			priorityMap[proj.Alias] = proj.Priority
		}

		if priorityMap["project-b"] != 1 {
			t.Errorf("project-b should have priority 1, got %d", priorityMap["project-b"])
		}
		if priorityMap["project-a"] != 2 {
			t.Errorf("project-a should have priority 2, got %d", priorityMap["project-a"])
		}
		if priorityMap["project-c"] != 3 {
			t.Errorf("project-c should have priority 3, got %d", priorityMap["project-c"])
		}
	})

	t.Run("map paths without explicit priority", func(t *testing.T) {
		group := &Group{
			Paths: map[string]interface{}{
				"project-a": "/path/to/a/.claude",
				"project-b": "/path/to/b/.claude",
			},
		}

		projects, err := group.GetProjectPaths()
		if err != nil {
			t.Fatalf("Failed to get project paths: %v", err)
		}

		if len(projects) != 2 {
			t.Fatalf("Expected 2 projects, got %d", len(projects))
		}

		// All projects should have priorities assigned
		for _, proj := range projects {
			if proj.Priority == 0 {
				t.Errorf("Project %s should have non-zero priority", proj.Alias)
			}
		}
	})

	t.Run("list paths format", func(t *testing.T) {
		group := &Group{
			Paths: []interface{}{
				"/path/to/project1/.claude",
				"/path/to/project2/.claude",
			},
		}

		projects, err := group.GetProjectPaths()
		if err != nil {
			t.Fatalf("Failed to get project paths: %v", err)
		}

		if len(projects) != 2 {
			t.Fatalf("Expected 2 projects, got %d", len(projects))
		}

		// Check aliases are derived from paths
		if projects[0].Alias != ".claude" {
			t.Errorf("Expected alias '.claude', got %s", projects[0].Alias)
		}

		// Check priorities are sequential
		if projects[0].Priority != 1 || projects[1].Priority != 2 {
			t.Error("Priorities should be sequential starting from 1")
		}
	})

	t.Run("priority by path when alias not in priority list", func(t *testing.T) {
		group := &Group{
			Paths: map[string]interface{}{
				"proj-a": "/path/to/project-a/.claude",
				"proj-b": "/path/to/project-b/.claude",
			},
			Priority: []string{"/path/to/project-b/.claude", "/path/to/project-a/.claude"},
		}

		projects, err := group.GetProjectPaths()
		if err != nil {
			t.Fatalf("Failed to get project paths: %v", err)
		}

		priorityMap := make(map[string]int)
		for _, proj := range projects {
			priorityMap[proj.Alias] = proj.Priority
		}

		// Should match by path when alias not in priority list
		if priorityMap["proj-b"] != 1 {
			t.Errorf("proj-b should have priority 1, got %d", priorityMap["proj-b"])
		}
		if priorityMap["proj-a"] != 2 {
			t.Errorf("proj-a should have priority 2, got %d", priorityMap["proj-a"])
		}
	})

	t.Run("projects not in priority list get lowest priority", func(t *testing.T) {
		group := &Group{
			Paths: map[string]interface{}{
				"project-a": "/path/to/a/.claude",
				"project-b": "/path/to/b/.claude",
				"project-c": "/path/to/c/.claude",
			},
			Priority: []string{"project-a"},
		}

		projects, err := group.GetProjectPaths()
		if err != nil {
			t.Fatalf("Failed to get project paths: %v", err)
		}

		priorityMap := make(map[string]int)
		for _, proj := range projects {
			priorityMap[proj.Alias] = proj.Priority
		}

		if priorityMap["project-a"] != 1 {
			t.Errorf("project-a should have priority 1, got %d", priorityMap["project-a"])
		}

		// project-b and project-c should have priority 2 (lowest)
		if priorityMap["project-b"] != 2 {
			t.Errorf("project-b should have priority 2, got %d", priorityMap["project-b"])
		}
		if priorityMap["project-c"] != 2 {
			t.Errorf("project-c should have priority 2, got %d", priorityMap["project-c"])
		}
	})

	t.Run("error on invalid path type in map", func(t *testing.T) {
		group := &Group{
			Paths: map[string]interface{}{
				"project-a": 123, // Invalid: should be string
			},
		}

		_, err := group.GetProjectPaths()
		if err == nil {
			t.Error("Should error on invalid path type")
		}
	})

	t.Run("error on invalid path type in list", func(t *testing.T) {
		group := &Group{
			Paths: []interface{}{
				123, // Invalid: should be string
			},
		}

		_, err := group.GetProjectPaths()
		if err == nil {
			t.Error("Should error on invalid path type")
		}
	})

	t.Run("error on invalid paths format", func(t *testing.T) {
		group := &Group{
			Paths: "invalid-string-format",
		}

		_, err := group.GetProjectPaths()
		if err == nil {
			t.Error("Should error on invalid paths format")
		}
	})
}

// TestYAMLRoundTrip tests YAML marshaling and unmarshaling
func TestYAMLRoundTrip(t *testing.T) {
	t.Run("round trip with map paths", func(t *testing.T) {
		original := &Config{
			Groups: map[string]*Group{
				"test-group": {
					Paths: map[string]interface{}{
						"project-a": "/path/to/a/.claude",
						"project-b": "/path/to/b/.claude",
					},
					Priority: []string{"project-a", "project-b"},
				},
			},
		}

		// Marshal to YAML
		data, err := yaml.Marshal(original)
		if err != nil {
			t.Fatalf("Failed to marshal: %v", err)
		}

		// Unmarshal back
		var decoded Config
		if err := yaml.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		// Verify structure
		if decoded.Groups == nil {
			t.Error("Groups should not be nil")
		}
		if _, ok := decoded.Groups["test-group"]; !ok {
			t.Error("test-group should exist after round trip")
		}
	})

	t.Run("round trip with list paths", func(t *testing.T) {
		original := &Config{
			Groups: map[string]*Group{
				"simple-group": {
					Paths: []interface{}{
						"/path/to/project1/.claude",
						"/path/to/project2/.claude",
					},
				},
			},
		}

		data, err := yaml.Marshal(original)
		if err != nil {
			t.Fatalf("Failed to marshal: %v", err)
		}

		var decoded Config
		if err := yaml.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		if _, ok := decoded.Groups["simple-group"]; !ok {
			t.Error("simple-group should exist after round trip")
		}
	})
}
