# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## CRITICAL: Git Remote Operations Policy

**NEVER execute the following commands without explicit user instruction:**
- `git push`
- `git push --force`
- `git push --force-with-lease`
- `git tag` (with push)
- Any command that modifies remote repository state

**Always ask for confirmation before:**
- Pushing commits to remote
- Creating or pushing tags
- Any operation that affects the remote repository

**Allowed without confirmation:**
- Local commits (`git commit`)
- Staging files (`git add`)
- Creating local branches
- Local git operations that don't affect remote

**If asked "did you commit?" or similar:**
- Answer only about the local commit status
- Do NOT proceed to push without explicit instruction
- Ask if the user wants to push

## Project Overview

dot-claude-sync is a CLI tool that synchronizes `.claude` directories across multiple independent projects in a workspace (particularly useful for git worktrees). It manages files in groups, performs bulk operations (push, rm, mv, backup), and resolves conflicts based on configurable priority settings.

## Development Commands

### Building
```bash
# Build both binaries
go build                      # Creates dot-claude-sync
go build -o dcs ./cmd/dcs     # Creates dcs (short alias)

# Install to $GOPATH/bin
go install                    # Installs dot-claude-sync
go install ./cmd/dcs          # Installs dcs
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run specific package tests
go test ./config
go test ./syncer
go test ./cmd

# Run single test
go test -v -run TestResolveConflicts ./syncer

# Check test coverage
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Code Quality
```bash
# Run linter (requires golangci-lint)
golangci-lint run

# Format code
gofmt -w .
goimports -w .

# Tidy dependencies
go mod tidy
go mod verify
```

### Running Without Building
```bash
# Run commands directly
go run main.go init
go run main.go push web-projects
go run main.go list
go run main.go detect ~/projects/app --group my-group
```

**Important**: Build artifacts (`dot-claude-sync`, `dcs`) are in `.gitignore`. Never commit binaries.

## Version Management Workflow

The project uses an automated version update workflow triggered by git tags:

### Release Process

1. **Create and push a version tag**:
   ```bash
   git tag v0.2.0
   git push origin v0.2.0
   ```

2. **Automated workflow**:
   - GitHub Actions workflow (`.github/workflows/version-update.yml`) is triggered
   - Extracts version number from tag (e.g., `v0.2.0` → `0.2.0`)
   - Updates `cmd/root.go` version field
   - Builds and verifies the binary
   - Creates a PR with the version update

3. **Review and merge**:
   - Review the automated PR
   - Merge the PR to update the version in main branch

4. **Create GitHub Release**:
   - Create a GitHub Release from the tag
   - Draft or publish the release with release notes

### Manual Version Update

If you need to update the version manually:
```bash
# Edit cmd/root.go
vim cmd/root.go
# Change: Version: "0.1.4" → Version: "0.2.0"

