package syncer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yugo-ibuki/dot-claude-sync/config"
)

func TestCollectFilesWithExclude(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "test-collect-exclude-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Setup test project with various files
	project1 := filepath.Join(tmpDir, "project1", ".claude")
	if err := os.MkdirAll(project1, 0755); err != nil {
		t.Fatal(err)
	}

	// Create test files
	testFiles := map[string]string{
		"prompts/auth.md":       "auth content",
		"prompts/api.md":        "api content",
		"prompts/test.bak":      "backup file",
		"temp/debug.log":        "debug log",
		"temp/cache.txt":        "cache file",
		".DS_Store":             "mac file",
		"commands/build.sh":     "build script",
		"config/settings.json":  "settings",
		"config/local.json.bak": "local backup",
	}

	for relPath, content := range testFiles {
		fullPath := filepath.Join(project1, relPath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name            string
		excludePatterns []string
		expectedCount   int
		shouldInclude   []string
		shouldExclude   []string
	}{
		{
			name:            "no exclude patterns",
			excludePatterns: nil,
			expectedCount:   9,
			shouldInclude:   []string{"prompts/auth.md", "temp/debug.log", ".DS_Store"},
			shouldExclude:   []string{},
		},
		{
			name:            "exclude .bak files",
			excludePatterns: []string{"*.bak"},
			expectedCount:   7,
			shouldInclude:   []string{"prompts/auth.md", "temp/debug.log"},
			shouldExclude:   []string{"prompts/test.bak", "config/local.json.bak"},
		},
		{
			name:            "exclude temp directory",
			excludePatterns: []string{"temp/*"},
			expectedCount:   7,
			shouldInclude:   []string{"prompts/auth.md", "commands/build.sh"},
			shouldExclude:   []string{"temp/debug.log", "temp/cache.txt"},
		},
		{
			name:            "exclude .DS_Store",
			excludePatterns: []string{".DS_Store"},
			expectedCount:   8,
			shouldInclude:   []string{"prompts/auth.md", "temp/debug.log"},
			shouldExclude:   []string{".DS_Store"},
		},
		{
			name:            "multiple exclude patterns",
			excludePatterns: []string{"*.bak", "temp/*", ".DS_Store"},
			expectedCount:   4,
			shouldInclude:   []string{"prompts/auth.md", "prompts/api.md", "commands/build.sh", "config/settings.json"},
			shouldExclude:   []string{"prompts/test.bak", "config/local.json.bak", "temp/debug.log", "temp/cache.txt", ".DS_Store"},
		},
		{
			name:            "exclude all in prompts",
			excludePatterns: []string{"prompts/*"},
			expectedCount:   6,
			shouldInclude:   []string{"temp/debug.log", "commands/build.sh"},
			shouldExclude:   []string{"prompts/auth.md", "prompts/api.md"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projects := []config.ProjectPath{
				{Alias: "project1", Path: project1, Priority: 1},
			}

			collected, err := CollectFiles(projects, tt.excludePatterns)
			if err != nil {
				t.Fatalf("CollectFiles failed: %v", err)
			}

			// Check count
			if len(collected) != tt.expectedCount {
				t.Errorf("Expected %d files, got %d", tt.expectedCount, len(collected))
			}

			// Check that expected files are included
			for _, expectedPath := range tt.shouldInclude {
				found := false
				for _, file := range collected {
					if file.RelPath == expectedPath {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected file %s to be included, but it was not", expectedPath)
				}
			}

			// Check that expected files are excluded
			for _, excludedPath := range tt.shouldExclude {
				for _, file := range collected {
					if file.RelPath == excludedPath {
						t.Errorf("Expected file %s to be excluded, but it was included", excludedPath)
					}
				}
			}
		})
	}
}

func TestShouldExclude(t *testing.T) {
	tests := []struct {
		name     string
		relPath  string
		patterns []string
		want     bool
	}{
		{
			name:     "no patterns",
			relPath:  "prompts/auth.md",
			patterns: nil,
			want:     false,
		},
		{
			name:     "match extension",
			relPath:  "prompts/test.bak",
			patterns: []string{"*.bak"},
			want:     true,
		},
		{
			name:     "no match extension",
			relPath:  "prompts/test.md",
			patterns: []string{"*.bak"},
			want:     false,
		},
		{
			name:     "match directory pattern",
			relPath:  "temp/debug.log",
			patterns: []string{"temp/*"},
			want:     true,
		},
		{
			name:     "match filename",
			relPath:  ".DS_Store",
			patterns: []string{".DS_Store"},
			want:     true,
		},
		{
			name:     "match nested file",
			relPath:  "prompts/nested/file.md",
			patterns: []string{"prompts/*"},
			want:     false, // filepath.Match doesn't match nested paths with single *
		},
		{
			name:     "match multiple patterns",
			relPath:  "temp/cache.txt",
			patterns: []string{"*.bak", "temp/*", ".DS_Store"},
			want:     true,
		},
		{
			name:     "no match multiple patterns",
			relPath:  "commands/build.sh",
			patterns: []string{"*.bak", "temp/*", ".DS_Store"},
			want:     false,
		},
		{
			name:     "match basename",
			relPath:  "config/local.json.bak",
			patterns: []string{"*.bak"},
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldExclude(tt.relPath, tt.patterns)
			if got != tt.want {
				t.Errorf("shouldExclude(%q, %v) = %v, want %v", tt.relPath, tt.patterns, got, tt.want)
			}
		})
	}
}
