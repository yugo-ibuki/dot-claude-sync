package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the root configuration structure
type Config struct {
	Groups map[string]*Group `yaml:"groups"`
}

// Group represents a project group configuration
type Group struct {
	Paths    interface{} `yaml:"paths"`    // Can be map[string]string or []string
	Priority []string    `yaml:"priority"` // Optional priority list
}

// ProjectPath represents a resolved project path with alias and priority
type ProjectPath struct {
	Alias    string
	Path     string
	Priority int
}

// Load loads the configuration file from the specified path or default location
func Load(configPath string) (*Config, error) {
	path, err := getConfigPath(configPath)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// getConfigPath returns the configuration file path
func getConfigPath(configPath string) (string, error) {
	// If explicit path is provided, use it
	if configPath != "" {
		if _, err := os.Stat(configPath); err != nil {
			return "", fmt.Errorf("config file not found: %s", configPath)
		}
		return configPath, nil
	}

	// Use fixed global config location
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	defaultPath := filepath.Join(homeDir, ".config", "claude-sync", "config.yaml")
	if _, err := os.Stat(defaultPath); err != nil {
		return "", fmt.Errorf("configuration file not found: %s\n\nRun 'claude-sync init' to create a configuration file, or use --config flag", defaultPath)
	}

	return defaultPath, nil
}

// GetGroup returns the specified group configuration
func (c *Config) GetGroup(name string) (*Group, error) {
	group, ok := c.Groups[name]
	if !ok {
		return nil, fmt.Errorf("group '%s' not found in configuration", name)
	}
	return group, nil
}

// ListGroups returns all group names
func (c *Config) ListGroups() []string {
	groups := make([]string, 0, len(c.Groups))
	for name := range c.Groups {
		groups = append(groups, name)
	}
	return groups
}

// GetProjectPaths returns resolved project paths with priorities
func (g *Group) GetProjectPaths() ([]ProjectPath, error) {
	var projects []ProjectPath

	// Parse paths (can be map or slice)
	switch paths := g.Paths.(type) {
	case map[string]interface{}:
		// Alias format
		for alias, path := range paths {
			pathStr, ok := path.(string)
			if !ok {
				return nil, fmt.Errorf("invalid path value for alias '%s'", alias)
			}
			projects = append(projects, ProjectPath{
				Alias: alias,
				Path:  pathStr,
			})
		}
	case []interface{}:
		// Simple list format
		for i, path := range paths {
			pathStr, ok := path.(string)
			if !ok {
				return nil, fmt.Errorf("invalid path value at index %d", i)
			}
			projects = append(projects, ProjectPath{
				Alias: filepath.Base(pathStr),
				Path:  pathStr,
			})
		}
	default:
		return nil, fmt.Errorf("invalid paths format: must be map or list")
	}

	// Assign priorities
	if len(g.Priority) > 0 {
		// Use explicit priority list
		priorityMap := make(map[string]int)
		for i, p := range g.Priority {
			priorityMap[p] = i + 1
		}

		for i := range projects {
			if priority, ok := priorityMap[projects[i].Alias]; ok {
				projects[i].Priority = priority
			} else if priority, ok := priorityMap[projects[i].Path]; ok {
				projects[i].Priority = priority
			} else {
				projects[i].Priority = len(g.Priority) + 1 // Lowest priority
			}
		}
	} else {
		// Use paths order as priority
		for i := range projects {
			projects[i].Priority = i + 1
		}
	}

	return projects, nil
}
