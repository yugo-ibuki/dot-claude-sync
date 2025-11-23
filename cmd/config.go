package cmd

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yugo-ibuki/dot-claude-sync/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long:  `Manage groups and projects in the configuration file.`,
}

var configListCmd = &cobra.Command{
	Use:   "list [group]",
	Short: "List configuration",
	Long:  `Display the entire configuration or details of a specific group.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runConfigList,
}

var configAddGroupCmd = &cobra.Command{
	Use:   "add-group <name>",
	Short: "Add a new group",
	Long:  `Add a new group to the configuration.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runConfigAddGroup,
}

var configRemoveGroupCmd = &cobra.Command{
	Use:   "remove-group [name]",
	Short: "Remove a group",
	Long: `Remove a group from the configuration.

Interactive mode (no arguments):
  dot-claude-sync config remove-group

  This will display a list of available groups to select from.

Argument mode (1 argument):
  dot-claude-sync config remove-group <name>`,
	Args: cobra.MaximumNArgs(1),
	RunE: runConfigRemoveGroup,
}

var configAddProjectCmd = &cobra.Command{
	Use:   "add-project [group] [alias] [path]",
	Short: "Add a project to a group",
	Long: `Add a new project to an existing group.

Interactive mode (no arguments):
  dot-claude-sync config add-project

  This will prompt you to:
  1. Select an existing group or create a new one
  2. Enter a project alias
  3. Enter the project path

Argument mode (3 arguments):
  dot-claude-sync config add-project <group> <alias> <path>`,
	Args: cobra.MaximumNArgs(3),
	RunE: runConfigAddProject,
}

var configRemoveProjectCmd = &cobra.Command{
	Use:   "remove-project [group] [alias]",
	Short: "Remove a project from a group",
	Long: `Remove a project from an existing group.

Interactive mode (no arguments):
  dot-claude-sync config remove-project

  This will prompt you to:
  1. Select a group
  2. Select a project to remove from that group

Argument mode (2 arguments):
  dot-claude-sync config remove-project <group> <alias>`,
	Args: cobra.MaximumNArgs(2),
	RunE: runConfigRemoveProject,
}

var configSetPriorityCmd = &cobra.Command{
	Use:   "set-priority [group] [alias1] [alias2] [alias3]...",
	Short: "Set priority order for a group",
	Long: `Set the priority order for projects in a group. First alias has highest priority.

Interactive mode (no arguments):
  dot-claude-sync config set-priority

  This will prompt you to:
  1. Select a group
  2. Select projects in priority order (highest to lowest)

Argument mode (2+ arguments):
  dot-claude-sync config set-priority <group> <alias1> [alias2] [alias3]...`,
	Args: cobra.MinimumNArgs(0),
	RunE: runConfigSetPriority,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configAddGroupCmd)
	configCmd.AddCommand(configRemoveGroupCmd)
	configCmd.AddCommand(configAddProjectCmd)
	configCmd.AddCommand(configRemoveProjectCmd)
	configCmd.AddCommand(configSetPriorityCmd)
}

func runConfigList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return err
	}

	if len(args) == 0 {
		// Show all groups
		fmt.Println("Configuration:")
		fmt.Println()

		groups := cfg.ListGroups()
		if len(groups) == 0 {
			fmt.Println("No groups configured")
			return nil
		}

		for _, groupName := range groups {
			group, _ := cfg.GetGroup(groupName)
			projects, err := group.GetProjectPaths()
			if err != nil {
				fmt.Printf("⚠ Group '%s': error parsing projects: %v\n", groupName, err)
				continue
			}

			fmt.Printf("📦 %s (%d projects)\n", groupName, len(projects))
			for _, proj := range projects {
				fmt.Printf("  [%d] %s → %s\n", proj.Priority, proj.Alias, proj.Path)
			}
			fmt.Println()
		}
	} else {
		// Show specific group
		groupName := args[0]
		group, err := cfg.GetGroup(groupName)
		if err != nil {
			return err
		}

		projects, err := group.GetProjectPaths()
		if err != nil {
			return fmt.Errorf("failed to get project paths: %w", err)
		}

		fmt.Printf("Group: %s\n", groupName)
		fmt.Println()
		fmt.Printf("Projects (%d):\n", len(projects))
		for _, proj := range projects {
			fmt.Printf("  [%d] %s\n", proj.Priority, proj.Alias)
			fmt.Printf("      %s\n", proj.Path)
		}

		if len(group.Priority) > 0 {
			fmt.Println()
			fmt.Printf("Priority: %v\n", group.Priority)
		}
	}

	return nil
}

