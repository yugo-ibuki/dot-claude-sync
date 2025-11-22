[![CI](https://github.com/yugo-ibuki/dot-claude-sync/actions/workflows/ci.yml/badge.svg)](https://github.com/yugo-ibuki/dot-claude-sync/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/yugo-ibuki/dot-claude-sync)](https://goreportcard.com/report/github.com/yugo-ibuki/dot-claude-sync)
[![codecov](https://codecov.io/gh/yugo-ibuki/dot-claude-sync/branch/main/graph/badge.svg)](https://codecov.io/gh/yugo-ibuki/dot-claude-sync)

# dot-claude-sync

[![CI](https://github.com/yugo-ibuki/dot-claude-sync/actions/workflows/ci.yml/badge.svg)](https://github.com/yugo-ibuki/dot-claude-sync/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/yugo-ibuki/dot-claude-sync)](https://goreportcard.com/report/github.com/yugo-ibuki/dot-claude-sync)
[![codecov](https://codecov.io/gh/yugo-ibuki/dot-claude-sync/branch/main/graph/badge.svg)](https://codecov.io/gh/yugo-ibuki/dot-claude-sync)

A CLI tool to synchronize `.claude` directories across multiple independent projects in your workspace.

## Overview

When working with Claude Code using git worktrees, it's tedious to share `.claude` directory contents (prompts, commands, skills) across worktrees. `dot-claude-sync` solves this problem by providing simple synchronization across multiple projects.

## Concept

### Why claude-sync?

When working with Claude Code, it's common and efficient to record progress within the `.claude` directory as long-term context or as spec documentation for future reference. Storing these documents in `.claude` is ideal because they're typically git-ignored (preventing repository pollution) while remaining accessible to Claude.

However, modern CLI Agent workflows increasingly rely on **git worktrees** to manage multiple development branches simultaneously. This creates a challenge: `.claude` documents aren't tracked by git, making it tedious to share and synchronize them across worktrees. Additionally, useful commands and skills created in one worktree often need to be available in others.

**claude-sync solves this problem** by providing a simple tool to share and synchronize `.claude` directory contents across worktrees and independent projects within your workspace.

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
<<<<<<< HEAD
| `claude-sync init` | Initialize configuration file interactively |
| `claude-sync detect <worktree-root> --group <group>` | Auto-detect .claude directories from git worktrees |
| `claude-sync push <group>` | Sync files across all projects in a group |
| `claude-sync rm <group> <path>` | Delete files from all projects in a group |
| `claude-sync mv <group> <from> <to>` | Move/rename files in all projects |
| `claude-sync list [group]` | Show groups or group details |
| `claude-sync config <subcommand>` | Manage configuration (add/remove groups and projects) |
||||||| f462719
| `claude-sync init` | Initialize configuration file interactively |
| `claude-sync push <group>` | Sync files across all projects in a group |
| `claude-sync rm <group> <path>` | Delete files from all projects in a group |
| `claude-sync mv <group> <from> <to>` | Move/rename files in all projects |
| `claude-sync list [group]` | Show groups or group details |
=======
| `init` | Initialize configuration file interactively |
| `detect <dir> --group <name>` | Auto-detect .claude directories from git worktrees |
| `push <group>` | Sync files across all projects in a group |
| `rm <group> <path>` | Delete files from all projects in a group |
| `mv <group> <from> <to>` | Move/rename files in all projects |
| `list [group]` | Show groups or group details |
| `config <subcommand>` | Manage configuration (add/remove groups and projects) |
>>>>>>> main

**Note**: All commands can use `dcs` instead of `dot-claude-sync` (e.g., `dcs init`, `dcs push <group>`).

### Global Options

```bash
<<<<<<< HEAD
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

### ⚙️ config - Manage Configuration

Manage groups and projects in the configuration file directly from the command line.

#### Show Configuration

```bash
# Show all groups and their projects
claude-sync config show

# Show details of a specific group
claude-sync config show web-projects
```

#### Manage Groups

```bash
# Add a new group
claude-sync config add-group mobile-projects

# Remove a group (with confirmation)
claude-sync config remove-group old-group

# Remove a group without confirmation
claude-sync config remove-group old-group --force
```

#### Manage Projects

```bash
# Add a project to a group
claude-sync config add-project web-projects new-app ~/projects/new-app/.claude

# Remove a project from a group
claude-sync config remove-project web-projects old-app
```

#### Set Priority

```bash
# Set priority order for a group (first = highest priority)
claude-sync config set-priority web-projects shared frontend backend

# The command above sets:
# 1. shared (highest priority)
# 2. frontend (second priority)
# 3. backend (third priority)
```

**Examples:**

```bash
# Create a new group and add projects
claude-sync config add-group client-projects
claude-sync config add-project client-projects acme ~/clients/acme/.claude
claude-sync config add-project client-projects beta ~/clients/beta/.claude

# Set priority
claude-sync config set-priority client-projects acme beta

# Verify configuration
claude-sync config show client-projects

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
||||||| f462719
# Interactive setup
claude-sync init

# Overwrite existing config without confirmation
claude-sync init --force

# Preview what will be created
claude-sync init --dry-run
```

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
=======
--config <path>   # Specify configuration file path
--dry-run         # Simulate execution without changes
>>>>>>> main
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

<<<<<<< HEAD
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
||||||| f462719
## Use Cases

### Case 1: Distribute New Prompt Across Related Projects
=======
### Distribute Files
>>>>>>> main

```bash
# Create new prompt in main project
cd ~/projects/main/.claude/prompts
vim new-feature.md

# Distribute to all projects in group
dcs push web-projects
```

<<<<<<< HEAD
### Case 3: Delete Old Prompts from All Projects
||||||| f462719
### Case 2: Delete Old Prompts from All Projects
=======
### Delete Files
>>>>>>> main

```bash
# Verify before deletion
dcs rm web-projects prompts/old.md --dry-run

# Execute deletion
dcs rm web-projects prompts/old.md
```

<<<<<<< HEAD
### Case 4: Standardize File Names Across Workspace
||||||| f462719
### Case 3: Standardize File Names Across Workspace
=======
### Manage Configuration
>>>>>>> main

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

<<<<<<< HEAD
### Case 5: Distribute Configuration from Master Project
||||||| f462719
### Case 4: Distribute Configuration from Master Project
=======
## Priority Rules
>>>>>>> main

- Priority is determined by order in `priority` list
- Projects not in the list have lowest priority
- If `priority` is not specified, `paths` order becomes priority
- Duplicate files are overwritten with content from higher priority projects

## Configuration File Location

<<<<<<< HEAD
### Case 6: Sync Client Project Templates
||||||| f462719
### Case 5: Sync Client Project Templates
=======
Default: `~/.config/dot-claude-sync/config.yaml`
>>>>>>> main

Override with `--config` flag

### Case 6: Add New Project to Existing Group

```bash
# New project created, add it to existing group
claude-sync config add-project web-projects new-service ~/projects/new-service/.claude

# Set it as highest priority if needed
claude-sync config set-priority web-projects new-service frontend backend

# Sync configuration to the new project
claude-sync push web-projects
```

### Case 7: Quick Configuration Management

```bash
# View current configuration
claude-sync config show

# Add a temporary project group for experimentation
claude-sync config add-group experimental
claude-sync config add-project experimental test-app ~/workspace/test-app/.claude

# Try it out
claude-sync push experimental

# Remove when done
claude-sync config remove-group experimental --force
```

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
