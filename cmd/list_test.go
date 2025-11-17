package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunList(t *testing.T) {
	// Create temporary test directory
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

	// Create test config file with multiple groups
	configContent := `groups:
  web-projects:
    paths:
      frontend: ` + project1 + `
      backend: ` + project2 + `
    priority:
      - frontend
      - backend
  go-projects:
    paths:
      - ` + project3 + `
  mixed-group:
    paths:
      proj1: ` + project1 + `
      proj2: ` + project2 + `
`
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	tests := []struct {
		name       string
		args       []string
		setupFlags func()
		wantErr    bool
	}{
		{
			name: "list all groups without arguments",
			args: []string{},
			setupFlags: func() {
				verbose = false
			},
			wantErr: false,
		},
		{
			name: "list specific group details - web-projects",
			args: []string{"web-projects"},
			setupFlags: func() {
				verbose = false
			},
			wantErr: false,
		},
		{
			name: "list specific group details - go-projects",
			args: []string{"go-projects"},
			setupFlags: func() {
				verbose = false
			},
			wantErr: false,
		},
		{
			name: "list specific group details - mixed-group",
			args: []string{"mixed-group"},
			setupFlags: func() {
				verbose = false
			},
			wantErr: false,
		},
		{
			name: "error when group does not exist",
			args: []string{"nonexistent-group"},
			setupFlags: func() {
				verbose = false
			},
			wantErr: true,
		},
		{
			name: "verbose flag enabled",
			args: []string{},
			setupFlags: func() {
				verbose = true
			},
			wantErr: false,
		},
		{
			name: "verbose flag with specific group",
			args: []string{"web-projects"},
			setupFlags: func() {
				verbose = true
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			tt.setupFlags()
			cfgFile = configPath

			// Run command
			err := runList(nil, tt.args)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("runList() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Reset flags
			verbose = false
		})
	}
}

func TestListGroupSorting(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test projects
	project1 := filepath.Join(tmpDir, "p1", ".claude")
	project2 := filepath.Join(tmpDir, "p2", ".claude")
	project3 := filepath.Join(tmpDir, "p3", ".claude")

	for _, dir := range []string{project1, project2, project3} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}
	}

	// Create config with priority and without priority
	configContent := `groups:
  with-priority:
    paths:
      high: ` + project1 + `
      medium: ` + project2 + `
      low: ` + project3 + `
    priority:
      - high
      - medium
      - low
  without-priority:
    paths:
      - ` + project1 + `
      - ` + project2 + `
      - ` + project3 + `
`
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	tests := []struct {
		name      string
		groupName string
		wantErr   bool
	}{
		{
			name:      "group with explicit priority order",
			groupName: "with-priority",
			wantErr:   false,
		},
		{
			name:      "group without explicit priority",
			groupName: "without-priority",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfgFile = configPath
			err := runList(nil, []string{tt.groupName})

			if (err != nil) != tt.wantErr {
				t.Errorf("runList() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestListEmptyConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// Create empty config
	configContent := `groups: {}`
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	cfgFile = configPath

	// List all groups (should show nothing but not error)
	err := runList(nil, []string{})
	if err != nil {
		t.Errorf("runList() should not error on empty config: %v", err)
	}
}

func TestListInvalidConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// Create config with invalid YAML
	configContent := `this is not valid yaml: [[[`
	configPath := filepath.Join(tmpDir, "invalid.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	cfgFile = configPath

	// Should return error when loading invalid config
	err := runList(nil, []string{})
	if err == nil {
		t.Error("runList() should error on invalid config")
	}
}

func TestListNonExistentConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nonexistent.yaml")

	cfgFile = configPath

	// Should return error when config file doesn't exist
	err := runList(nil, []string{})
	if err == nil {
		t.Error("runList() should error when config file doesn't exist")
	}
}