func runConfigAddGroup(cmd *cobra.Command, args []string) error {
	groupName := args[0]

	cfg, err := config.Load(cfgFile)
	if err != nil {
		return err
	}

	if err := cfg.AddGroup(groupName); err != nil {
		return err
	}

	if err := cfg.Save(cfgFile); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("✓ Added group '%s'\n", groupName)
	return nil
}

func runConfigRemoveGroup(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return err
	}

	var groupName string

	// Handle different argument modes based on argument count
	switch len(args) {
	case 0:
		// Interactive mode: No arguments provided
		reader := bufio.NewReader(os.Stdin)

		// List available groups
		groupNames := make([]string, 0, len(cfg.Groups))
		for name := range cfg.Groups {
			groupNames = append(groupNames, name)
		}
		sort.Strings(groupNames)

		if len(groupNames) == 0 {
			fmt.Println("No groups found")
			return nil
		}

		// Display groups with project counts
		fmt.Println("Available groups:")
		fmt.Println()
		for i, name := range groupNames {
			group, _ := cfg.GetGroup(name)
			projects, _ := group.GetProjectPaths()
			fmt.Printf("  %d. %s (%d projects)\n", i+1, name, len(projects))
		}
		fmt.Println()

		// Prompt for selection
		fmt.Print("Select group number to remove (or press Enter to cancel): ")
		selection, _ := reader.ReadString('\n')
		selection = strings.TrimSpace(selection)

		if selection == "" {
			fmt.Println("Cancelled")
			return nil
		}

		idx, err := strconv.Atoi(selection)
		if err != nil || idx < 1 || idx > len(groupNames) {
			return fmt.Errorf("invalid selection: %s", selection)
		}
		groupName = groupNames[idx-1]

	case 1:
		// Argument mode: 1 argument provided
		groupName = args[0]

	default:
		return fmt.Errorf("invalid number of arguments: expected 0 or 1, got %d", len(args))
	}

	// Confirm deletion
	if !force {
		group, err := cfg.GetGroup(groupName)
		if err != nil {
			return err
		}

		projects, _ := group.GetProjectPaths()
		fmt.Printf("This will remove group '%s' with %d projects.\n", groupName, len(projects))
		fmt.Print("Continue? [y/N]: ")

		var response string
		if _, err := fmt.Scanln(&response); err != nil {
			// If Scanln fails (e.g., EOF), treat as cancelled
			fmt.Println("\nCancelled")
			return nil
		}
		if response != "y" && response != "Y" {
			fmt.Println("Cancelled")
			return nil
		}
	}

	if err := cfg.RemoveGroup(groupName); err != nil {
		return err
	}

	if err := cfg.Save(cfgFile); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("✓ Removed group '%s'\n", groupName)
	return nil
}

