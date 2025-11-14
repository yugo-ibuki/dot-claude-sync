package utils

import (
	"os"
	"path/filepath"
	"testing"
)

// TestFileExists tests the FileExists function
func TestFileExists(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "existing file",
			path:     testFile,
			expected: true,
		},
		{
			name:     "existing directory",
			path:     tmpDir,
			expected: true,
		},
		{
			name:     "non-existent file",
			path:     filepath.Join(tmpDir, "nonexistent.txt"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FileExists(tt.path)
			if result != tt.expected {
				t.Errorf("FileExists(%s) = %v, expected %v", tt.path, result, tt.expected)
			}
		})
	}
}

// TestIsDirectory tests the IsDirectory function
func TestIsDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "directory",
			path:     tmpDir,
			expected: true,
		},
		{
			name:     "file",
			path:     testFile,
			expected: false,
		},
		{
			name:     "non-existent path",
			path:     filepath.Join(tmpDir, "nonexistent"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsDirectory(tt.path)
			if result != tt.expected {
				t.Errorf("IsDirectory(%s) = %v, expected %v", tt.path, result, tt.expected)
			}
		})
	}
}

// TestCopyFile tests the CopyFile function
func TestCopyFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create source file with content
	srcFile := filepath.Join(tmpDir, "source.txt")
	content := []byte("Hello, World!\nThis is test content.")
	if err := os.WriteFile(srcFile, content, 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Test successful copy
	t.Run("successful copy", func(t *testing.T) {
		dstFile := filepath.Join(tmpDir, "destination.txt")

		err := CopyFile(srcFile, dstFile)
		if err != nil {
			t.Fatalf("CopyFile failed: %v", err)
		}

		// Verify destination file exists
		if !FileExists(dstFile) {
			t.Error("Destination file does not exist")
		}

		// Verify content is identical
		dstContent, err := os.ReadFile(dstFile)
		if err != nil {
			t.Fatalf("Failed to read destination file: %v", err)
		}

		if string(dstContent) != string(content) {
			t.Errorf("Content mismatch: got %q, expected %q", string(dstContent), string(content))
		}

		// Verify permissions are preserved
		srcInfo, _ := os.Stat(srcFile)
		dstInfo, _ := os.Stat(dstFile)
		if srcInfo.Mode() != dstInfo.Mode() {
			t.Errorf("Permissions not preserved: got %v, expected %v", dstInfo.Mode(), srcInfo.Mode())
		}
	})

	// Test copy to subdirectory (should create parent dirs)
	t.Run("copy to nested directory", func(t *testing.T) {
		dstFile := filepath.Join(tmpDir, "subdir", "nested", "file.txt")

		err := CopyFile(srcFile, dstFile)
		if err != nil {
			t.Fatalf("CopyFile to nested directory failed: %v", err)
		}

		if !FileExists(dstFile) {
			t.Error("Destination file in nested directory does not exist")
		}
	})

	// Test copy non-existent source
	t.Run("non-existent source", func(t *testing.T) {
		err := CopyFile(filepath.Join(tmpDir, "nonexistent.txt"), filepath.Join(tmpDir, "dest.txt"))
		if err == nil {
			t.Error("Expected error when copying non-existent file, got nil")
		}
	})
}

// TestCopyDir tests the CopyDir function
func TestCopyDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create source directory structure
	srcDir := filepath.Join(tmpDir, "source")
	if err := os.MkdirAll(filepath.Join(srcDir, "subdir"), 0755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}

	// Create test files
	files := map[string]string{
		"file1.txt":        "content 1",
		"file2.txt":        "content 2",
		"subdir/file3.txt": "content 3",
	}

	for path, content := range files {
		fullPath := filepath.Join(srcDir, path)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", path, err)
		}
	}

	// Test successful directory copy
	t.Run("successful copy", func(t *testing.T) {
		dstDir := filepath.Join(tmpDir, "destination")

		err := CopyDir(srcDir, dstDir)
		if err != nil {
			t.Fatalf("CopyDir failed: %v", err)
		}

		// Verify all files were copied
		for path, expectedContent := range files {
			dstPath := filepath.Join(dstDir, path)

			if !FileExists(dstPath) {
				t.Errorf("File %s does not exist in destination", path)
				continue
			}

			content, err := os.ReadFile(dstPath)
			if err != nil {
				t.Errorf("Failed to read %s: %v", path, err)
				continue
			}

			if string(content) != expectedContent {
				t.Errorf("Content mismatch for %s: got %q, expected %q", path, string(content), expectedContent)
			}
		}
	})

	// Test copy non-directory
	t.Run("source is file not directory", func(t *testing.T) {
		srcFile := filepath.Join(srcDir, "file1.txt")
		dstDir := filepath.Join(tmpDir, "dest2")

		err := CopyDir(srcFile, dstDir)
		if err == nil {
			t.Error("Expected error when copying file as directory, got nil")
		}
	})
}

// TestRemoveFile tests the RemoveFile function
func TestRemoveFile(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("remove file", func(t *testing.T) {
		// Create test file
		testFile := filepath.Join(tmpDir, "remove_test.txt")
		if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Remove file
		err := RemoveFile(testFile)
		if err != nil {
			t.Fatalf("RemoveFile failed: %v", err)
		}

		// Verify file is removed
		if FileExists(testFile) {
			t.Error("File still exists after removal")
		}
	})

	t.Run("remove directory", func(t *testing.T) {
		// Create test directory with files
		testDir := filepath.Join(tmpDir, "remove_dir")
		if err := os.MkdirAll(filepath.Join(testDir, "subdir"), 0755); err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}
		if err := os.WriteFile(filepath.Join(testDir, "file.txt"), []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Remove directory
		err := RemoveFile(testDir)
		if err != nil {
			t.Fatalf("RemoveFile failed: %v", err)
		}

		// Verify directory is removed
		if FileExists(testDir) {
			t.Error("Directory still exists after removal")
		}
	})

	t.Run("remove non-existent file", func(t *testing.T) {
		// Should not error when removing non-existent file
		err := RemoveFile(filepath.Join(tmpDir, "nonexistent.txt"))
		if err != nil {
			t.Errorf("RemoveFile failed for non-existent file: %v", err)
		}
	})
}

