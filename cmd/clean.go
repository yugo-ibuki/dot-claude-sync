package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/yugo-ibuki/dot-claude-sync/utils"
)

var cleanCmd = &cobra.Command{
	Use:   "clean [path]",
	Short: "Delete all empty folders recursively",
	Long: `Delete all empty directories recursively from the specified path.
This command walks through the directory tree and removes all empty folders.
It uses post-order traversal to ensure empty parent directories are also deleted.

Example:
  dot-claude-sync clean ~/projects
  dot-claude-sync clean .`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		targetPath := args[0]

		if verbose {
			fmt.Printf("Scanning directory: %s\n", targetPath)
		}

		// Delete empty folders
		deletedFolders, err := utils.DeleteEmptyFolders(targetPath)
		if err != nil {
			return fmt.Errorf("failed to delete empty folders: %w", err)
		}

		// Display results
		if len(deletedFolders) == 0 {
			fmt.Println("No empty folders found.")
		} else {
			fmt.Printf("Successfully deleted %d empty folder(s):\n", len(deletedFolders))
			for _, folder := range deletedFolders {
				fmt.Printf("  - %s\n", folder)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}
