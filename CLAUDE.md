# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

claude-sync is a CLI tool that synchronizes `.claude` directories across multiple independent projects in a workspace. It manages files in groups, performs bulk operations (add, overwrite, delete, move), and resolves conflicts based on configurable priority settings.

## Installation and Deployment

### For End Users

**Option 1: Install via `go install` (Recommended)**
```bash
go install github.com/yugo-ibuki/dot-claude-sync@latest
```
This installs the binary to `$GOPATH/bin` (usually `~/go/bin`).

**Option 2: Build from source**
```bash
git clone https://github.com/yugo-ibuki/dot-claude-sync.git
cd dot-claude-sync
go build
./dot-claude-sync
```

**Option 3: Download pre-built binaries**
Download from GitHub Releases (if available).

### For Development

```bash
# Build the binary
go build -o dot-claude-sync

# Install locally to $GOPATH/bin
go install

# Run directly without building
go run main.go <command>

# Run with specific command
go run main.go init
go run main.go push web-projects
go run main.go list

# Check dependencies
go mod tidy
go mod verify
```

**Important**: Build artifacts (`dot-claude-sync`, `claude-sync`) are excluded from git via `.gitignore`. Do not commit binaries to the repository.

## Architecture

### Command Structure (Cobra-based CLI)

The project uses `github.com/spf13/cobra` for CLI commands:

- **main.go**: Entry point that calls `cmd.Execute()`
- **cmd/root.go**: Root command with global flags (`--config`, `--dry-run`, `--verbose`, `--force`)
- **cmd/init.go**: Interactive configuration file creation
- **cmd/push.go**: File synchronization across projects (TODO: implementation pending)
- **cmd/rm.go**: Delete files from all projects in a group
- **cmd/mv.go**: Move/rename files across all projects
- **cmd/list.go**: Display groups and group details

### Configuration System

**config/config.go** handles YAML configuration loading and parsing:

- **Config struct**: Root configuration with `Groups` map
- **Group struct**: Flexible `Paths` (map or slice) and optional `Priority` list
- **ProjectPath struct**: Resolved paths with alias and priority ranking

**Configuration priority rules**:
1. If `priority` list is specified, use that order
2. If no `priority`, use `paths` order as default priority
3. Projects not in `priority` list get lowest priority (len(priority) + 1)

**Configuration file location** (fixed):
- Default: `~/.config/claude-sync/config.yaml`
- Override with `--config` flag

### Data Flow for `push` Command (To Be Implemented)

1. Load configuration from `~/.config/claude-sync/config.yaml`
2. Parse group and resolve project paths with priorities
3. **Collection phase**: Gather all files from `.claude` directories across projects
4. **Conflict resolution**: For duplicate filenames, select from highest priority project
5. **Distribution phase**: Copy resolved files to all projects in the group

### Core Packages

- **cmd/**: Cobra command definitions and execution logic
- **config/**: Configuration loading, parsing, and priority resolution
- **syncer/**: (Empty) Intended for file collection, conflict resolution, and sync logic
- **utils/**: (Empty) Intended for file operations and user prompts

## Implementation Status

**Completed**:
- CLI structure with Cobra commands
- Configuration file loading and parsing
- Priority resolution system
- Interactive `init` command for config creation

**TODO (marked in code)**:
- File collection logic in `push` command
- Conflict resolution implementation
- File synchronization implementation
- `rm` command file deletion logic
- `mv` command file moving logic
- Error handling for missing/invalid paths
- Unit tests

## Configuration Examples

See README.md for detailed examples. Key formats:

**With aliases**:
```yaml
groups:
  web-projects:
    paths:
      frontend: ~/projects/web-frontend/.claude
      backend: ~/projects/web-backend/.claude
    priority:
      - frontend
```

**Without aliases (simple list)**:
```yaml
groups:
  go-projects:
    paths:
      - ~/go/src/project-a/.claude
      - ~/go/src/project-b/.claude
```

## Key Implementation Notes

- All commands support `--dry-run` for safe testing
- `--force` skips confirmation prompts (for `rm`/`mv`)
- Configuration path resolution handles `~` expansion
- Priority system allows both explicit `priority` list and implicit `paths` order
- Interactive `init` command guides users through initial setup
- Version is hardcoded in `root.go` (0.1.0)