func runConfigAddProject(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return err
	}

	var groupName, alias, path string
	var createNew bool

	// Handle different argument modes based on argument count
	switch len(args) {
	case 0:
		// Interactive mode: No arguments provided
		// Prompts user to select an existing group or create a new one,
		// then asks for project alias and path interactively
		reader := bufio.NewReader(os.Stdin)

		// Step 1: List available groups and let user select or create new one
		groupNames := make([]string, 0, len(cfg.Groups))
		for name := range cfg.Groups {
			groupNames = append(groupNames, name)
		}
		sort.Strings(groupNames)

		// Display existing groups if any
		if len(groupNames) > 0 {
			fmt.Println("Available groups:")
			fmt.Println()
			for i, name := range groupNames {
				fmt.Printf("  %d. %s\n", i+1, name)
			}
			fmt.Println()
		}

		// Step 2: Prompt user to select existing group or create new one
		var prompt string
		if len(groupNames) > 0 {
			prompt = "Select group number (or press Enter to create a new group): "
		} else {
			prompt = "No groups found. Press Enter to create a new group: "
		}
		fmt.Print(prompt)
		selection, _ := reader.ReadString('\n')
		selection = strings.TrimSpace(selection)

		if selection == "" {
			// User pressed Enter without input: Create new group
			createNew = true
			fmt.Print("New group name: ")
			groupName, _ = reader.ReadString('\n')
			groupName = strings.TrimSpace(groupName)
			if groupName == "" {
				return fmt.Errorf("group name cannot be empty")
			}
			// Add new group to configuration
			if err := cfg.AddGroup(groupName); err != nil {
				return fmt.Errorf("failed to add group: %w", err)
			}
		} else {
			// User entered a number: Select existing group
			idx, err := strconv.Atoi(selection)
			if err != nil || idx < 1 || idx > len(groupNames) {
				return fmt.Errorf("invalid selection: %s", selection)
			}
			groupName = groupNames[idx-1]
		}

		fmt.Println()
		fmt.Printf("Adding project to group: %s\n", groupName)
		fmt.Println()

		// Step 3: Get project details (alias and path)
		fmt.Print("Project alias: ")
		alias, _ = reader.ReadString('\n')
		alias = strings.TrimSpace(alias)

		if alias == "" {
			return fmt.Errorf("project alias cannot be empty")
		}

		fmt.Print("Project path (absolute path to .claude directory): ")
		path, _ = reader.ReadString('\n')
		path = strings.TrimSpace(path)

		if path == "" {
			return fmt.Errorf("project path cannot be empty")
		}

		if dryRun {
			if createNew {
				fmt.Printf("DRY RUN: Would create new group '%s'\n", groupName)
			}
			fmt.Printf("DRY RUN: Would add project '%s' -> %s to group '%s'\n", alias, path, groupName)
			return nil
		}

	case 3:
		// Argument mode: 3 arguments provided
		// Usage: dot-claude-sync config add-project <group> <alias> <path>
		// Non-interactive mode for scripting and CI/CD workflows
		groupName = args[0]
		alias = args[1]
		path = args[2]

		if dryRun {
			fmt.Printf("DRY RUN: Would add project '%s' -> %s to group '%s'\n", alias, path, groupName)
			return nil
		}

	default:
		// Invalid argument count: Must be either 0 (interactive) or 3 (argument mode)
		return fmt.Errorf("invalid number of arguments: expected 0 or 3, got %d", len(args))
	}

	if err := cfg.AddProject(groupName, alias, path); err != nil {
		return err
	}

	if err := cfg.Save(cfgFile); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println()
	fmt.Printf("✓ Added project '%s' to group '%s'\n", alias, groupName)
	fmt.Printf("  Path: %s\n", path)
	return nil
}

func runConfigRemoveProject(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return err
	}

	var groupName, alias string

	// Handle different argument modes based on argument count
	switch len(args) {
	case 0:
		// Interactive mode: No arguments provided
		reader := bufio.NewReader(os.Stdin)

		// Step 1: List available groups
		groupNames := make([]string, 0, len(cfg.Groups))
		for name := range cfg.Groups {
			groupNames = append(groupNames, name)
		}
		sort.Strings(groupNames)

		if len(groupNames) == 0 {
			fmt.Println("No groups found")
			return nil
		}

		// Display groups
		fmt.Println("Available groups:")
		fmt.Println()
		for i, name := range groupNames {
			group, _ := cfg.GetGroup(name)
			projects, _ := group.GetProjectPaths()
			fmt.Printf("  %d. %s (%d projects)\n", i+1, name, len(projects))
		}
		fmt.Println()

		// Prompt for group selection
		fmt.Print("Select group number (or press Enter to cancel): ")
		selection, _ := reader.ReadString('\n')
		selection = strings.TrimSpace(selection)

		if selection == "" {
			fmt.Println("Cancelled")
			return nil
		}

		idx, err := strconv.Atoi(selection)
		if err != nil || idx < 1 || idx > len(groupNames) {
			return fmt.Errorf("invalid selection: %s", selection)
		}
		groupName = groupNames[idx-1]

		// Step 2: List projects in selected group
		group, err := cfg.GetGroup(groupName)
		if err != nil {
			return err
		}

		projects, err := group.GetProjectPaths()
		if err != nil {
			return fmt.Errorf("failed to get project paths: %w", err)
		}

		if len(projects) == 0 {
			fmt.Printf("Group '%s' has no projects\n", groupName)
			return nil
		}

		fmt.Println()
		fmt.Printf("Projects in group '%s':\n", groupName)
		fmt.Println()
		for i, proj := range projects {
			fmt.Printf("  %d. %s\n", i+1, proj.Alias)
			fmt.Printf("      %s\n", proj.Path)
		}
		fmt.Println()

		// Prompt for project selection
		fmt.Print("Select project number to remove (or press Enter to cancel): ")
		selection, _ = reader.ReadString('\n')
		selection = strings.TrimSpace(selection)

		if selection == "" {
			fmt.Println("Cancelled")
			return nil
		}

		idx, err = strconv.Atoi(selection)
		if err != nil || idx < 1 || idx > len(projects) {
			return fmt.Errorf("invalid selection: %s", selection)
		}
		alias = projects[idx-1].Alias

	case 2:
		// Argument mode: 2 arguments provided
		groupName = args[0]
		alias = args[1]

	default:
		return fmt.Errorf("invalid number of arguments: expected 0 or 2, got %d", len(args))
	}

	if err := cfg.RemoveProject(groupName, alias); err != nil {
		return err
	}

	if err := cfg.Save(cfgFile); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println()
	fmt.Printf("✓ Removed project '%s' from group '%s'\n", alias, groupName)
	return nil
}

