package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yugo-ibuki/dot-claude-sync/config"
)

var rmCmd = &cobra.Command{
	Use:   "rm <group> <path>",
	Short: "Remove a file or directory from all projects in a group",
	Long: `Delete the specified file or directory from all projects in the group.
Prompts for confirmation unless --force is specified.`,
	Args: cobra.ExactArgs(2),
	RunE: runRm,
}

func init() {
	rootCmd.AddCommand(rmCmd)
}

func runRm(cmd *cobra.Command, args []string) error {
	groupName := args[0]
	targetPath := args[1]

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

	fmt.Printf("This will delete from '%s' group:\n", groupName)

	// TODO: Implement file search and display logic
	for _, project := range projects {
		fmt.Printf("- %s/%s\n", project.Path, targetPath)
	}

	if !force && !dryRun {
		fmt.Print("\nContinue? [y/N]: ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Cancelled")
			return nil
		}
	}

	// TODO: Implement deletion logic
	fmt.Println("\nDeleting...")
	for _, project := range projects {
		fmt.Printf("✓ Deleted from %s\n", project.Alias)
	}

	fmt.Println("\nSummary: Implementation pending")

	return nil
}
