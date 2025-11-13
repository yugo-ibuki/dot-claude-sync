# claude-sync

A CLI tool to synchronize `.claude` directories across multiple projects.
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

### 1. Create Configuration File

Create `.claude-sync.yaml` at your project root:

```yaml
groups:
  frontend:
    paths:
      web: ./packages/web/.claude
      mobile: ./packages/mobile/.claude
      admin: ./packages/admin/.claude
    priority:
      - web      # highest priority
      - mobile   # second priority
      # admin has lowest priority
```

### 2. Run Sync

```bash
# Sync all projects in the frontend group
claude-sync push frontend
```

This distributes `.claude` directory contents across all projects based on priority settings.

## Features

### 📤 push - Synchronize Files

Collects `.claude` directory files from all projects in a group and distributes them based on priority.

```bash
claude-sync push <group>
```

**How it works:**
1. Collects files from all projects in the group
2. For duplicate filenames, uses the file from the highest priority project
3. Distributes collected files to all projects in the group

### 🗑️ rm - Delete Files

Deletes specified files or directories from all projects in a group.

```bash
claude-sync rm <group> <path>

# Examples
claude-sync rm frontend prompts/old-prompt.md
claude-sync rm backend prompts/deprecated/  # Delete entire directory
```

### 📝 mv - Move/Rename Files

Moves or renames files or directories across all projects in a group.

```bash
claude-sync mv <group> <from> <to>

# Examples
claude-sync mv frontend prompts/old.md prompts/new.md
claude-sync mv backend old-dir/ new-dir/
```

### 📋 list - List Groups

Shows list of configured groups or details of a specific group.

```bash
# Show all groups
claude-sync list

# Show details of a specific group
claude-sync list frontend
```

## Configuration File

### File Location

`.claude-sync.yaml` is searched in the following order:

1. Current directory
2. Parent directories (traversing upwards)
3. `~/.config/claude-sync/config.yaml` (global config)

### Configuration Examples

#### Basic Form (with aliases)

```yaml
groups:
  frontend:
    paths:
      web: ./packages/web/.claude
      mobile: ./packages/mobile/.claude
      admin: ./packages/admin/.claude
    priority:
      - web      # highest priority
      - mobile   # second priority
      # admin has lowest priority (not specified in priority)

  backend:
    paths:
      api: ./services/api/.claude
      worker: ./services/worker/.claude
    priority:
      - api
```

#### Simple Form (without aliases)

```yaml
groups:
  frontend:
    paths:
      - ./packages/web/.claude
      - ./packages/mobile/.claude
      - ./packages/admin/.claude
    priority:
      - ./packages/web/.claude
      - ./packages/mobile/.claude
```

#### Without Priority (paths order becomes default priority)

```yaml
groups:
  infra:
    paths:
      terraform: ./terraform/.claude
      k8s: ./k8s/.claude
    # Without priority specification, paths order becomes priority
    # 1. terraform (highest)
    # 2. k8s (second)
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
claude-sync push frontend --dry-run
claude-sync rm backend old.md --force
claude-sync push frontend --config ./custom-config.yaml
```

## Use Cases

### Case 1: Distribute New Prompt Across Frontend Projects

```bash
# Create new prompt in web project
cd packages/web/.claude/prompts
vim new-feature.md

# Distribute to entire frontend group
claude-sync push frontend
```

### Case 2: Delete Old Prompts from All Projects

```bash
# Verify before deletion
claude-sync rm backend prompts/deprecated/ --dry-run

# Execute deletion
claude-sync rm backend prompts/deprecated/
```

### Case 3: Standardize File Names

```bash
# Bulk rename across all projects
claude-sync mv frontend old-name.md new-name.md
```

### Case 4: Distribute Configuration from Master Project

```yaml
# Set web as master
frontend:
  paths:
    web: ./packages/web/.claude
    mobile: ./packages/mobile/.claude
    admin: ./packages/admin/.claude
  priority:
    - web  # web settings take priority
```

```bash
# All projects will be unified with web's configuration
claude-sync push frontend
```

## Important Notes

1. **Backup Recommended**: Backup `.claude` directories before first execution
2. **Understanding Conflicts**: Duplicate filenames are overwritten with content from higher priority projects
3. **Deletion is Irreversible**: `rm` command cannot be undone; use `--dry-run` for verification first
4. **Git Management**:
   - Configuration file (`.claude-sync.yaml`) should be under Git control
   - `.claude` directory itself should also be managed by Git as needed

## License

MIT

## Detailed Specification

For detailed operational specifications, see [spec/doc.md](./spec/doc.md).
