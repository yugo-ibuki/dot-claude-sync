package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// CopyFile copies a file from src to dst
func CopyFile(src, dst string) error {
	// Expand home directory
	src = expandPath(src)
	dst = expandPath(dst)

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

// RemoveFile removes a file
func RemoveFile(path string) error {
	path = expandPath(path)

	if !FileExists(path) {
		return nil // Already doesn't exist
	}

	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to remove file: %w", err)
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
