package syncer

import (
	"fmt"
	"sort"
	"strings"
	"time"
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
// folderFilter: list of folder names to resolve by modification time only (ignoring priority)
func ResolveConflicts(files []FileInfo, folderFilter []string) ([]ResolvedFile, []Conflict, error) {
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
			// Check if this file is in a folder that should ignore priority
			isInFilteredFolder := isFileInFilteredFolder(relPath, folderFilter)

			var winner FileInfo
			if isInFilteredFolder {
				// Use modification time only (ignore priority)
				winner = resolveConflictByModTime(candidates)
			} else {
				// Use standard resolution (modification time + priority)
				winner = resolveConflict(candidates)
			}

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

// resolveConflict selects the file based on modification time (newest wins)
// If multiple files have the same timestamp (within 1 second), priority is used as fallback
func resolveConflict(candidates []FileInfo) FileInfo {
	if len(candidates) == 0 {
		panic("resolveConflict called with empty candidates")
	}

	// Find the candidate with the latest modification time
	latest := candidates[0]
	for _, candidate := range candidates[1:] {
		if candidate.ModTime.After(latest.ModTime) {
			latest = candidate
		}
	}

	// Find all candidates with timestamps within 1 second of the latest
	threshold := latest.ModTime.Add(-1 * time.Second)
	var recent []FileInfo
	for _, candidate := range candidates {
		if candidate.ModTime.After(threshold) || candidate.ModTime.Equal(latest.ModTime) {
			recent = append(recent, candidate)
		}
	}

	// If multiple files have similar timestamps, use priority as fallback
	if len(recent) > 1 {
		winner := recent[0]
		for _, candidate := range recent[1:] {
			if candidate.Priority < winner.Priority {
				winner = candidate
			}
		}
		return winner
	}

	return latest
}

// resolveConflictByModTime selects the file based on modification time only (ignoring priority)
// This is used for folders in folderFilter where priority should be ignored
func resolveConflictByModTime(candidates []FileInfo) FileInfo {
	if len(candidates) == 0 {
		panic("resolveConflictByModTime called with empty candidates")
	}

	// Find the candidate with the latest modification time
	winner := candidates[0]
	for _, candidate := range candidates[1:] {
		if candidate.ModTime.After(winner.ModTime) {
			winner = candidate
		}
	}

	return winner
}

// isFileInFilteredFolder checks if a file's relative path is in any of the filtered folders
func isFileInFilteredFolder(relPath string, folderFilter []string) bool {
	if len(folderFilter) == 0 {
		return false
	}

	for _, folder := range folderFilter {
		// Normalize folder name (remove trailing slashes)
		folder = strings.TrimSuffix(folder, "/")

		// Check if the file starts with the folder path
		if relPath == folder || strings.HasPrefix(relPath, folder+"/") {
			return true
		}
	}

	return false
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
