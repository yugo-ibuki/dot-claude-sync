package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/yugo-ibuki/dot-claude-sync/config"
	"github.com/yugo-ibuki/dot-claude-sync/utils"
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

type deleteTarget struct {
	project  config.ProjectPath
	fullPath string
	exists   bool
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

	// Search for files in all projects
	var targets []deleteTarget
	for _, project := range projects {
		fullPath := filepath.Join(project.Path, targetPath)
		exists := utils.FileExists(fullPath)

		targets = append(targets, deleteTarget{
			project:  project,
			fullPath: fullPath,
			exists:   exists,
		})
	}

	// Display targets
	fmt.Printf("This will delete from '%s' group:\n", groupName)
	existsCount := 0
	for _, target := range targets {
		if target.exists {
			existsCount++
			fileType := "file"
			if utils.IsDirectory(target.fullPath) {
				fileType = "directory"
			}
			fmt.Printf("- %s (%s)\n", target.fullPath, fileType)
		}
	}

	if existsCount == 0 {
		fmt.Println("\nNo files found to delete")
		return nil
	}

	// Confirmation prompt
	if !force && !dryRun {
		fmt.Print("\nContinue? [y/N]: ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Cancelled")
			return nil
		}
	}

	// Execute deletion
	if !dryRun {
		fmt.Println("\nDeleting...")
	} else {
		fmt.Println("\nWould delete:")
	}

	successCount := 0
	skipCount := 0
	failCount := 0

	for _, target := range targets {
		if !target.exists {
			if verbose {
				fmt.Printf("✗ Not found in %s (skipped)\n", target.project.Alias)
			}
			skipCount++
			continue
		}

		if !dryRun {
			err := utils.RemoveFile(target.fullPath)
			if err != nil {
				fmt.Printf("✗ Failed to delete from %s: %v\n", target.project.Alias, err)
				failCount++
				continue
			}
		}

		fmt.Printf("✓ Deleted from %s\n", target.project.Alias)
		successCount++
	}

	// Summary
	fmt.Printf("\nSummary: %d deleted", successCount)
	if skipCount > 0 {
		fmt.Printf(", %d skipped", skipCount)
	}
	if failCount > 0 {
		fmt.Printf(", %d failed", failCount)
	}
	fmt.Println()

	if failCount > 0 {
		return fmt.Errorf("some deletions failed")
	}

	return nil
}
