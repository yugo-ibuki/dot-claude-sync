package syncer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yugo-ibuki/dot-claude-sync/config"
)

// FileInfo represents information about a collected file
type FileInfo struct {
	RelPath  string    // Relative path from .claude directory
	AbsPath  string    // Absolute path to the file
	Project  string    // Project alias
	Priority int       // Project priority
	ModTime  time.Time // File modification time
}

// CollectFiles collects all files from .claude directories across projects
func CollectFiles(projects []config.ProjectPath, excludePatterns []string) ([]FileInfo, error) {
	var allFiles []FileInfo

	for _, project := range projects {
		files, err := collectFromProject(project, excludePatterns)
		if err != nil {
			// Don't fail the entire operation if one project fails
			fmt.Fprintf(os.Stderr, "Warning: failed to collect from %s: %v\n", project.Alias, err)
			continue
		}
		allFiles = append(allFiles, files...)
	}

	if len(allFiles) == 0 {
		return nil, fmt.Errorf("no files collected from any project")
	}

	return allFiles, nil
}

// collectFromProject collects files from a single project's .claude directory
func collectFromProject(project config.ProjectPath, excludePatterns []string) ([]FileInfo, error) {
	claudeDir := expandPath(project.Path)

	// Check if .claude directory exists
	info, err := os.Stat(claudeDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf(".claude directory does not exist: %s", claudeDir)
		}
		return nil, fmt.Errorf("failed to stat .claude directory: %w", err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", claudeDir)
	}

	var files []FileInfo

	// Walk through the .claude directory
	err = filepath.Walk(claudeDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the .claude directory itself
		if path == claudeDir {
			return nil
		}

		// Skip bk directory (backup directory)
		if info.IsDir() && info.Name() == "bk" {
			return filepath.SkipDir
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Calculate relative path from .claude directory
		relPath, err := filepath.Rel(claudeDir, path)
		if err != nil {
			return fmt.Errorf("failed to calculate relative path: %w", err)
		}

		// Normalize path separators to forward slashes
		relPath = filepath.ToSlash(relPath)

		// Check if file matches any exclude pattern
		if shouldExclude(relPath, excludePatterns) {
			return nil
		}

		files = append(files, FileInfo{
			RelPath:  relPath,
			AbsPath:  path,
			Project:  project.Alias,
			Priority: project.Priority,
			ModTime:  info.ModTime(),
		})

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return files, nil
}

// shouldExclude checks if a file path matches any of the exclude patterns
func shouldExclude(relPath string, patterns []string) bool {
	if len(patterns) == 0 {
		return false
	}

	for _, pattern := range patterns {
		// Try matching the full relative path
		matched, err := filepath.Match(pattern, relPath)
		if err != nil {
			// Invalid pattern, skip it
			continue
		}
		if matched {
			return true
		}

		// Also try matching against the base name
		baseName := filepath.Base(relPath)
		matched, err = filepath.Match(pattern, baseName)
		if err != nil {
			continue
		}
		if matched {
			return true
		}

		// Check if pattern matches any directory component
		// For patterns like "temp/*"
		if strings.Contains(pattern, "/") {
			matched, err = filepath.Match(pattern, relPath)
			if err == nil && matched {
				return true
			}
		}
	}

	return false
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

	// Handle both ~/ and ~ followed by path separator
	if path[1] == '/' || path[1] == filepath.Separator {
		return filepath.Join(homeDir, path[2:])
	}

	return filepath.Join(homeDir, path[1:])
}

// GroupFilesByRelPath groups files by their relative path
func GroupFilesByRelPath(files []FileInfo) map[string][]FileInfo {
	grouped := make(map[string][]FileInfo)

	for _, file := range files {
		grouped[file.RelPath] = append(grouped[file.RelPath], file)
	}

	return grouped
}

// GetUniqueRelPaths returns a sorted list of unique relative paths
func GetUniqueRelPaths(files []FileInfo) []string {
	seen := make(map[string]bool)
	var paths []string

	for _, file := range files {
		if !seen[file.RelPath] {
			seen[file.RelPath] = true
			paths = append(paths, file.RelPath)
		}
	}

	return paths
}

// FilterFilesByProject filters files belonging to a specific project
func FilterFilesByProject(files []FileInfo, projectAlias string) []FileInfo {
	var filtered []FileInfo

	for _, file := range files {
		if file.Project == projectAlias {
			filtered = append(filtered, file)
		}
	}

	return filtered
}

// NormalizePath normalizes a path for consistent comparison
func NormalizePath(path string) string {
	// Convert to forward slashes
	normalized := filepath.ToSlash(path)
	// Remove trailing slashes
	normalized = strings.TrimSuffix(normalized, "/")
	return normalized
}