func runConfigSetPriority(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return err
	}

	var groupName string
	var aliases []string

	// Handle different argument modes based on argument count
	switch len(args) {
	case 0:
		// Interactive mode: No arguments provided
		reader := bufio.NewReader(os.Stdin)

		// Step 1: List available groups
		groupNames := make([]string, 0, len(cfg.Groups))
		for name := range cfg.Groups {
			groupNames = append(groupNames, name)
		}
		sort.Strings(groupNames)

		if len(groupNames) == 0 {
			fmt.Println("No groups found")
			return nil
		}

		// Display groups
		fmt.Println("Available groups:")
		fmt.Println()
		for i, name := range groupNames {
			group, _ := cfg.GetGroup(name)
			projects, _ := group.GetProjectPaths()
			fmt.Printf("  %d. %s (%d projects)\n", i+1, name, len(projects))
		}
		fmt.Println()

		// Prompt for group selection
		fmt.Print("Select group number (or press Enter to cancel): ")
		selection, _ := reader.ReadString('\n')
		selection = strings.TrimSpace(selection)

		if selection == "" {
			fmt.Println("Cancelled")
			return nil
		}

		idx, err := strconv.Atoi(selection)
		if err != nil || idx < 1 || idx > len(groupNames) {
			return fmt.Errorf("invalid selection: %s", selection)
		}
		groupName = groupNames[idx-1]

		// Step 2: List projects in selected group
		group, err := cfg.GetGroup(groupName)
		if err != nil {
			return err
		}

		projects, err := group.GetProjectPaths()
		if err != nil {
			return fmt.Errorf("failed to get project paths: %w", err)
		}

		if len(projects) == 0 {
			fmt.Printf("Group '%s' has no projects\n", groupName)
			return nil
		}

		fmt.Println()
		fmt.Printf("Projects in group '%s':\n", groupName)
		fmt.Println()
		for i, proj := range projects {
			fmt.Printf("  %d. %s (current priority: %d)\n", i+1, proj.Alias, proj.Priority)
		}
		fmt.Println()

		// Step 3: Prompt for priority order
		fmt.Println("Enter project numbers in priority order (highest to lowest).")
		fmt.Println("Separate numbers with spaces or commas (e.g., '1 3 2' or '1,3,2').")
		fmt.Println("Press Enter to use current order, or 'cancel' to abort.")
		fmt.Print("Priority order: ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			fmt.Println("Using current priority order")
			return nil
		}

		if strings.ToLower(input) == "cancel" {
			fmt.Println("Cancelled")
			return nil
		}

		// Parse input (support both space and comma separators)
		input = strings.ReplaceAll(input, ",", " ")
		parts := strings.Fields(input)

		selectedIndices := make([]int, 0, len(parts))
		for _, part := range parts {
			idx, err := strconv.Atoi(part)
			if err != nil || idx < 1 || idx > len(projects) {
				return fmt.Errorf("invalid project number: %s", part)
			}
			selectedIndices = append(selectedIndices, idx-1)
		}

		// Check for duplicates
		seen := make(map[int]bool)
		for _, idx := range selectedIndices {
			if seen[idx] {
				return fmt.Errorf("duplicate project number: %d", idx+1)
			}
			seen[idx] = true
		}

		// Build aliases list from selected indices
		aliases = make([]string, 0, len(selectedIndices))
		for _, idx := range selectedIndices {
			aliases = append(aliases, projects[idx].Alias)
		}

	default:
		// Argument mode: 1+ arguments provided
		if len(args) < 2 {
			return fmt.Errorf("argument mode requires at least 2 arguments: <group> <alias1> [alias2]...")
		}
		groupName = args[0]
		aliases = args[1:]
	}

	if err := cfg.SetPriority(groupName, aliases); err != nil {
		return err
	}

	if err := cfg.Save(cfgFile); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println()
	fmt.Printf("✓ Updated priority for group '%s'\n", groupName)
	fmt.Println("  Priority order:")
	for i, alias := range aliases {
		fmt.Printf("  [%d] %s\n", i+1, alias)
	}
	return nil
}
