package cmd

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/yugo-ibuki/dot-claude-sync/config"
)

var listCmd = &cobra.Command{
	Use:   "list [group]",
	Short: "List all groups or details of a specific group",
	Long: `List all available groups in the configuration,
or show detailed information about a specific group.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	if verbose {
		fmt.Printf("Loading configuration...\n")
	}

	cfg, err := config.Load(cfgFile)
	if err != nil {
		return err
	}

	// If no group specified, list all groups
	if len(args) == 0 {
		groups := cfg.ListGroups()
		sort.Strings(groups)

		fmt.Println("Groups:")
		for _, groupName := range groups {
			group, _ := cfg.GetGroup(groupName)
			projects, _ := group.GetProjectPaths()
			fmt.Printf("  %s (%d projects)\n", groupName, len(projects))
		}
		return nil
	}

	// Show specific group details
	groupName := args[0]
	group, err := cfg.GetGroup(groupName)
	if err != nil {
		availableGroups := cfg.ListGroups()
		return fmt.Errorf("%w\nAvailable groups: %v", err, availableGroups)
	}

	projects, err := group.GetProjectPaths()
	if err != nil {
		return fmt.Errorf("failed to parse group paths: %w", err)
	}

	// Sort by priority
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].Priority < projects[j].Priority
	})

	fmt.Printf("Group: %s\n", groupName)
	fmt.Println("Priority order:")
	for _, project := range projects {
		priorityLabel := fmt.Sprintf("priority: %d", project.Priority)
		if project.Priority == len(projects) {
			priorityLabel = "default priority"
		}
		fmt.Printf("  %d. %s: %s (%s)\n", project.Priority, project.Alias, project.Path, priorityLabel)
	}

	return nil
}
