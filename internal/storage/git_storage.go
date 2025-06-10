package storage

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/tylerkeyes/dot-sync/internal/db"
	"github.com/tylerkeyes/dot-sync/internal/shared"
)

type GitStorage struct {
	RemoteURL string
}

func (s *GitStorage) InitializeStorage() error {
	home := shared.FindHomeDir()
	dir := filepath.Join(home, shared.GetDotSyncDir())

	// Ensure the directory exists before running git commands
	if err := shared.EnsureDir(dir); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Initialize git repo if not already
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to init git repo: %w", err)
	}
	// Add remote if not present
	cmd = exec.Command("git", "remote", "add", "origin", s.RemoteURL)
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		// If remote already exists, ignore
		if !isRemoteExistsError(err) {
			return fmt.Errorf("failed to add remote: %w", err)
		}
	}
	// Ensure storage_provider table and insert provider info
	database, err := db.OpenDotSyncDB()
	if err != nil {
		return fmt.Errorf("failed to open db: %w", err)
	}
	defer database.Close()

	if err := db.EnsureStorageTable(database); err != nil {
		return fmt.Errorf("failed to ensure storage_provider table: %w", err)
	}

	_, _, err = db.GetStorageProvider(database)
	if err != nil {
		if err := db.InsertStorageProvider(database, "git", s.RemoteURL); err != nil {
			return fmt.Errorf("failed to insert storage provider: %w", err)
		}
	} else {
		if err := db.UpdateStorageProvider(database, "git", s.RemoteURL); err != nil {
			return fmt.Errorf("failed to update storage provider: %w", err)
		}
	}

	return nil
}

func (s *GitStorage) PushToStorage(filePath string) error {
	fmt.Println("Pushing contents to storage...")

	if err := shared.RunCmd(filePath, "git", "add", "."); err != nil {
		return err
	}
	_ = shared.RunCmd(filePath, "git", "commit", "-m", "sync: update dotfiles")
	branch, err := getCurrentGitBranch(filePath)
	if err != nil || branch == "" {
		branch = "main"
	}
	if err := shared.RunCmd(filePath, "git", "push", "--force", "-u", "origin", branch); err != nil {
		return err
	}

	return nil
}

func (s *GitStorage) PullFromStorage(filePath string) error {
	fmt.Println("Pulling contents from storage...")

	if err := shared.RunCmd(filePath, "git", "fetch", "origin"); err != nil {
		return fmt.Errorf("failed to fetch from remote: %w", err)
	}

	branch, err := getCurrentGitBranch(filePath)
	if err != nil || branch == "" {
		branch = "main"
	}

	// Reset any local changes and pull from remote
	if err := shared.RunCmd(filePath, "git", "reset", "--hard", "origin/"+branch); err != nil {
		return fmt.Errorf("failed to reset to remote state: %w", err)
	}

	if err := shared.RunCmd(filePath, "git", "clean", "-fd"); err != nil {
		return fmt.Errorf("failed to clean untracked files: %w", err)
	}

	return nil
}

func isRemoteExistsError(err error) bool {
	// git returns exit code 3 or 128 and message contains "remote origin already exists"
	return err != nil && (err.Error() == "exit status 3" || err.Error() == "exit status 128" ||
		(err.Error() != "" && containsRemoteExistsMsg(err.Error())))
}

func containsRemoteExistsMsg(msg string) bool {
	return msg != "" && (msg == "fatal: remote origin already exists." ||
		msg == "fatal: remote origin already exists" ||
		msg == "error: remote origin already exists." ||
		msg == "error: remote origin already exists")
}

func getCurrentGitBranch(repoPath string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = repoPath
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
