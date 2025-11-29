package utils

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// CopyFile copies a file from src to dst
func CopyFile(src, dst string) error {
	// Expand home directory
	src = expandPath(src)
	dst = expandPath(dst)

	// Check if source and destination are the same file
	srcAbs, err := filepath.Abs(src)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for source: %w", err)
	}
	dstAbs, err := filepath.Abs(dst)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for destination: %w", err)
	}

	if srcAbs == dstAbs {
		// Source and destination are the same file, skip copy
		return nil
	}

	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	// Ensure destination directory exists
	dstDir := filepath.Dir(dst)
	if err := EnsureDir(dstDir); err != nil {
		return err
	}

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	// Preserve file permissions
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source file: %w", err)
	}

	if err := os.Chmod(dst, sourceInfo.Mode()); err != nil {
		return fmt.Errorf("failed to set file permissions: %w", err)
	}

	return nil
}

// CopyDir recursively copies a directory from src to dst
func CopyDir(src, dst string) error {
	src = expandPath(src)
	dst = expandPath(dst)

	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source directory: %w", err)
	}

	if !srcInfo.IsDir() {
		return fmt.Errorf("source is not a directory: %s", src)
	}

	// Create destination directory
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := CopyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := CopyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// CopyDirExclude recursively copies a directory from src to dst, excluding specified directories
func CopyDirExclude(src, dst string, excludeDirs []string) error {
	src = expandPath(src)
	dst = expandPath(dst)

	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source directory: %w", err)
	}

	if !srcInfo.IsDir() {
		return fmt.Errorf("source is not a directory: %s", src)
	}

	// Create destination directory
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	// Create a map for quick lookup of excluded directories
	excludeMap := make(map[string]bool)
	for _, dir := range excludeDirs {
		excludeMap[dir] = true
	}

	for _, entry := range entries {
		// Skip excluded directories
		if entry.IsDir() && excludeMap[entry.Name()] {
			continue
		}

		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := CopyDirExclude(srcPath, dstPath, excludeDirs); err != nil {
				return err
			}
		} else {
			if err := CopyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// RemoveFile removes a file or directory (recursively if directory)
func RemoveFile(path string) error {
	path = expandPath(path)

	if !FileExists(path) {
		return nil // Already doesn't exist
	}

	// Use RemoveAll to handle both files and directories
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("failed to remove: %w", err)
	}

	return nil
}

// RemoveDir recursively removes a directory
func RemoveDir(path string) error {
	path = expandPath(path)

	if !FileExists(path) {
		return nil // Already doesn't exist
	}

	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("failed to remove directory: %w", err)
	}

	return nil
}

// MoveFile moves or renames a file from src to dst
func MoveFile(src, dst string) error {
	src = expandPath(src)
	dst = expandPath(dst)

	// Ensure destination directory exists
	dstDir := filepath.Dir(dst)
	if err := EnsureDir(dstDir); err != nil {
		return err
	}

	if err := os.Rename(src, dst); err != nil {
		return fmt.Errorf("failed to move file: %w", err)
	}

	return nil
}

// EnsureDir creates a directory and all parent directories if they don't exist
func EnsureDir(path string) error {
	path = expandPath(path)

	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return nil
}

// FileExists checks if a file or directory exists
func FileExists(path string) bool {
	path = expandPath(path)
	_, err := os.Stat(path)
	return err == nil
}

// FileHash calculates SHA256 hash of a file
func FileHash(path string) (string, error) {
	path = expandPath(path)

	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to calculate hash: %w", err)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// IsDirectory checks if the given path is a directory
func IsDirectory(path string) bool {
	path = expandPath(path)
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// FormatSize formats file size in human-readable format
func FormatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
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

	return filepath.Join(homeDir, path[1:])
}

// Confirm prompts the user for yes/no confirmation
func Confirm(message string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s (y/n): ", message)

	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

// ValidateAndNormalizePath validates and normalizes a path relative to .claude directory.
// It expects the path to start with ".claude/" and removes that prefix.
// Returns the normalized path without the ".claude/" prefix and an error if validation fails.
func ValidateAndNormalizePath(path string) (string, error) {
	// Trim whitespace
	path = strings.TrimSpace(path)

	if path == "" {
		return "", fmt.Errorf("path cannot be empty")
	}

	// Check for parent directory traversal before normalization
	if strings.Contains(path, "..") {
		return "", fmt.Errorf("path cannot contain '..' (parent directory references)")
	}

	// Check if path starts with .claude or .claude/
	if !strings.HasPrefix(path, ".claude") {
		return "", fmt.Errorf("path must start with '.claude/' (e.g., '.claude/commands/foo.md')")
	}

	// Remove .claude prefix
	// Handle both ".claude" and ".claude/"
	if path == ".claude" {
		return "", fmt.Errorf("path cannot be just '.claude', must specify a file or directory inside .claude")
	}

	// Remove ".claude/" prefix
	normalized := strings.TrimPrefix(path, ".claude/")
	normalized = strings.TrimPrefix(normalized, ".claude\\") // Handle Windows path separator

	if normalized == "" || normalized == path {
		return "", fmt.Errorf("path must specify a file or directory inside .claude (e.g., '.claude/commands/foo.md')")
	}

	// Normalize path separators
	normalized = filepath.Clean(normalized)

	// Additional validation: ensure no absolute path
	if filepath.IsAbs(normalized) {
		return "", fmt.Errorf("path cannot be absolute")
	}

	return normalized, nil
}

// DeleteEmptyFolders recursively deletes all empty directories in the given path
// Returns a list of deleted folder paths and any error encountered
func DeleteEmptyFolders(rootPath string) ([]string, error) {
	rootPath = expandPath(rootPath)

	if !FileExists(rootPath) {
		return []string{}, fmt.Errorf("path does not exist: %s", rootPath)
	}

	if !IsDirectory(rootPath) {
		return []string{}, fmt.Errorf("path is not a directory: %s", rootPath)
	}

	var deletedFolders []string

	// Walk the directory tree from bottom up to delete empty directories
	// We use a post-order traversal to delete children before parents
	err := deleteEmptyFoldersRecursive(rootPath, &deletedFolders)
	if err != nil {
		return deletedFolders, err
	}

	return deletedFolders, nil
}

// deleteEmptyFoldersRecursive is a helper function that recursively processes directories
// It uses post-order traversal (process children before parent) to properly delete empty dirs
func deleteEmptyFoldersRecursive(dirPath string, deletedFolders *[]string) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", dirPath, err)
	}

	// Process all subdirectories first
	for _, entry := range entries {
		if entry.IsDir() {
			subDirPath := filepath.Join(dirPath, entry.Name())
			if err := deleteEmptyFoldersRecursive(subDirPath, deletedFolders); err != nil {
				// Continue processing other directories even if one fails
				fmt.Fprintf(os.Stderr, "Warning: error processing %s: %v\n", subDirPath, err)
			}
		}
	}

	// After processing subdirectories, check if current directory is empty
	// Re-read the directory to get updated contents
	entries, err = os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", dirPath, err)
	}

	if len(entries) == 0 {
		// Directory is empty, delete it
		if err := os.RemoveAll(dirPath); err != nil {
			return fmt.Errorf("failed to delete empty directory %s: %w", dirPath, err)
		}
		*deletedFolders = append(*deletedFolders, dirPath)
	}

	return nil
}
