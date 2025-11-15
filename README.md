# claude-sync

[![CI](https://github.com/yugo-ibuki/dot-claude-sync/actions/workflows/ci.yml/badge.svg)](https://github.com/yugo-ibuki/dot-claude-sync/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/yugo-ibuki/dot-claude-sync)](https://goreportcard.com/report/github.com/yugo-ibuki/dot-claude-sync)
[![codecov](https://codecov.io/gh/yugo-ibuki/dot-claude-sync/branch/main/graph/badge.svg)](https://codecov.io/gh/yugo-ibuki/dot-claude-sync)

A CLI tool to synchronize `.claude` directories across multiple independent projects in your workspace.
Manage files in groups and perform bulk operations like add, overwrite, delete, and move.

## Installation

```bash
go install github.com/yugo-ibuki/dot-claude-sync@latest
```

Or build from source:

```bash
git clone https://github.com/yugo-ibuki/dot-claude-sync.git
cd dot-claude-sync
go build -o claude-sync
```

## Quick Start

### 1. Initialize Configuration

Run the init command to create your configuration file:

```bash
claude-sync init
```

This will interactively guide you through creating `~/.config/claude-sync/config.yaml`.

Alternatively, you can create the file manually:

```bash
mkdir -p ~/.config/claude-sync
vim ~/.config/claude-sync/config.yaml
```

Example configuration:

```yaml
groups:
  web-projects:
    paths:
      main-site: ~/projects/website/.claude
      api-server: ~/projects/api/.claude
      admin-dashboard: ~/projects/admin/.claude
    priority:
      - main-site  # highest priority
      - api-server # second priority
      # admin-dashboard has lowest priority
```

### 2. Run Sync

```bash
# Sync all projects in the web-projects group
claude-sync push web-projects
```

This distributes `.claude` directory contents across all projects based on priority settings.

## Commands

| Command | Description |
|---------|-------------|
| `claude-sync init` | Initialize configuration file interactively |
| `claude-sync detect <worktree-root> --group <group>` | Auto-detect .claude directories from git worktrees |
| `claude-sync push <group>` | Sync files across all projects in a group |
| `claude-sync rm <group> <path>` | Delete files from all projects in a group |
| `claude-sync mv <group> <from> <to>` | Move/rename files in all projects |
| `claude-sync list [group]` | Show groups or group details |

## Features

### 🎬 init - Initialize Configuration

Creates the configuration file with interactive prompts.

```bash
# Interactive setup
claude-sync init

# Overwrite existing config without confirmation
claude-sync init --force

# Preview what will be created
claude-sync init --dry-run
```

### 🔍 detect - Auto-Detect Worktree Paths

Automatically detects `.claude` directories from git worktrees and adds them to a group.

```bash
# Detect .claude directories in all worktrees
claude-sync detect <worktree-root> --group <group-name>

# Examples
claude-sync detect ~/ghq/github.com/user --group go-projects
claude-sync detect . --group current-project

# Preview detection without adding
claude-sync detect ~/projects --group web-projects --dry-run

# Add without confirmation
claude-sync detect ~/workspace --group python-services --force
```

**How it works:**
1. Runs `git worktree list` in the specified directory
2. Scans each worktree for `.claude` directories
3. Adds detected paths to the specified group in your configuration
4. If the group doesn't exist, creates it automatically

**Use case:**
Perfect for projects using git worktrees where you have multiple branches checked out simultaneously. Instead of manually adding each worktree path, `detect` automatically finds all `.claude` directories.

### 📤 push - Synchronize Files

Collects `.claude` directory files from all projects in a group and distributes them based on priority.

```bash
claude-sync push <group>
```

**How it works:**
1. Collects files from all projects in the group (across your workspace)
2. For duplicate filenames, uses the file from the highest priority project
3. Distributes collected files to all projects in the group

### 🗑️ rm - Delete Files

Deletes specified files or directories from all projects in a group.

```bash
claude-sync rm <group> <path>

# Examples
claude-sync rm web-projects prompts/old-prompt.md
claude-sync rm python-services prompts/deprecated/  # Delete entire directory
```

### 📝 mv - Move/Rename Files

Moves or renames files or directories across all projects in a group.

```bash
claude-sync mv <group> <from> <to>

# Examples
claude-sync mv web-projects prompts/old.md prompts/new.md
claude-sync mv python-services old-dir/ new-dir/
```

### 📋 list - List Groups

Shows list of configured groups or details of a specific group.

```bash
# Show all groups
claude-sync list

# Show details of a specific group
claude-sync list web-projects
```


## Configuration File

### File Location

Configuration file is located at a **fixed location**:

```
~/.config/claude-sync/config.yaml
```

This allows you to run `claude-sync` from any directory and it will always use the same configuration.

You can override this with the `--config` flag to use a different configuration file.

### Configuration Examples

#### Basic Form (with aliases)

```yaml
groups:
  web-projects:
    paths:
      frontend: ~/projects/web-frontend/.claude
      backend: ~/projects/web-backend/.claude
      shared: ~/projects/shared-components/.claude
    priority:
      - shared    # highest priority - shared config across projects
      - frontend  # second priority
      # backend has lowest priority (not specified in priority)

  python-services:
    paths:
      api: ~/workspace/python-api/.claude
      worker: ~/workspace/python-worker/.claude
      batch: ~/workspace/python-batch/.claude
    priority:
      - api  # api project has master configuration
```

#### Simple Form (without aliases)

```yaml
groups:
  go-projects:
    paths:
      - ~/go/src/github.com/user/project-a/.claude
      - ~/go/src/github.com/user/project-b/.claude
      - ~/go/src/github.com/user/project-c/.claude
    priority:
      - ~/go/src/github.com/user/project-a/.claude
      - ~/go/src/github.com/user/project-b/.claude
```

#### Without Priority (paths order becomes default priority)

```yaml
groups:
  client-projects:
    paths:
      acme-corp: ~/clients/acme-corp/.claude
      beta-inc: ~/clients/beta-inc/.claude
      gamma-ltd: ~/clients/gamma-ltd/.claude
    # Without priority specification, paths order becomes priority
    # 1. acme-corp (highest)
    # 2. beta-inc (second)
    # 3. gamma-ltd (lowest)
```

### Priority Rules

**When priority is specified:**
- Order in `priority` list determines precedence
- Projects not in the list have lowest priority

**When priority is not specified:**
- Order in `paths` becomes the priority order

## Global Options

Options available for all commands:

```bash
--config <path>   # Explicitly specify configuration file path
--dry-run         # Simulate execution without making changes
--verbose         # Output detailed logs
--force           # Skip confirmation prompts (for rm, mv commands)
```

**Examples:**
```bash
claude-sync push web-projects --dry-run
claude-sync rm python-services old.md --force
claude-sync push client-projects --config ~/.config/claude-sync/custom-config.yaml
```

## Use Cases

### Case 1: Quick Setup with Git Worktrees

If you're using git worktrees, you can quickly set up a group by auto-detecting all `.claude` directories:

```bash
# Detect and add all worktree .claude directories
claude-sync detect ~/projects/my-app --group my-app-features

# Verify detected paths
claude-sync list my-app-features

# Start syncing
claude-sync push my-app-features
```

This is much faster than manually adding each worktree path to your configuration.

### Case 2: Distribute New Prompt Across Related Projects

```bash
# Create new prompt in your main project
cd ~/projects/web-frontend/.claude/prompts
vim new-feature.md

# Distribute to all projects in the group
claude-sync push web-projects
```

### Case 3: Delete Old Prompts from All Projects

```bash
# Verify before deletion
claude-sync rm python-services prompts/deprecated/ --dry-run

# Execute deletion
claude-sync rm python-services prompts/deprecated/
```

### Case 4: Standardize File Names Across Workspace

```bash
# Bulk rename across all projects in the group
claude-sync mv web-projects old-name.md new-name.md
```

### Case 5: Distribute Configuration from Master Project

```yaml
# Set a shared project as master for consistent configuration
web-projects:
  paths:
    shared: ~/projects/shared-config/.claude
    frontend: ~/projects/web-frontend/.claude
    backend: ~/projects/web-backend/.claude
    mobile: ~/projects/mobile-app/.claude
  priority:
    - shared  # shared configuration takes priority
```

```bash
# All projects will be unified with shared configuration
claude-sync push web-projects
```

### Case 6: Sync Client Project Templates

```bash
# Set up configuration for multiple client projects
# ~/.config/claude-sync/config.yaml
groups:
  client-templates:
    paths:
      template: ~/templates/client-template/.claude
      client-a: ~/clients/client-a/.claude
      client-b: ~/clients/client-b/.claude
      client-c: ~/clients/client-c/.claude
    priority:
      - template  # template is the source of truth

# Distribute template to all client projects
claude-sync push client-templates
```

## Important Notes

1. **Backup Recommended**: Backup `.claude` directories before first execution
2. **Understanding Conflicts**: Duplicate filenames are overwritten with content from higher priority projects
3. **Deletion is Irreversible**: `rm` command cannot be undone; use `--dry-run` for verification first
4. **Git Management**:
   - Configuration file should be under Git control if shared across team
   - `.claude` directory itself should also be managed by Git as needed

## Uninstall

To completely remove claude-sync:

```bash
# Remove the binary
rm $(which claude-sync)

# Remove the configuration directory
rm -rf ~/.config/claude-sync
```

## License

MIT
