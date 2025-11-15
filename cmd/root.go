package cmd

import (
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	dryRun  bool
	verbose bool
	force   bool
)

var rootCmd = &cobra.Command{
	Use:   "claude-sync",
	Short: "Sync .claude directories across multiple projects",
	Long: `A CLI tool to synchronize .claude directories across multiple projects.
Manage groups of projects and perform batch operations like add, overwrite, delete, and move files.`,
	Version: "0.1.2",
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path (default: search for .claude-sync.yaml)")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "simulate execution without making changes")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "enable verbose logging")
	rootCmd.PersistentFlags().BoolVar(&force, "force", false, "skip confirmation prompts")
}
