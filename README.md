# dot-claude-sync

[![CI](https://github.com/yugo-ibuki/dot-claude-sync/actions/workflows/ci.yml/badge.svg)](https://github.com/yugo-ibuki/dot-claude-sync/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/yugo-ibuki/dot-claude-sync)](https://goreportcard.com/report/github.com/yugo-ibuki/dot-claude-sync)
[![codecov](https://codecov.io/gh/yugo-ibuki/dot-claude-sync/branch/main/graph/badge.svg)](https://codecov.io/gh/yugo-ibuki/dot-claude-sync)

A CLI tool to synchronize `.claude` directories across multiple independent projects in your workspace.
Manage files in groups and perform bulk operations like add, overwrite, delete, and move.

## Concept

### Why dot-claude-sync?

When working with Claude Code, it's common and efficient to record progress within the `.claude` directory as long-term context or as spec documentation for future reference. Storing these documents in `.claude` is ideal because they're typically git-ignored (preventing repository pollution) while remaining accessible to Claude.

However, modern CLI Agent workflows increasingly rely on **git worktrees** to manage multiple development branches simultaneously. This creates a challenge: `.claude` documents aren't tracked by git, making it tedious to share and synchronize them across worktrees. Additionally, useful commands and skills created in one worktree often need to be available in others.

**dot-claude-sync solves this problem** by providing a simple tool to share and synchronize `.claude` directory contents across worktrees and independent projects within your workspace.

## Installation

```bash
go install github.com/yugo-ibuki/dot-claude-sync@latest
```

Or build from source:

```bash
git clone https://github.com/yugo-ibuki/dot-claude-sync.git
cd dot-claude-sync
go build
```

## Quick Start

### 1. Initialize Configuration

Run the init command to create your configuration file:

```bash
dot-claude-sync init
```

This will interactively guide you through creating `~/.config/dot-claude-sync/config.yaml`.

Alternatively, you can create the file manually:

```bash
mkdir -p ~/.config/dot-claude-sync
vim ~/.config/dot-claude-sync/config.yaml
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
dot-claude-sync push web-projects
```

This distributes `.claude` directory contents across all projects based on priority settings.

## Commands

| Command | Description |
|---------|-------------|
| `dot-claude-sync init` | Initialize configuration file interactively |
| `dot-claude-sync detect <worktree-root> --group <group>` | Auto-detect .claude directories from git worktrees |
| `dot-claude-sync push <group>` | Sync files across all projects in a group |
| `dot-claude-sync rm <group> <path>` | Delete files from all projects in a group |
| `dot-claude-sync mv <group> <from> <to>` | Move/rename files in all projects |
| `dot-claude-sync list [group]` | Show groups or group details |
| `dot-claude-sync config <subcommand>` | Manage configuration (add/remove groups and projects) |

## Features

### 🎬 init - Initialize Configuration

Creates the configuration file with interactive prompts.

```bash
# Interactive setup
dot-claude-sync init

# Overwrite existing config without confirmation
dot-claude-sync init --force

# Preview what will be created
dot-claude-sync init --dry-run
```

### 🔍 detect - Auto-Detect Worktree Paths

Automatically detects `.claude` directories from git worktrees and adds them to a group.

```bash
# Detect .claude directories in all worktrees
dot-claude-sync detect <worktree-root> --group <group-name>

# Examples
dot-claude-sync detect ~/ghq/github.com/user --group go-projects
dot-claude-sync detect . --group current-project

# Preview detection without adding
dot-claude-sync detect ~/projects --group web-projects --dry-run

# Add without confirmation
dot-claude-sync detect ~/workspace --group python-services --force
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
dot-claude-sync push <group>
```

**How it works:**
1. Collects files from all projects in the group (across your workspace)
2. For duplicate filenames, uses the file from the highest priority project
3. Distributes collected files to all projects in the group

### 🗑️ rm - Delete Files

Deletes specified files or directories from all projects in a group.

```bash
dot-claude-sync rm <group> <path>

# Examples
dot-claude-sync rm web-projects prompts/old-prompt.md
dot-claude-sync rm python-services prompts/deprecated/  # Delete entire directory
```

### 📝 mv - Move/Rename Files

Moves or renames files or directories across all projects in a group.

```bash
dot-claude-sync mv <group> <from> <to>

# Examples
dot-claude-sync mv web-projects prompts/old.md prompts/new.md
dot-claude-sync mv python-services old-dir/ new-dir/
```

### 📋 list - List Groups

Shows list of configured groups or details of a specific group.

```bash
# Show all groups
dot-claude-sync list

# Show details of a specific group
dot-claude-sync list web-projects
```

### ⚙️ config - Manage Configuration

Manage groups and projects in the configuration file directly from the command line.

#### Show Configuration

```bash
# Show all groups and their projects
dot-claude-sync config show

# Show details of a specific group
dot-claude-sync config show web-projects
```

#### Manage Groups

```bash
# Add a new group
dot-claude-sync config add-group mobile-projects

# Remove a group (with confirmation)
dot-claude-sync config remove-group old-group

# Remove a group without confirmation
dot-claude-sync config remove-group old-group --force
```

#### Manage Projects

```bash
# Add a project to a group
dot-claude-sync config add-project web-projects new-app ~/projects/new-app/.claude

# Remove a project from a group
dot-claude-sync config remove-project web-projects old-app
```

#### Set Priority

```bash
# Set priority order for a group (first = highest priority)
dot-claude-sync config set-priority web-projects shared frontend backend

# The command above sets:
# 1. shared (highest priority)
# 2. frontend (second priority)
# 3. backend (third priority)
```

**Examples:**

```bash
# Create a new group and add projects
dot-claude-sync config add-group client-projects
dot-claude-sync config add-project client-projects acme ~/clients/acme/.claude
dot-claude-sync config add-project client-projects beta ~/clients/beta/.claude

# Set priority
dot-claude-sync config set-priority client-projects acme beta

# Verify configuration
dot-claude-sync config show client-projects

# Output:
# Group: client-projects
#
# Projects (2):
#   [1] acme
#       ~/clients/acme/.claude
#   [2] beta
#       ~/clients/beta/.claude
#
# Priority: [acme beta]
```


## Configuration File

### File Location

Configuration file is located at a **fixed location**:

```
~/.config/dot-claude-sync/config.yaml
```

This allows you to run `dot-claude-sync` from any directory and it will always use the same configuration.

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
dot-claude-sync push web-projects --dry-run
dot-claude-sync rm python-services old.md --force
dot-claude-sync push client-projects --config ~/.config/dot-claude-sync/custom-config.yaml
```

## Use Cases

### Case 1: Quick Setup with Git Worktrees

If you're using git worktrees, you can quickly set up a group by auto-detecting all `.claude` directories:

```bash
# Detect and add all worktree .claude directories
dot-claude-sync detect ~/projects/my-app --group my-app-features

# Verify detected paths
dot-claude-sync list my-app-features

# Start syncing
dot-claude-sync push my-app-features
```

This is much faster than manually adding each worktree path to your configuration.

### Case 2: Distribute New Prompt Across Related Projects

```bash
# Create new prompt in your main project
cd ~/projects/web-frontend/.claude/prompts
vim new-feature.md

# Distribute to all projects in the group
dot-claude-sync push web-projects
```

### Case 3: Delete Old Prompts from All Projects

```bash
# Verify before deletion
dot-claude-sync rm python-services prompts/deprecated/ --dry-run

# Execute deletion
dot-claude-sync rm python-services prompts/deprecated/
```

### Case 4: Standardize File Names Across Workspace

```bash
# Bulk rename across all projects in the group
dot-claude-sync mv web-projects old-name.md new-name.md
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
dot-claude-sync push web-projects
```

### Case 6: Sync Client Project Templates

```bash
# Set up configuration for multiple client projects
# ~/.config/dot-claude-sync/config.yaml
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
dot-claude-sync push client-templates
```

### Case 6: Add New Project to Existing Group

```bash
# New project created, add it to existing group
dot-claude-sync config add-project web-projects new-service ~/projects/new-service/.claude

# Set it as highest priority if needed
dot-claude-sync config set-priority web-projects new-service frontend backend

# Sync configuration to the new project
dot-claude-sync push web-projects
```

### Case 7: Quick Configuration Management

```bash
# View current configuration
dot-claude-sync config show

# Add a temporary project group for experimentation
dot-claude-sync config add-group experimental
dot-claude-sync config add-project experimental test-app ~/workspace/test-app/.claude

# Try it out
dot-claude-sync push experimental

# Remove when done
dot-claude-sync config remove-group experimental --force
```

## Important Notes

1. **Backup Recommended**: Backup `.claude` directories before first execution
2. **Understanding Conflicts**: Duplicate filenames are overwritten with content from higher priority projects
3. **Deletion is Irreversible**: `rm` command cannot be undone; use `--dry-run` for verification first
4. **Git Management**:
   - Configuration file should be under Git control if shared across team
   - `.claude` directory itself should also be managed by Git as needed

## Uninstall

To completely remove dot-claude-sync:

```bash
# Remove the binary
rm $(which dot-claude-sync)

# Remove the configuration directory
rm -rf ~/.config/dot-claude-sync
```

## License

MIT
