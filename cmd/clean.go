package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/yugo-ibuki/dot-claude-sync/utils"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean up empty or redundant .claude/custom-document directories",
	Long: `Scan .claude/custom-document directories and identify those containing only empty files.
Displays a confirmation dialog listing directories to be deleted, then removes them after confirmation.

This command searches for directories that contain:
- Only empty files (0 bytes)
- Only empty subdirectories
- Combinations of the above

Example:
  dot-claude-sync clean`,
	RunE: func(cmd *cobra.Command, args []string) error {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}

		customDocPath := filepath.Join(homeDir, ".claude", "custom-document")

		if verbose {
			fmt.Printf("Scanning directory: %s\n", customDocPath)
		}

		// Check if custom-document directory exists
		if !utils.FileExists(customDocPath) {
			fmt.Println("No .claude/custom-document directory found.")
			return nil
		}

		// Find directories with only empty files
		candidateDirs, err := utils.FindDirectoriesWithOnlyEmptyFiles(customDocPath)
		if err != nil {
			return fmt.Errorf("failed to scan directory: %w", err)
		}

		if len(candidateDirs) == 0 {
			fmt.Println("No directories with only empty files found.")
			return nil
		}

		// Display directories to be deleted
		fmt.Printf("\nFound %d directory(ies) with only empty files:\n", len(candidateDirs))
		for i, dir := range candidateDirs {
			fmt.Printf("  %d. %s\n", i+1, dir)
		}

		// Ask for confirmation
		if !force && !utils.Confirm("\nDelete these directories?") {
			fmt.Println("Cancelled.")
			return nil
		}

		if dryRun {
			fmt.Printf("\nDRY RUN: Would delete %d directory(ies)\n", len(candidateDirs))
			return nil
		}

		// Delete the directories
		deletedCount := 0
		for _, dir := range candidateDirs {
			if err := os.RemoveAll(dir); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to delete %s: %v\n", dir, err)
			} else {
				deletedCount++
				if verbose {
					fmt.Printf("✓ Deleted: %s\n", dir)
				}
			}
		}

		fmt.Printf("\nSuccessfully deleted %d directory(ies)\n", deletedCount)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}
