package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yugo-ibuki/dot-claude-sync/config"
	"github.com/yugo-ibuki/dot-claude-sync/syncer"
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
		fmt.Println()
	}

	// Phase 1: Collect files
	fmt.Printf("Collecting files from group '%s'...\n", groupName)

	allFiles, err := syncer.CollectFiles(projects)
	if err != nil {
		return fmt.Errorf("failed to collect files: %w", err)
	}

	// Show collection results
	filesByProject := make(map[string]int)
	for _, file := range allFiles {
		filesByProject[file.Project]++
	}

	for _, project := range projects {
		count := filesByProject[project.Alias]
		if count > 0 {
			fmt.Printf("✓ %s: %d file(s) (priority: %d)\n", project.Alias, count, project.Priority)
		} else {
			fmt.Printf("✗ %s: no files found (priority: %d)\n", project.Alias, project.Priority)
		}
	}

	if len(allFiles) == 0 {
		fmt.Println("\nNo files to sync")
		return nil
	}

	// Phase 2: Resolve conflicts
	fmt.Println("\nResolving conflicts...")

	resolved, conflicts, err := syncer.ResolveConflicts(allFiles)
	if err != nil {
		return fmt.Errorf("failed to resolve conflicts: %w", err)
	}

	if len(conflicts) > 0 {
		for _, conflict := range conflicts {
			fmt.Printf("- %s: using %s (priority: %d)\n",
				conflict.RelPath,
				conflict.Resolved.Project,
				conflict.Resolved.Priority)
		}
	} else {
		fmt.Println("No conflicts detected")
	}

	if verbose {
		fmt.Printf("\nTotal files to sync: %d\n", len(resolved))
	}

	// Phase 3: Sync files
	fmt.Println("\nSyncing...")

	results, err := syncer.SyncFiles(resolved, projects, dryRun, verbose)
	if err != nil {
		return fmt.Errorf("failed to sync files: %w", err)
	}

	if !verbose {
		syncer.PrintSyncResults(results, verbose)
	}

	// Print summary
	fmt.Print(syncer.GetSyncSummary(results))

	// Exit with error if any sync operations failed
	if syncer.HasErrors(results) {
		return fmt.Errorf("some sync operations failed")
	}

	return nil
}