// TestMoveFile tests the MoveFile function
func TestMoveFile(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("move file", func(t *testing.T) {
		// Create source file
		srcFile := filepath.Join(tmpDir, "move_src.txt")
		content := []byte("move test content")
		if err := os.WriteFile(srcFile, content, 0644); err != nil {
			t.Fatalf("Failed to create source file: %v", err)
		}

		// Move file
		dstFile := filepath.Join(tmpDir, "move_dst.txt")
		err := MoveFile(srcFile, dstFile)
		if err != nil {
			t.Fatalf("MoveFile failed: %v", err)
		}

		// Verify source no longer exists
		if FileExists(srcFile) {
			t.Error("Source file still exists after move")
		}

		// Verify destination exists with correct content
		if !FileExists(dstFile) {
			t.Error("Destination file does not exist")
		}

		dstContent, err := os.ReadFile(dstFile)
		if err != nil {
			t.Fatalf("Failed to read destination file: %v", err)
		}

		if string(dstContent) != string(content) {
			t.Errorf("Content mismatch: got %q, expected %q", string(dstContent), string(content))
		}
	})

	t.Run("move to nested directory", func(t *testing.T) {
		// Create source file
		srcFile := filepath.Join(tmpDir, "move_src2.txt")
		if err := os.WriteFile(srcFile, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create source file: %v", err)
		}

		// Move to nested directory
		dstFile := filepath.Join(tmpDir, "nested", "dir", "file.txt")
		err := MoveFile(srcFile, dstFile)
		if err != nil {
			t.Fatalf("MoveFile to nested directory failed: %v", err)
		}

		if !FileExists(dstFile) {
			t.Error("Destination file does not exist in nested directory")
		}
	})
}

// TestEnsureDir tests the EnsureDir function
func TestEnsureDir(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("create single directory", func(t *testing.T) {
		testDir := filepath.Join(tmpDir, "test_dir")

		err := EnsureDir(testDir)
		if err != nil {
			t.Fatalf("EnsureDir failed: %v", err)
		}

		if !IsDirectory(testDir) {
			t.Error("Directory was not created")
		}
	})

	t.Run("create nested directories", func(t *testing.T) {
		testDir := filepath.Join(tmpDir, "nested", "deep", "directory")

		err := EnsureDir(testDir)
		if err != nil {
			t.Fatalf("EnsureDir failed: %v", err)
		}

		if !IsDirectory(testDir) {
			t.Error("Nested directory was not created")
		}
	})

	t.Run("directory already exists", func(t *testing.T) {
		testDir := filepath.Join(tmpDir, "existing")
		if err := os.MkdirAll(testDir, 0755); err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}

		// Should not error when directory already exists
		err := EnsureDir(testDir)
		if err != nil {
			t.Errorf("EnsureDir failed for existing directory: %v", err)
		}
	})
}

// TestFileHash tests the FileHash function
func TestFileHash(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("hash calculation", func(t *testing.T) {
		// Create test file
		testFile := filepath.Join(tmpDir, "hash_test.txt")
		content := []byte("test content for hashing")
		if err := os.WriteFile(testFile, content, 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Calculate hash
		hash, err := FileHash(testFile)
		if err != nil {
			t.Fatalf("FileHash failed: %v", err)
		}

		// Verify hash is not empty
		if hash == "" {
			t.Error("Hash is empty")
		}

		// Verify hash is consistent
		hash2, err := FileHash(testFile)
		if err != nil {
			t.Fatalf("Second FileHash failed: %v", err)
		}

		if hash != hash2 {
			t.Errorf("Hash inconsistent: got %s and %s", hash, hash2)
		}

		// Verify different content produces different hash
		testFile2 := filepath.Join(tmpDir, "hash_test2.txt")
		if err := os.WriteFile(testFile2, []byte("different content"), 0644); err != nil {
			t.Fatalf("Failed to create second test file: %v", err)
		}

		hash3, err := FileHash(testFile2)
		if err != nil {
			t.Fatalf("FileHash for second file failed: %v", err)
		}

		if hash == hash3 {
			t.Error("Different files produced same hash")
		}
	})

	t.Run("non-existent file", func(t *testing.T) {
		_, err := FileHash(filepath.Join(tmpDir, "nonexistent.txt"))
		if err == nil {
			t.Error("Expected error for non-existent file, got nil")
		}
	})
}

// TestFormatSize tests the FormatSize function
func TestFormatSize(t *testing.T) {
	tests := []struct {
		name     string
		size     int64
		expected string
	}{
		{
			name:     "bytes",
			size:     512,
			expected: "512 B",
		},
		{
			name:     "kilobytes",
			size:     1536, // 1.5 KB
			expected: "1.5 KB",
		},
		{
			name:     "megabytes",
			size:     1572864, // 1.5 MB
			expected: "1.5 MB",
		},
		{
			name:     "gigabytes",
			size:     1610612736, // 1.5 GB
			expected: "1.5 GB",
		},
		{
			name:     "zero",
			size:     0,
			expected: "0 B",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatSize(tt.size)
			if result != tt.expected {
				t.Errorf("FormatSize(%d) = %s, expected %s", tt.size, result, tt.expected)
			}
		})
	}
}
