# EnvTUI - Terminal UI for .env File Management

A production-ready TUI application for managing `.env` files built with Go and Charm's Bubble Tea framework.

![EnvTUI Screenshot - Main View](Screenshot%202026-02-03%20003318.png)

![EnvTUI Screenshot - Multi-File Mode](Screenshot%202026-02-03%20011929.png)

## Features

- **Multi-file support** - Manage multiple .env files with visual tabs (.env, .env.local, .env.production)
- **Git integration** - Visual git status icons in file tabs (? untracked, M modified, S staged, ✓ clean)
- **File comparison** - Compare values across different env files (press `c`)
- **Undo/Redo** - Press `u` to undo, `r` to redo changes
- **Diff view** - View unsaved changes before saving (press `v`)
- **Backup management** - View, restore, and delete backups (press `b`)
- **Bulk operations** - Multi-select entries with spacebar, bulk delete with `D`
- **Sorting** - Cycle through sort modes: alphabetical, category, value length (press `s`)
- **Copy between files** - Copy entries from one file to another (press `y`)
- **Quick templates** - Insert common env patterns in add/edit mode (press `t`)
- **Full CRUD operations** - Add, edit, delete .env entries
- **Secret detection** - automatically masks sensitive values (PASSWORD, SECRET, TOKEN, KEY)
- **Input validation** - detects duplicates, suspicious values, and formatting issues
- **Category-based color coding** - Database (blue), AWS (orange), API (green)
- **Fuzzy search** - filter entries with `/`
- **Vim-style navigation** - j/k for up/down
- **Import/Export** - JSON and YAML format support
- **Shell integration** - Export as shell commands, completions, and aliases
- **Comment and blank line preservation**
- **Export keyword support**
- **Multiline value support** - handles quoted multiline values
- **Escape sequence handling** - \n, \t, \r, \\, etc.
- **Atomic file writes** - automatic backups before modifications

## Installation

```bash
go mod download
go build -o envtui cmd/envtui/main.go

# Optional: Install to PATH
sudo cp envtui /usr/local/bin/
```

## Usage

### Single File Mode

```bash
# Use default .env file
./envtui

# Specify custom file
./envtui --files ".env.production"
```

### Multi-File Mode (Recommended)

View and manage multiple .env files simultaneously with visual tabs:

```bash
# Compare development, staging, and production configs
./envtui --files ".env,.env.staging,.env.production"

# Or shorter notation
./envtui --files ".env,.env.local,.env.prod"

# Any env files work
./envtui --files "config.env,secrets.env,override.env"

# Using the helper script (edit it to customize your files)
./run_multifile.sh
```

**Requirements:** Files must exist before running the command.

## Keybindings

### Navigation
- `↑/k` - Move up
- `↓/j` - Move down
- `/` - Search entries
- `Esc` - Cancel search/edit

### File Operations
- `a` - Add new entry
- `e` - Edit selected entry  
- `d` - Delete selected entry
- `D` - Bulk delete selected entries (multi-select mode)
- `x` - Toggle secret visibility

### History & Comparison
- `u` - Undo last change
- `r` - Redo last undone change
- `v` - View diff (show unsaved changes)
- `c` - Toggle comparison mode (shows ⚠ next to differing values)

### Multi-File Mode (when using --files)
- `1-9` - Switch between files (tabs shown at top)
- `y` - Copy selected entry to another file

### Organization & Management
- `s` - Cycle sort modes: alphabetical → category → value length
- `Space` - Toggle selection for bulk operations
- `b` - Open backup manager (view/restore/delete backups)

### Templates (in Add/Edit mode)
- `t` - Show quick templates menu (DATABASE_URL, API_KEY, etc.)

### Application
- `q` or `Ctrl+C` - Quit

## Example Workflows

### Basic Multi-File Workflow

```bash
# Start with multiple env files
./envtui --files ".env,.env.staging,.env.production"

# You'll see:
# [1:.env ✓ (12)] [2:.env.staging M (10)] [3:.env.production ? (11)]
# 📁 .env (12 entries)
#
# Press 2 to switch to .env.staging
# Press 3 to switch to .env.production
# Press c to see which values differ between files
# Press v to see unsaved changes in current file
# Press u to undo last change, r to redo
```

### Bulk Delete Workflow

```bash
# Start envtui with your files
./envtui --files ".env,.env.local"

# Select multiple entries for deletion:
# 1. Navigate to first entry to delete
# 2. Press Space to select it (shows ✓ next to entry)
# 3. Navigate to next entry
# 4. Press Space to select it
# 5. Repeat for all entries you want to delete
# 6. Press D to delete all selected entries at once
```

### Sorting Workflow

```bash
# Start envtui
./envtui --files ".env"

# Cycle through sort modes:
# Press s once - entries sorted alphabetically (A-Z)
# Press s again - entries grouped by category (Database, AWS, API, etc.)
# Press s again - entries sorted by value length (shortest first)
# Press s again - back to alphabetical
```

### Copy Between Files Workflow

