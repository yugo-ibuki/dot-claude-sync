package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yugo-ibuki/dot-claude-sync/config"
)

var pushCmd = &cobra.Command{
	Use:   "push <group>",
	Short: "Sync .claude files across all projects in a group",
	Long: `Collect .claude files from all projects in the specified group,
resolve conflicts based on priority, and distribute to all projects.`,
	Args: cobra.ExactArgs(1),
	RunE: runPush,
}

func init() {
	rootCmd.AddCommand(pushCmd)
}

func runPush(cmd *cobra.Command, args []string) error {
	groupName := args[0]

	if verbose {
		fmt.Printf("Loading configuration...\n")
	}

	cfg, err := config.Load(cfgFile)
	if err != nil {
		return err
	}

	group, err := cfg.GetGroup(groupName)
	if err != nil {
		availableGroups := cfg.ListGroups()
		return fmt.Errorf("%w\nAvailable groups: %v", err, availableGroups)
	}

	projects, err := group.GetProjectPaths()
	if err != nil {
		return fmt.Errorf("failed to parse group paths: %w", err)
	}

	if dryRun {
		fmt.Println("DRY RUN MODE - No changes will be made")
	}

	fmt.Printf("Collecting files from group '%s'...\n", groupName)

	// TODO: Implement file collection logic
	for _, project := range projects {
		fmt.Printf("✓ %s: %s (priority: %d)\n", project.Alias, project.Path, project.Priority)
	}

	fmt.Println("\nResolving conflicts...")
	// TODO: Implement conflict resolution logic

	fmt.Println("\nSyncing...")
	// TODO: Implement sync logic

	fmt.Println("\nSummary: Implementation pending")

	return nil
}
