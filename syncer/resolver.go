package syncer

import (
	"fmt"
	"sort"
)

// ResolvedFile represents a file after conflict resolution
type ResolvedFile struct {
	RelPath  string // Relative path from .claude directory
	AbsPath  string // Absolute path to the source file
	Source   string // Source project alias
	Priority int    // Priority of the source project
}

// Conflict represents a conflict between multiple files with the same path
type Conflict struct {
	RelPath    string     // The conflicting relative path
	Candidates []FileInfo // All candidate files
	Resolved   FileInfo   // The resolved file (highest priority)
}

// ResolveConflicts resolves conflicts between files based on priority
func ResolveConflicts(files []FileInfo) ([]ResolvedFile, []Conflict, error) {
	if len(files) == 0 {
		return nil, nil, fmt.Errorf("no files to resolve")
	}

	// Group files by relative path
	grouped := GroupFilesByRelPath(files)

	var resolved []ResolvedFile
	var conflicts []Conflict

	// Process each group
	for relPath, candidates := range grouped {
		if len(candidates) == 1 {
			// No conflict - single file
			file := candidates[0]
			resolved = append(resolved, ResolvedFile{
				RelPath:  file.RelPath,
				AbsPath:  file.AbsPath,
				Source:   file.Project,
				Priority: file.Priority,
			})
		} else {
			// Conflict - multiple files with same path
			winner := resolveConflict(candidates)
			resolved = append(resolved, ResolvedFile{
				RelPath:  winner.RelPath,
				AbsPath:  winner.AbsPath,
				Source:   winner.Project,
				Priority: winner.Priority,
			})

			conflicts = append(conflicts, Conflict{
				RelPath:    relPath,
				Candidates: candidates,
				Resolved:   winner,
			})
		}
	}

	// Sort resolved files by relative path for consistent output
	sort.Slice(resolved, func(i, j int) bool {
		return resolved[i].RelPath < resolved[j].RelPath
	})

	// Sort conflicts by relative path for consistent output
	sort.Slice(conflicts, func(i, j int) bool {
		return conflicts[i].RelPath < conflicts[j].RelPath
	})

	return resolved, conflicts, nil
}

// resolveConflict selects the file with the highest priority (lowest priority number)
func resolveConflict(candidates []FileInfo) FileInfo {
	if len(candidates) == 0 {
		panic("resolveConflict called with empty candidates")
	}

	// Find the candidate with the lowest priority number (highest priority)
	winner := candidates[0]
	for _, candidate := range candidates[1:] {
		if candidate.Priority < winner.Priority {
			winner = candidate
		}
	}

	return winner
}

// GetConflictSummary returns a formatted summary of conflicts
func GetConflictSummary(conflicts []Conflict) string {
	if len(conflicts) == 0 {
		return "No conflicts detected"
	}

	summary := fmt.Sprintf("%d conflict(s) resolved:\n", len(conflicts))
	for _, conflict := range conflicts {
		summary += fmt.Sprintf("  - %s: using %s (priority: %d)\n",
			conflict.RelPath,
			conflict.Resolved.Project,
			conflict.Resolved.Priority)
	}

	return summary
}

// GetResolvedSummary returns a summary of resolved files
func GetResolvedSummary(resolved []ResolvedFile) string {
	if len(resolved) == 0 {
		return "No files resolved"
	}

	// Group by source project
	byProject := make(map[string]int)
	for _, file := range resolved {
		byProject[file.Source]++
	}

	summary := fmt.Sprintf("%d unique file(s) resolved\n", len(resolved))
	summary += "Files by source project:\n"

	// Sort project names for consistent output
	var projects []string
	for project := range byProject {
		projects = append(projects, project)
	}
	sort.Strings(projects)

	for _, project := range projects {
		count := byProject[project]
		summary += fmt.Sprintf("  - %s: %d file(s)\n", project, count)
	}

	return summary
}

// HasConflicts returns true if there are any conflicts
func HasConflicts(conflicts []Conflict) bool {
	return len(conflicts) > 0
}

// GetConflictCount returns the number of conflicts
func GetConflictCount(conflicts []Conflict) int {
	return len(conflicts)
}

// GetResolvedCount returns the number of resolved files
func GetResolvedCount(resolved []ResolvedFile) int {
	return len(resolved)
}