```bash
# Start with multiple env files
./envtui --files ".env,.env.staging"

# Copy an entry from .env to .env.staging:
# 1. Navigate to the entry you want to copy
# 2. Press y
# 3. Select destination file from the list
# 4. Entry is copied with same key and value
```

### Quick Templates Workflow

```bash
# Start envtui
./envtui

# Add a new entry with a template:
# 1. Press a to add new entry
# 2. Enter the key name
# 3. When cursor is in value field, press t
# 4. Select template (e.g., DATABASE_URL)
# 5. Template fills in: postgresql://user:pass@localhost:5432/db
# 6. Edit as needed and press Enter to save

# Available templates:
# - DATABASE_URL - PostgreSQL connection string
# - API_KEY - Generic API key placeholder
# - SECRET_KEY - Django-style secret key
# - JWT_SECRET - JWT signing secret
# - REDIS_URL - Redis connection URL
```

### Backup Management Workflow

```bash
# Start envtui
./envtui --files ".env"

# Manage backups:
# 1. Press b to open backup manager
# 2. View list of all backups with timestamps
# 3. Navigate to a backup
# 4. Press r to restore that backup
# 5. Press d to delete a specific backup
# 6. Press D to delete all backups
# 7. Press Esc to return to main view
```

### Git Integration Workflow

```bash
# Start envtui in a git repository
./envtui --files ".env,.env.local,.env.production"

# Git status icons in tabs show:
# ? - File is untracked (not in git)
# M - File has uncommitted modifications
# S - File has staged changes
# ✓ - File is clean (committed or not tracked)

# Use this to track which env files need attention
# before committing your changes
```

## Import/Export

### Export to JSON or YAML

```bash
# Export to JSON
./envtui --files ".env" --export "backup.json" --format json

# Export to YAML
./envtui --files ".env" --export "backup.yaml" --format yaml

# Export as shell commands (for sourcing)
./envtui --files ".env" --format shell
# Output: KEY=value format
```

### Import from JSON or YAML

```bash
# Import as new file
./envtui --import "backup.json"

# Import and merge with existing file
./envtui --import "backup.json" --merge

# Import and overwrite existing values
./envtui --import "backup.json" --merge --overwrite
```

## Shell Integration

### Export Environment Variables

Export env vars directly to your shell:

```bash
# Export all variables (for eval)
eval $(./envtui --files ".env" --format shell)

# With export keyword
eval $(./envtui --files ".env" --format shell --format export)
```

### Shell Completion

Generate completion scripts for your shell:

```bash
# Bash completion
./envtui --completion bash >> ~/.bashrc

# Zsh completion
./envtui --completion zsh >> ~/.zshrc

# Fish completion
./envtui --completion fish >> ~/.config/fish/config.fish
```

### Shell Aliases

Show shell integration instructions:

```bash
./envtui --install
```

This will output shell functions you can add to your `.bashrc`, `.zshrc`, or `.bash_profile`:

```bash
# Function to launch envtui with files
envtui() {
    local files=".env"
    if [ -n "$1" ]; then
        files="$1"
    fi
    /usr/local/bin/envtui --files "$files"
}

# Function to export env vars
envtui-export() {
    local file=".env"
    if [ -n "$1" ]; then
        file="$1"
    fi
    eval $(/usr/local/bin/envtui --files "$file" --format shell)
}

# Quick aliases
alias et='envtui'
alias et-export='envtui-export'
```

## Project Structure

```
cmd/envtui/          # Entry point
internal/
  app/               # Bubble Tea app with undo/redo
  model/             # Domain models and change tracking
  parser/            # .env lexer/parser
  storage/           # File I/O, backups, import/export, shell
  ui/
    styles/          # Lipgloss themes
    views/           # TUI views (list, edit, diff)
```

## Testing

```bash
# Run all tests
go test ./... -v

# Run parser tests
go test ./internal/parser -v

# Run with coverage
go test ./... -cover
```

## Current Status

**Production Ready ✅**

All features listed above are fully implemented and tested.

## Quick Reference

| Command | Description |
|---------|-------------|
| `./envtui` | Open default .env file |
| `./envtui --files ".env,.env.local"` | Open multiple files |
| `./envtui --export backup.json` | Export to JSON |
| `./envtui --import backup.json --merge` | Import and merge |
| `./envtui --format shell` | Export as shell commands |
| `./envtui --completion bash` | Generate bash completions |
| `./envtui --install` | Show shell integration |

| Key | Action |
|-----|--------|
| `a` | Add entry |
| `e` | Edit entry |
| `d` | Delete entry |
| `D` | Bulk delete selected entries |
| `Space` | Toggle selection (for bulk ops) |
| `u` | Undo |
| `r` | Redo |
| `v` | View diff |
| `c` | Compare files |
| `b` | Backup manager |
| `s` | Cycle sort modes |
| `y` | Copy to another file |
| `t` | Quick templates (in add/edit) |
| `x` | Toggle secrets |
| `/` | Search |
| `1-9` | Switch file |
| `q` | Quit |
