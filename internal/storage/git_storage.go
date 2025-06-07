package storage

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/tylerkeyes/dot-sync/internal"
	"github.com/tylerkeyes/dot-sync/internal/shared"
)

type GitStorage struct {
	RemoteURL string
}

func (s *GitStorage) InitializeStorage() error {
	home := shared.FindHomeDir()
	dir := filepath.Join(home, shared.GetDotSyncFilesDir())
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
	db, err := internal.OpenDotSyncDB()
	if err != nil {
		return fmt.Errorf("failed to open db: %w", err)
	}
	defer db.Close()
	if err := internal.EnsureStorageTable(db); err != nil {
		return fmt.Errorf("failed to ensure storage_provider table: %w", err)
	}
	if err := internal.InsertStorageProvider(db, "git", s.RemoteURL); err != nil {
		return fmt.Errorf("failed to insert storage provider: %w", err)
	}
	return nil
}

func (s *GitStorage) PushToStorage(filePath string) error {
	return nil
}

func (s *GitStorage) PullFromStorage(filePath string) error {
	return nil
}

func isRemoteExistsError(err error) bool {
	// git returns exit code 128 and message contains "remote origin already exists"
	return err != nil && (err.Error() == "exit status 128" ||
		(err.Error() != "" && containsRemoteExistsMsg(err.Error())))
}

func containsRemoteExistsMsg(msg string) bool {
	return msg != "" && (msg == "fatal: remote origin already exists." ||
		msg == "fatal: remote origin already exists")
}
