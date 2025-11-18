# dot-claude-sync

A CLI tool to synchronize `.claude` directories across multiple independent projects in your workspace.

## Overview

When working with Claude Code using git worktrees, it's tedious to share `.claude` directory contents (prompts, commands, skills) across worktrees. `dot-claude-sync` solves this problem by providing simple synchronization across multiple projects.

## Installation

```bash
# Install the main binary
go install github.com/yugo-ibuki/dot-claude-sync@latest

# Optionally install the shorter alias
go install github.com/yugo-ibuki/dot-claude-sync/cmd/dcs@latest
```

**Command Aliases**: Both `dot-claude-sync` and `dcs` work identically.

Or build from source:

```bash
git clone https://github.com/yugo-ibuki/dot-claude-sync.git
cd dot-claude-sync
go build                    # builds dot-claude-sync
go build -o dcs ./cmd/dcs   # builds dcs alias
```

## Quick Start

### 1. Initialize Configuration

```bash
dot-claude-sync init
# or using the short alias
dcs init
```

Or create manually:

```bash
mkdir -p ~/.config/dot-claude-sync
vim ~/.config/dot-claude-sync/config.yaml
```

### 2. Configuration Example

```yaml
groups:
  web-projects:
    paths:
      main: ~/projects/main/.claude
      feature-a: ~/projects/feature-a/.claude
      feature-b: ~/projects/feature-b/.claude
    priority:
      - main  # highest priority (used on conflicts)
```

### 3. Sync Files

```bash
dot-claude-sync push web-projects
# or
dcs push web-projects
```

## Commands

| Command | Description |
|---------|-------------|
| `init` | Initialize configuration file interactively |
| `detect <dir> --group <name>` | Auto-detect .claude directories from git worktrees |
| `push <group>` | Sync files across all projects in a group |
| `rm <group> <path>` | Delete files from all projects in a group |
| `mv <group> <from> <to>` | Move/rename files in all projects |
| `list [group]` | Show groups or group details |
| `config <subcommand>` | Manage configuration (add/remove groups and projects) |

**Note**: All commands can use `dcs` instead of `dot-claude-sync` (e.g., `dcs init`, `dcs push <group>`).

### Global Options

```bash
--config <path>   # Specify configuration file path
--dry-run         # Simulate execution without changes
--verbose         # Output detailed logs
--force           # Skip confirmation prompts
```

## Common Use Cases

### Auto-Detect Git Worktrees

```bash
# Auto-detect .claude directories from worktrees and add to group
dcs detect ~/projects/my-app --group my-app

# Verify detected paths
dcs list my-app

# Start syncing
dcs push my-app
```

### Distribute Files

```bash
# Create new prompt in main project
cd ~/projects/main/.claude/prompts
vim new-feature.md

# Distribute to all projects in group
dcs push web-projects
```

### Delete Files

```bash
# Verify before deletion
dcs rm web-projects prompts/old.md --dry-run

# Execute deletion
dcs rm web-projects prompts/old.md
```

### Manage Configuration

```bash
# Create new group
dcs config add-group mobile-projects

# Add projects
dcs config add-project mobile-projects ios ~/projects/ios-app/.claude
dcs config add-project mobile-projects android ~/projects/android-app/.claude

# Set priority
dcs config set-priority mobile-projects ios android

# Verify
dcs config show mobile-projects
```

## Priority Rules

- Priority is determined by order in `priority` list
- Projects not in the list have lowest priority
- If `priority` is not specified, `paths` order becomes priority
- Duplicate files are overwritten with content from higher priority projects

## Configuration File Location

Default: `~/.config/dot-claude-sync/config.yaml`

Override with `--config` flag

## Important Notes

- Backup `.claude` directories before first execution
- `rm` command is irreversible; use `--dry-run` for verification
- Files are overwritten based on priority when duplicates exist

## Uninstall

```bash
# Remove binaries
rm $(which dot-claude-sync)
rm $(which dcs)  # if dcs is installed

# Remove configuration directory
rm -rf ~/.config/dot-claude-sync
```

## License

MIT
