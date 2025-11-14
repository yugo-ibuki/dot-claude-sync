package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration file",
	Long:  `Create a new configuration file at ~/.config/claude-sync/config.yaml with interactive prompts.`,
	RunE:  runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "claude-sync")
	configPath := filepath.Join(configDir, "config.yaml")

	// Check if config already exists
	if _, statErr := os.Stat(configPath); statErr == nil && !force {
		fmt.Printf("Configuration file already exists: %s\n", configPath)
		fmt.Print("Overwrite? [y/N]: ")
		var response string
		if _, scanErr := fmt.Scanln(&response); scanErr != nil {
			fmt.Println("Cancelled")
			return nil
		}
		if response != "y" && response != "Y" {
			fmt.Println("Cancelled")
			return nil
		}
	}

	if dryRun {
		fmt.Printf("DRY RUN: Would create %s\n", configPath)
		return nil
	}

	// Create config directory
	if mkdirErr := os.MkdirAll(configDir, 0755); mkdirErr != nil {
		return fmt.Errorf("failed to create config directory: %w", mkdirErr)
	}

	// Interactive configuration
	reader := bufio.NewReader(os.Stdin)
	groups := make(map[string]map[string]interface{})

	fmt.Println("Configuration file will be created at:", configPath)
	fmt.Println()
	fmt.Println("Let's set up your first group.")
	fmt.Println("You can add more groups later by editing the config file.")
	fmt.Println()

	// Get group name
	fmt.Print("Group name (e.g., 'web-projects', 'python-services'): ")
	groupName, _ := reader.ReadString('\n')
	groupName = strings.TrimSpace(groupName)
	if groupName == "" {
		groupName = "my-projects"
		fmt.Printf("Using default: %s\n", groupName)
	}

	// Get project paths
	fmt.Println()
	fmt.Println("Enter project paths (absolute paths to .claude directories)")
	fmt.Println("Example: /Users/yugo/projects/web-app/.claude")
	fmt.Println("Enter an empty line when done.")
	fmt.Println()

	paths := make(map[string]string)
	pathList := []string{}
	i := 1

	for {
		fmt.Printf("Project %d alias (or press Enter to skip): ", i)
		alias, _ := reader.ReadString('\n')
		alias = strings.TrimSpace(alias)

		fmt.Printf("Project %d path: ", i)
		path, _ := reader.ReadString('\n')
		path = strings.TrimSpace(path)

		if path == "" {
			break
		}

		if alias != "" {
			paths[alias] = path
		} else {
			pathList = append(pathList, path)
		}

		i++
		if i > 10 {
			fmt.Println("Maximum 10 projects reached.")
			break
		}
	}

	if len(paths) == 0 && len(pathList) == 0 {
		// Create example config
		fmt.Println()
		fmt.Println("No paths provided. Creating example configuration...")
		paths = map[string]string{
			"project-a": "~/workspace/project-a/.claude",
			"project-b": "~/workspace/project-b/.claude",
			"project-c": "~/workspace/project-c/.claude",
		}
	}

	// Build group config
	group := make(map[string]interface{})
	if len(paths) > 0 {
		group["paths"] = paths
		// Ask about priority
		if len(paths) > 1 {
			fmt.Println()
			fmt.Print("Set priority order? [y/N]: ")
			var setPriority string
			if _, scanErr := fmt.Scanln(&setPriority); scanErr == nil && (setPriority == "y" || setPriority == "Y") {
				priority := make([]string, 0)
				for alias := range paths {
					priority = append(priority, alias)
				}
				group["priority"] = priority
				fmt.Println("Default priority order created. Edit config file to customize.")
			}
		}
	} else {
		group["paths"] = pathList
	}

	groups[groupName] = group

	// Create YAML structure
	config := map[string]interface{}{
		"groups": groups,
	}

	// Write to file
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Println()
	fmt.Printf("✓ Configuration file created: %s\n", configPath)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Printf("  1. Edit the config file: vim %s\n", configPath)
	fmt.Printf("  2. View your groups: claude-sync list\n")
	fmt.Printf("  3. Sync files: claude-sync push %s\n", groupName)

	return nil
}
