package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/yugo-ibuki/dot-claude-sync/config"
	"github.com/yugo-ibuki/dot-claude-sync/utils"
)

var backupCmd = &cobra.Command{
	Use:   "backup <group>",
	Short: "Backup .claude files to bk directory in all projects in a group",
	Long: `Create backups of .claude directories for all projects in the specified group.
Backups are stored in a timestamped subdirectory within each project's .claude/bk directory.
Example: .claude/bk/20250117-143025/`,
	Args: cobra.ExactArgs(1),
	RunE: runBackup,
}

func init() {
	rootCmd.AddCommand(backupCmd)
}

// BackupResult represents the result of backing up a project
type BackupResult struct {
	Project    string
	Success    bool
	Skipped    bool
	SkipReason string
	Error      error
	FileCount  int
}

func runBackup(cmd *cobra.Command, args []string) error {
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

	fmt.Printf("Creating backups for group '%s'...\n", groupName)

	var results []BackupResult
	timestamp := time.Now().Format("20060102-150405")

	for _, project := range projects {
		result := backupProject(project, timestamp, dryRun, verbose)
		results = append(results, result)
	}

	// Print results
	fmt.Println()
	successCount := 0
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
			fmt.Printf("✗ %s: %v\n", result.Project, result.Error)
		case result.Success:
			successCount++
			fmt.Printf("✓ %s: backed up %d files\n", result.Project, result.FileCount)
		}
	}

	// Print summary
	fmt.Printf("\nSummary: %d backed up", successCount)
	if skippedCount > 0 {
		fmt.Printf(", %d skipped", skippedCount)
	}
	if failedCount > 0 {
		fmt.Printf(", %d failed", failedCount)
	}
	fmt.Println()

	if failedCount > 0 {
		return fmt.Errorf("some backup operations failed")
	}

	return nil
}

// backupProject creates a backup of a single project's .claude directory
func backupProject(project config.ProjectPath, timestamp string, dryRun, verbose bool) BackupResult {
	result := BackupResult{
		Project: project.Alias,
	}

	claudeDir := expandPath(project.Path)

	// Check if .claude directory exists
	if !utils.FileExists(claudeDir) {
		result.Skipped = true
		result.SkipReason = ".claude directory does not exist"
		return result
	}

	if !utils.IsDirectory(claudeDir) {
		result.Skipped = true
		result.SkipReason = "path is not a directory"
		return result
	}

	// Create backup directory path with timestamp
	backupDir := filepath.Join(claudeDir, "bk", timestamp)

	if dryRun {
		if verbose {
			fmt.Printf("  [DRY RUN] Would backup %s to %s\n", claudeDir, backupDir)
		}
		result.Success = true
		result.FileCount = 0 // In dry run mode, we don't count files
		return result
	}

	// Create backup directory if it doesn't exist
	if err := utils.EnsureDir(backupDir); err != nil {
		result.Error = fmt.Errorf("failed to create backup directory: %w", err)
		return result
	}

	// Copy .claude directory contents to bk directory, excluding bk itself
	if err := utils.CopyDirExclude(claudeDir, backupDir, []string{"bk"}); err != nil {
		result.Error = fmt.Errorf("failed to copy files: %w", err)
		return result
	}

	// Count backed up files (optional, for reporting)
	fileCount, err := countFiles(backupDir)
	if err != nil {
		if verbose {
			fmt.Printf("  Warning: failed to count files: %v\n", err)
		}
		fileCount = 0
	}

	result.Success = true
	result.FileCount = fileCount

	return result
}

// countFiles counts the number of files in a directory recursively
func countFiles(dir string) (int, error) {
	count := 0
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			count++
		}
		return nil
	})
	return count, err
}