# Verify
go build -o dot-claude-sync
./dot-claude-sync --version
```

## Architecture Overview

### Three-Phase Sync Pipeline

The core synchronization follows a collect → resolve → distribute pattern:

1. **Collection Phase** (`syncer/collector.go`):
   - `CollectFiles()`: Walks `.claude` directories in all projects
   - `GroupFilesByRelPath()`: Groups files by relative path (e.g., `prompts/auth.md`)
   - Returns `[]FileInfo` with source project metadata

2. **Conflict Resolution Phase** (`syncer/resolver.go`):
   - `ResolveConflicts()`: Handles duplicate filenames across projects
   - Priority-based selection: highest priority project wins
   - Returns `[]ResolvedFile` with winning source and conflict metadata

3. **Distribution Phase** (`syncer/syncer.go`):
   - `SyncFiles()`: Copies resolved files to all projects in group
   - `syncToProject()`: Handles per-project file operations
   - Skips overwriting identical files (hash comparison)
   - Returns `SyncResult` with operation summaries

### Configuration System

**Priority Resolution** (`config/config.go`):
- `Group.GetProjectPaths()` resolves both map and slice path formats
- Priority order: explicit `priority` list > `paths` order > unranked (lowest)
- `ProjectPath` struct contains `Alias`, `Path`, and `Priority` ranking

**Path Handling**:
- `~` expansion in paths (e.g., `~/projects/.claude`)
- Absolute path resolution via `filepath.Abs()`
- Validation via `utils.ValidateAndNormalizePath()`

### Command Structure

**Cobra Commands** (`cmd/`):
- `root.go`: Global flags (--config, --dry-run, --verbose, --force)
- `init.go`: Interactive config creation with prompts
- `push.go`: Three-phase sync execution
- `detect.go`: Git worktree auto-detection via `git worktree list --porcelain`
- `config.go`: Subcommands (add-group, remove-group, add-project, remove-project, set-priority)
- `backup.go`: Timestamped backups to `~/.local/share/dot-claude-sync/backups/`
- `rm.go`, `mv.go`, `list.go`: Bulk file operations

**Entry Points**:
- `main.go`: Calls `cmd.Execute()` for `dot-claude-sync` binary
- `cmd/dcs/main.go`: Same logic for `dcs` alias

### Key Data Structures

```go
// syncer/collector.go
type FileInfo struct {
    RelPath    string   // Relative path within .claude dir
    FullPath   string   // Absolute path on filesystem
    Hash       string   // SHA256 file hash
    ProjectIdx int      // Index in project list
    Alias      string   // Project alias from config
}

// syncer/resolver.go
type ResolvedFile struct {
    RelPath       string
    SourcePath    string
    SourceProject string
    SourcePriority int
    Hash          string
}

type Conflict struct {
    RelPath  string
    Projects []ConflictEntry  // All projects with this file
    Winner   ConflictEntry    // Chosen by priority
}

// syncer/syncer.go
type SyncResult struct {
    ProjectAlias string
    Copied       []string  // Successfully copied files
    Skipped      []string  // Identical files (no copy needed)
    Failed       []string  // Copy failures
    Overwrites   []OverwriteInfo
    Error        error
}
```

## Configuration File

**Location**: `~/.config/dot-claude-sync/config.yaml` (override with `--config`)

**Format Options**:

```yaml
# Option 1: Map with aliases and explicit priority
groups:
  web-projects:
    paths:
      main: ~/projects/main/.claude
      feature-a: ~/projects/feature-a/.claude
      feature-b: ~/projects/feature-b/.claude
    priority:
      - main        # Highest priority (rank 1)
      - feature-a   # Rank 2
      # feature-b gets rank 3 (not in priority list)

# Option 2: Simple list (order = priority)
groups:
  go-projects:
    paths:
      - ~/go/src/project-a/.claude  # Priority 1
      - ~/go/src/project-b/.claude  # Priority 2
```

## Key Implementation Patterns

### Table-Driven Tests
Tests use `[]struct` pattern with subtests:
```go
tests := []struct {
    name    string
    setup   func(t *testing.T) string
    want    expectedResult
}{
    {name: "test case 1", setup: ..., want: ...},
}
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        // Test implementation
    })
}
```

### Error Handling
- Early returns with `fmt.Errorf()` wrapping
- User-friendly error messages with context
- `--verbose` flag for detailed logging
- `--dry-run` for safe preview mode

### File Operations
**Atomic Operations** (`utils/file.go`):
- `CopyFile()`: Creates parent dirs, preserves permissions
- `FileHash()`: SHA256 hashing for duplicate detection
- `EnsureDir()`: Recursive directory creation
- `ValidateAndNormalizePath()`: Path validation with ~ expansion

### Git Worktree Integration
`detect` command parses `git worktree list --porcelain`:
```
worktree /path/to/worktree
HEAD <commit>
branch refs/heads/branch-name

worktree /path/to/another
...
```
Extracts paths, checks for `.claude` directories, adds to group config.
