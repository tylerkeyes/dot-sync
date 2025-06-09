# dot-sync

## Synchronize your configuration files across machines

The goal of this tool is to make it simple to keep my development environment consistent across machines.
I often will update a configuration on one machine, then wish I had it an another.
This tool will allow marking certain files or directories, then pushing or pulling changes from a remote.
This remote will likely be a github repo / gist, or other centralized location.
There would still be the issue of forgetting to sync my changes, but having a single command should aid my laziness.

## Usage

### Initial Setup

Before using dot-sync, you need to initialize a storage provider (currently supports Git):

```bash
# Initialize Git storage with an existing remote repository
dot-sync storage init --provider git --remote-url https://github.com/your-username/dotfiles.git
```

### Core Workflow

The typical workflow involves marking files for tracking, syncing them to remote storage, and pulling them on other machines:

**1. Mark files and directories for syncing:**
```bash
# Mark individual files
dot-sync mark ~/.vimrc ~/.bashrc

# Mark directories
dot-sync mark ~/.config/nvim ~/.ssh

# Mark files using relative paths (from current directory)
dot-sync mark .vimrc config/
```

**2. Sync your marked files to remote storage:**
```bash
# Push all marked files to your configured remote
dot-sync sync
```

**3. Pull files on another machine:**
```bash
# Pull and restore all synced files to their original locations
dot-sync pull
```

### Management Commands

**View currently tracked files:**
```bash
# List all files currently marked for syncing
dot-sync show
```

**Remove files from tracking:**
```bash
# Remove specific files from sync tracking and remote storage
dot-sync delete ~/.old-config ~/.unused-script

# Remove directories from tracking
dot-sync delete ~/.config/old-app
```

### Common Use Cases

- **New Machine Setup**: Initialize storage, then `dot-sync pull` to restore your entire development environment
- **Daily Sync**: After updating configs, run `dot-sync sync` to backup changes
- **Selective Syncing**: Use `dot-sync mark` to add new configuration files as your setup evolves
- **Cleanup**: Use `dot-sync delete` to remove outdated configurations from sync tracking

### File Management

The tool automatically handles:
- Both files and directories
- Absolute and relative paths
- Directory structure preservation
- Conflict resolution during pulls

All tracked files are stored locally in `~/.dot-sync/files/` and synchronized with your configured remote storage.

