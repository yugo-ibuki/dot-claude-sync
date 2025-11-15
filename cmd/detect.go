package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yugo-ibuki/dot-claude-sync/config"
	"gopkg.in/yaml.v3"
)

var detectCmd = &cobra.Command{
	Use:   "detect <worktree-root> --group <group-name>",
	Short: "Detect .claude directories in git worktrees",
	Long: `Automatically detect .claude directories from git worktrees and add them to a group.
This command runs 'git worktree list' in the specified directory and finds all .claude directories.`,
	Args: cobra.ExactArgs(1),
	RunE: runDetect,
}

var groupName string

func init() {
	rootCmd.AddCommand(detectCmd)
	detectCmd.Flags().StringVarP(&groupName, "group", "g", "", "group name to add detected paths (required)")
	detectCmd.MarkFlagRequired("group")
}

func runDetect(cmd *cobra.Command, args []string) error {
	worktreeRoot := args[0]

	// Expand ~ to home directory
	if strings.HasPrefix(worktreeRoot, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		worktreeRoot = filepath.Join(homeDir, worktreeRoot[1:])
	}

	// Resolve to absolute path
	absRoot, err := filepath.Abs(worktreeRoot)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	// Check if directory exists
	if _, err := os.Stat(absRoot); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", absRoot)
	}

	// Execute git worktree list
	worktrees, err := getWorktreePaths(absRoot)
	if err != nil {
		return fmt.Errorf("failed to get worktree list: %w", err)
	}

	if len(worktrees) == 0 {
		fmt.Println("No worktrees found.")
		return nil
	}

	// Detect .claude directories
	claudeDirs := []string{}
	for _, wt := range worktrees {
		claudeDir := filepath.Join(wt, ".claude")
		if info, err := os.Stat(claudeDir); err == nil && info.IsDir() {
			claudeDirs = append(claudeDirs, claudeDir)
		}
	}

	if len(claudeDirs) == 0 {
		fmt.Println("No .claude directories found in worktrees.")
		return nil
	}

	// Display detected paths
	fmt.Printf("Found %d .claude director%s:\n", len(claudeDirs), pluralize(len(claudeDirs)))
	for i, dir := range claudeDirs {
		fmt.Printf("  %d. %s\n", i+1, dir)
	}

	if dryRun {
		fmt.Printf("\nDRY RUN: Would add these paths to group '%s'\n", groupName)
		return nil
	}

	// Confirm before adding
	if !force {
		fmt.Printf("\nAdd these paths to group '%s'? [y/N]: ", groupName)
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(response)
		if response != "y" && response != "Y" {
			fmt.Println("Cancelled")
			return nil
		}
	}

	// Load existing config
	cfg, err := loadOrCreateConfig()
	if err != nil {
		return err
	}

	// Add paths to group
	if err := addPathsToGroup(cfg, groupName, claudeDirs); err != nil {
		return err
	}

	// Save config
	if err := saveConfig(cfg); err != nil {
		return err
	}

	fmt.Printf("\n✓ Added %d path%s to group '%s'\n", len(claudeDirs), pluralize(len(claudeDirs)), groupName)
	fmt.Println("\nNext steps:")
	fmt.Printf("  1. Review configuration: claude-sync list %s\n", groupName)
	fmt.Printf("  2. Sync files: claude-sync push %s\n", groupName)

	return nil
}

// getWorktreePaths executes 'git worktree list --porcelain' and returns worktree paths
func getWorktreePaths(rootDir string) ([]string, error) {
	cmd := exec.Command("git", "-C", rootDir, "worktree", "list", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git worktree list failed (is this a git repository?): %w", err)
	}

	// Parse porcelain output
	// Format:
	// worktree /path/to/worktree
	// HEAD <commit>
	// branch refs/heads/branch-name
	//
	// worktree /path/to/another
	// ...
	var paths []string
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "worktree ") {
			path := strings.TrimPrefix(line, "worktree ")
			paths = append(paths, path)
		}
	}

	return paths, nil
}

// loadOrCreateConfig loads existing config or creates a new one
func loadOrCreateConfig() (*config.Config, error) {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		// If config doesn't exist, create a new one
		if strings.Contains(err.Error(), "not found") {
			fmt.Println("Configuration file not found. Creating new configuration...")
			return &config.Config{
				Groups: make(map[string]*config.Group),
			}, nil
		}
		return nil, err
	}
	return cfg, nil
}

// addPathsToGroup adds detected paths to the specified group
func addPathsToGroup(cfg *config.Config, groupName string, paths []string) error {
	group, exists := cfg.Groups[groupName]
	if !exists {
		// Create new group with simple list format
		cfg.Groups[groupName] = &config.Group{
			Paths: paths,
		}
		return nil
	}

	// Add to existing group
	switch existingPaths := group.Paths.(type) {
	case []interface{}:
		// Append to list
		for _, p := range paths {
			existingPaths = append(existingPaths, p)
		}
		group.Paths = existingPaths
	case map[string]interface{}:
		// Convert to list and append
		pathList := make([]interface{}, 0)
		for _, v := range existingPaths {
			pathList = append(pathList, v)
		}
		for _, p := range paths {
			pathList = append(pathList, p)
		}
		group.Paths = pathList
	case []string:
		// Append to list (should not happen after YAML unmarshal, but handle it)
		for _, p := range paths {
			existingPaths = append(existingPaths, p)
		}
		// Convert to []interface{} for YAML marshaling
		interfaceList := make([]interface{}, len(existingPaths))
		for i, v := range existingPaths {
			interfaceList[i] = v
		}
		group.Paths = interfaceList
	default:
		return fmt.Errorf("unexpected paths format in group '%s'", groupName)
	}

	return nil
}

// saveConfig saves the configuration to file
func saveConfig(cfg *config.Config) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "claude-sync")
	configPath := filepath.Join(configDir, "config.yaml")

	// Use custom config path if specified
	if cfgFile != "" {
		configPath = cfgFile
	}

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal config to YAML
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	if verbose {
		fmt.Printf("Configuration saved to: %s\n", configPath)
	}

	return nil
}

// pluralize returns "ies" for count != 1, otherwise "y"
func pluralize(count int) string {
	if count == 1 {
		return "y"
	}
	return "ies"
}
