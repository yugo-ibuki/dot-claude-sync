package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/yugo-ibuki/dot-claude-sync/config"
	"github.com/yugo-ibuki/dot-claude-sync/utils"
)

var mvCmd = &cobra.Command{
	Use:   "mv <group> <from> <to>",
	Short: "Move or rename a file or directory in all projects in a group",
	Long: `Move or rename the specified file or directory in all projects in the group.
Prompts for confirmation unless --force is specified.`,
	Args: cobra.ExactArgs(3),
	RunE: runMv,
}

func init() {
	rootCmd.AddCommand(mvCmd)
}

// MoveResult represents the result of moving files in a project
type MoveResult struct {
	Project    string
	Moved      bool
	Skipped    bool
	SkipReason string
	Error      error
}

func runMv(cmd *cobra.Command, args []string) error {
	groupName := args[0]
	fromPath := args[1]
	toPath := args[2]

	// Prevent moving bk directory (backup directory)
	if fromPath == "bk" || filepath.Clean(fromPath) == "bk" {
		return fmt.Errorf("cannot move 'bk' directory: it is reserved for backups")
	}
	if toPath == "bk" || filepath.Clean(toPath) == "bk" {
		return fmt.Errorf("cannot move to 'bk' directory: it is reserved for backups")
	}

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

	// Check which projects have the source file/directory
	var foundProjects []config.ProjectPath
	var notFoundProjects []config.ProjectPath

	for _, project := range projects {
		claudeDir := expandPath(project.Path)
		srcFullPath := filepath.Join(claudeDir, fromPath)

		if utils.FileExists(srcFullPath) {
			foundProjects = append(foundProjects, project)
		} else {
			notFoundProjects = append(notFoundProjects, project)
		}
	}

	if len(foundProjects) == 0 {
		fmt.Printf("Source path '%s' not found in any project in group '%s'\n", fromPath, groupName)
		return nil
	}

	if dryRun {
		fmt.Println("DRY RUN MODE - No changes will be made")
		fmt.Println()
	}

	// Show what will be moved
	fmt.Printf("This will rename in '%s' group:\n", groupName)
	fmt.Printf("%s → %s\n", fromPath, toPath)
	fmt.Println()
	fmt.Printf("Found in %d project(s):\n", len(foundProjects))
	for _, project := range foundProjects {
		fmt.Printf("  - %s\n", project.Alias)
	}

	if len(notFoundProjects) > 0 {
		fmt.Printf("\nNot found in %d project(s) (will be skipped):\n", len(notFoundProjects))
		for _, project := range notFoundProjects {
			fmt.Printf("  - %s\n", project.Alias)
		}
	}

	// Confirmation prompt
	if !force && !dryRun {
		fmt.Print("\nContinue? [y/N]: ")
		var response string
		if _, err := fmt.Scanln(&response); err != nil {
			fmt.Println("Cancelled")
			return nil
		}
		if response != "y" && response != "Y" {
			fmt.Println("Cancelled")
			return nil
		}
	}

	// Execute move operation
	fmt.Println("\nMoving...")
	var results []MoveResult

	for _, project := range projects {
		result := moveInProject(project, fromPath, toPath, dryRun, verbose)
		results = append(results, result)
	}

	// Print results
	fmt.Println()
	movedCount := 0
	skippedCount := 0
	failedCount := 0

	for _, result := range results {
		switch {
		case result.Skipped:
			skippedCount++
			if verbose {
				fmt.Printf("✗ %s: %s\n", result.Project, result.SkipReason)
			}
		case result.Error != nil:
			failedCount++
			fmt.Fprintf(os.Stderr, "✗ %s: %v\n", result.Project, result.Error)
		case result.Moved:
			movedCount++
			fmt.Printf("✓ %s: moved\n", result.Project)
		}
	}

	// Print summary
	fmt.Printf("\nSummary: %d moved", movedCount)
	if skippedCount > 0 {
		fmt.Printf(", %d skipped", skippedCount)
	}
	if failedCount > 0 {
		fmt.Printf(", %d failed", failedCount)
	}
	fmt.Println()

	if failedCount > 0 {
		return fmt.Errorf("some move operations failed")
	}

	return nil
}

// moveInProject moves a file or directory in a single project
func moveInProject(project config.ProjectPath, fromPath, toPath string, dryRun, verbose bool) MoveResult {
	result := MoveResult{
		Project: project.Alias,
	}

	claudeDir := expandPath(project.Path)
	srcFullPath := filepath.Join(claudeDir, fromPath)
	dstFullPath := filepath.Join(claudeDir, toPath)

	// Check if source exists
	if !utils.FileExists(srcFullPath) {
		result.Skipped = true
		result.SkipReason = "source not found"
		return result
	}

	// Check if destination already exists
	if utils.FileExists(dstFullPath) {
		result.Skipped = true
		result.SkipReason = "destination already exists"
		return result
	}

	if dryRun {
		result.Moved = true
		if verbose {
			fmt.Printf("  [DRY RUN] %s: would move %s → %s\n", project.Alias, fromPath, toPath)
		}
		return result
	}

	// Perform the move
	if err := utils.MoveFile(srcFullPath, dstFullPath); err != nil {
		result.Error = err
		return result
	}

	result.Moved = true
	if verbose {
		fmt.Printf("  %s: moved %s → %s\n", project.Alias, fromPath, toPath)
	}

	return result
}

// expandPath expands ~ to home directory
func expandPath(path string) string {
	if len(path) == 0 || path[0] != '~' {
		return path
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return path
	}

	if len(path) == 1 {
		return homeDir
	}

	if path[1] == '/' || path[1] == filepath.Separator {
		return filepath.Join(homeDir, path[2:])
	}

	return filepath.Join(homeDir, path[1:])
}
