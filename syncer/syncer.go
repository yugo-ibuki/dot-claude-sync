package syncer

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/yugo-ibuki/dot-claude-sync/config"
	"github.com/yugo-ibuki/dot-claude-sync/utils"
)

// SyncResult represents the result of syncing to a single project
type SyncResult struct {
	Project     string  // Project alias
	NewFiles    int     // Number of new files added
	Overwritten int     // Number of existing files overwritten
	Failed      int     // Number of failed operations
	Errors      []error // List of errors encountered
	Skipped     bool    // Whether the project was skipped
	SkipReason  string  // Reason for skipping
}

// OverwriteInfo holds information about files that will be overwritten
type OverwriteInfo struct {
	Project string // Project alias
	RelPath string // Relative path of the file
}

// SyncFiles distributes resolved files to all projects
func SyncFiles(resolved []ResolvedFile, projects []config.ProjectPath, dryRun bool, verbose bool, force bool) ([]SyncResult, error) {
	if len(resolved) == 0 {
		return nil, fmt.Errorf("no files to sync")
	}

	// Collect files that would be overwritten
	var overwriteInfo []OverwriteInfo
	for _, project := range projects {
		claudeDir := expandPath(project.Path)
		if !utils.FileExists(claudeDir) {
			continue
		}

		for _, file := range resolved {
			dstPath := filepath.Join(claudeDir, file.RelPath)
			if utils.FileExists(dstPath) {
				overwriteInfo = append(overwriteInfo, OverwriteInfo{
					Project: project.Alias,
					RelPath: file.RelPath,
				})
			}
		}
	}

	// Show warning and ask for confirmation if overwrites would occur
	if len(overwriteInfo) > 0 && !dryRun && !force {
		fmt.Println("\n⚠️  Warning: The following files will be overwritten:")
		fmt.Println()

		// Group by project
		byProject := make(map[string][]string)
		for _, info := range overwriteInfo {
			byProject[info.Project] = append(byProject[info.Project], info.RelPath)
		}

		for project, files := range byProject {
			fmt.Printf("  %s:\n", project)
			for _, file := range files {
				fmt.Printf("    - %s\n", file)
			}
		}

		fmt.Println()
		if !utils.Confirm("Do you want to continue?") {
			return nil, fmt.Errorf("sync cancelled by user")
		}
		fmt.Println()
	}

	var results []SyncResult

	for _, project := range projects {
		result := syncToProject(resolved, project, dryRun, verbose)
		results = append(results, result)
	}

	return results, nil
}

// syncToProject syncs files to a single project
func syncToProject(resolved []ResolvedFile, project config.ProjectPath, dryRun bool, verbose bool) SyncResult {
	result := SyncResult{
		Project: project.Alias,
		Errors:  []error{},
	}

	claudeDir := expandPath(project.Path)

	// Check if .claude directory exists
	if !utils.FileExists(claudeDir) {
		result.Skipped = true
		result.SkipReason = fmt.Sprintf(".claude directory does not exist: %s", claudeDir)
		return result
	}

	// Sync each resolved file
	for _, file := range resolved {
		dstPath := filepath.Join(claudeDir, file.RelPath)

		// Check if destination file already exists
		fileExists := utils.FileExists(dstPath)

		if dryRun {
			if verbose {
				if fileExists {
					fmt.Printf("  [DRY RUN] Would overwrite: %s\n", file.RelPath)
				} else {
					fmt.Printf("  [DRY RUN] Would create: %s\n", file.RelPath)
				}
			}

			if fileExists {
				result.Overwritten++
			} else {
				result.NewFiles++
			}
			continue
		}

		// Actual file copy
		if err := utils.CopyFile(file.AbsPath, dstPath); err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Errorf("%s: %w", file.RelPath, err))
			if verbose {
				fmt.Fprintf(os.Stderr, "  ✗ Failed to sync %s: %v\n", file.RelPath, err)
			}
			continue
		}

		if fileExists {
			result.Overwritten++
			if verbose {
				fmt.Printf("  ✓ Overwritten: %s\n", file.RelPath)
			}
		} else {
			result.NewFiles++
			if verbose {
				fmt.Printf("  ✓ Created: %s\n", file.RelPath)
			}
		}
	}

	return result
}

// GetSyncSummary returns a formatted summary of sync results
func GetSyncSummary(results []SyncResult) string {
	totalNew := 0
	totalOverwritten := 0
	totalFailed := 0
	successfulProjects := 0
	skippedProjects := 0

	for _, result := range results {
		if result.Skipped {
			skippedProjects++
			continue
		}

		totalNew += result.NewFiles
		totalOverwritten += result.Overwritten
		totalFailed += result.Failed

		if result.Failed == 0 {
			successfulProjects++
		}
	}

	summary := "\nSummary:\n"
	summary += fmt.Sprintf("  Projects: %d successful", successfulProjects)
	if skippedProjects > 0 {
		summary += fmt.Sprintf(", %d skipped", skippedProjects)
	}
	if totalFailed > 0 {
		summary += fmt.Sprintf(", %d with failures", countFailedProjects(results))
	}
	summary += "\n"

	summary += fmt.Sprintf("  Files: %d new, %d overwritten", totalNew, totalOverwritten)
	if totalFailed > 0 {
		summary += fmt.Sprintf(", %d failed", totalFailed)
	}
	summary += "\n"

	return summary
}

// countFailedProjects counts projects that had failures
func countFailedProjects(results []SyncResult) int {
	count := 0
	for _, result := range results {
		if result.Failed > 0 {
			count++
		}
	}
	return count
}

// PrintSyncResults prints detailed sync results for each project
func PrintSyncResults(results []SyncResult, verbose bool) {
	for _, result := range results {
		if result.Skipped {
			fmt.Printf("✗ %s: skipped (%s)\n", result.Project, result.SkipReason)
			continue
		}

		if result.Failed > 0 {
			fmt.Printf("✗ %s: %d new, %d overwritten, %d failed\n",
				result.Project, result.NewFiles, result.Overwritten, result.Failed)
			if verbose {
				for _, err := range result.Errors {
					fmt.Fprintf(os.Stderr, "    Error: %v\n", err)
				}
			}
		} else {
			status := ""
			if result.NewFiles > 0 {
				status += fmt.Sprintf("%d new", result.NewFiles)
			}
			if result.Overwritten > 0 {
				if status != "" {
					status += ", "
				}
				status += fmt.Sprintf("%d overwritten", result.Overwritten)
			}
			if status == "" {
				status = "no changes"
			}

			fmt.Printf("✓ %s: %s\n", result.Project, status)
		}
	}
}

// HasErrors returns true if any sync operation had errors
func HasErrors(results []SyncResult) bool {
	for _, result := range results {
		if result.Failed > 0 {
			return true
		}
	}
	return false
}

// GetTotalFiles returns the total number of files synced
func GetTotalFiles(results []SyncResult) int {
	total := 0
	for _, result := range results {
		total += result.NewFiles + result.Overwritten
	}
	return total
}
