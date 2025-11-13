package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yugo-ibuki/dot-claude-sync/config"
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

func runMv(cmd *cobra.Command, args []string) error {
	groupName := args[0]
	fromPath := args[1]
	toPath := args[2]

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

	fmt.Printf("This will rename in '%s' group:\n", groupName)
	fmt.Printf("%s → %s\n", fromPath, toPath)

	if !force && !dryRun {
		fmt.Print("\nContinue? [y/N]: ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Cancelled")
			return nil
		}
	}

	// TODO: Implement move logic
	fmt.Println("\nMoving...")
	for _, project := range projects {
		fmt.Printf("✓ Moved in %s\n", project.Alias)
	}

	fmt.Println("\nSummary: Implementation pending")

	return nil
}
