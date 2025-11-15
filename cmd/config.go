package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yugo-ibuki/dot-claude-sync/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long:  `Manage groups and projects in the configuration file.`,
}

var configShowCmd = &cobra.Command{
	Use:   "show [group]",
	Short: "Show configuration",
	Long:  `Display the entire configuration or details of a specific group.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runConfigShow,
}

var configAddGroupCmd = &cobra.Command{
	Use:   "add-group <name>",
	Short: "Add a new group",
	Long:  `Add a new group to the configuration.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runConfigAddGroup,
}

var configRemoveGroupCmd = &cobra.Command{
	Use:   "remove-group <name>",
	Short: "Remove a group",
	Long:  `Remove a group from the configuration.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runConfigRemoveGroup,
}

var configAddProjectCmd = &cobra.Command{
	Use:   "add-project <group> <alias> <path>",
	Short: "Add a project to a group",
	Long:  `Add a new project to an existing group.`,
	Args:  cobra.ExactArgs(3),
	RunE:  runConfigAddProject,
}

var configRemoveProjectCmd = &cobra.Command{
	Use:   "remove-project <group> <alias>",
	Short: "Remove a project from a group",
	Long:  `Remove a project from an existing group.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runConfigRemoveProject,
}

var configSetPriorityCmd = &cobra.Command{
	Use:   "set-priority <group> <alias1> [alias2] [alias3]...",
	Short: "Set priority order for a group",
	Long:  `Set the priority order for projects in a group. First alias has highest priority.`,
	Args:  cobra.MinimumNArgs(2),
	RunE:  runConfigSetPriority,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configAddGroupCmd)
	configCmd.AddCommand(configRemoveGroupCmd)
	configCmd.AddCommand(configAddProjectCmd)
	configCmd.AddCommand(configRemoveProjectCmd)
	configCmd.AddCommand(configSetPriorityCmd)
}

func runConfigShow(cmd *cobra.Command, args []string) error {
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
	groupName := args[0]

	cfg, err := config.Load(cfgFile)
	if err != nil {
		return err
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
	groupName := args[0]
	alias := args[1]
	path := args[2]

	cfg, err := config.Load(cfgFile)
	if err != nil {
		return err
	}

	if err := cfg.AddProject(groupName, alias, path); err != nil {
		return err
	}

	if err := cfg.Save(cfgFile); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("✓ Added project '%s' to group '%s'\n", alias, groupName)
	fmt.Printf("  Path: %s\n", path)
	return nil
}

func runConfigRemoveProject(cmd *cobra.Command, args []string) error {
	groupName := args[0]
	alias := args[1]

	cfg, err := config.Load(cfgFile)
	if err != nil {
		return err
	}

	if err := cfg.RemoveProject(groupName, alias); err != nil {
		return err
	}

	if err := cfg.Save(cfgFile); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("✓ Removed project '%s' from group '%s'\n", alias, groupName)
	return nil
}

func runConfigSetPriority(cmd *cobra.Command, args []string) error {
	groupName := args[0]
	aliases := args[1:]

	cfg, err := config.Load(cfgFile)
	if err != nil {
		return err
	}

	if err := cfg.SetPriority(groupName, aliases); err != nil {
		return err
	}

	if err := cfg.Save(cfgFile); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("✓ Updated priority for group '%s'\n", groupName)
	fmt.Println("  Priority order:")
	for i, alias := range aliases {
		fmt.Printf("  [%d] %s\n", i+1, alias)
	}
	return nil
}
