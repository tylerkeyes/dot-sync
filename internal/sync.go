package internal

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tylerkeyes/dot-sync/internal/shared"
	"github.com/tylerkeyes/dot-sync/internal/storage"
)

func NewSyncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Sync dotfiles to remote storage",
		Run:   syncHandler,
	}
}

func syncHandler(cmd *cobra.Command, args []string) {
	// ctx := cmd.Context() // Only use if you need the storage provider

	dotSyncFilesPath := filepath.Join(shared.FindHomeDir(), shared.GetDotSyncFilesDir())

	db, err := OpenDotSyncDB()
	if err != nil {
		fmt.Println("Failed to open .dot-sync.db:", err)
		return
	}
	defer db.Close()

	records, err := GetAllFilePaths(db)
	if err != nil {
		fmt.Println("Failed to read file paths from database:", err)
		return
	}

	for _, rec := range records {
		if err := copyToDotSyncFilesByID(rec, dotSyncFilesPath); err != nil {
			fmt.Printf("Failed to copy %s: %v\n", rec.Path, err)
		}
	}

	// if err := ensureGitRepo(dotSyncFilesPath); err != nil {
	// 	fmt.Println("Failed to ensure git repository:", err)
	// 	return
	// }

	// if err := ensureRemoteRepoExists(); err != nil {
	// 	fmt.Println("Failed to ensure remote repository exists:", err)
	// 	return
	// }

	sp := cmd.Context().Value(storageProviderKey).(storage.StorageProvider)

	if err := gitAddCommitPush(dotSyncFilesPath); err != nil {
		fmt.Println("Failed to add/commit/push files:", err)
		return
	}

	fmt.Println("Sync complete.")
}

func copyToDotSyncFilesByID(rec FileRecord, dotSyncFilesPath string) error {
	info, err := os.Lstat(rec.Path)
	if err != nil {
		return err
	}
	dst := filepath.Join(dotSyncFilesPath, fmt.Sprintf("%d", rec.ID))
	if info.IsDir() {
		return copyDir(rec.Path, dst)
	}
	return copyFile(rec.Path, dst)
}

func copyFile(src, dst string) error {
	if err := shared.EnsureDir(filepath.Dir(dst)); err != nil {
		return err
	}
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	_, err = io.Copy(dstFile, srcFile)
	return err
}

func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && d.Name() == ".git" {
			return fs.SkipDir
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, rel)
		if d.IsDir() {
			return shared.EnsureDir(dstPath)
		}
		return copyFile(path, dstPath)
	})
}

func ensureGitRepo(repoPath string) error {
	gitDir := filepath.Join(repoPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		if err := runCmd(repoPath, "git", "init"); err != nil {
			return err
		}
		if err := runCmd(repoPath, "git", "remote", "add", "origin", shared.GetGitRemoteURL()); err != nil && !strings.Contains(err.Error(), "remote origin already exists") {
			return err
		}
	}
	return nil
}

func gitAddCommitPush(repoPath string) error {
	if err := runCmd(repoPath, "git", "add", "."); err != nil {
		return err
	}
	_ = runCmd(repoPath, "git", "commit", "-m", "sync: update dotfiles")
	branch, err := getCurrentGitBranch(repoPath)
	if err != nil || branch == "" {
		branch = "main"
	}
	if err := runCmd(repoPath, "git", "push", "-u", "origin", branch); err != nil {
		return err
	}
	return nil
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

func runCmd(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func ensureRemoteRepoExists() error {
	remoteURL := shared.GetGitRemoteURL()
	ownerRepo := parseGitHubOwnerRepo(remoteURL)
	if ownerRepo == "" {
		return fmt.Errorf("could not parse owner/repo from remote URL: %s", remoteURL)
	}
	// Check if repo exists
	checkCmd := exec.Command("gh", "repo", "view", ownerRepo)
	if err := checkCmd.Run(); err == nil {
		// Repo exists
		return nil
	}
	fmt.Printf("Remote repository %s does not exist. Creating...\n", ownerRepo)
	createCmd := exec.Command("gh", "repo", "create", ownerRepo, "--public", "--confirm")
	createCmd.Stdout = os.Stdout
	createCmd.Stderr = os.Stderr
	if err := createCmd.Run(); err != nil {
		return fmt.Errorf("failed to create remote repo: %w", err)
	}
	return nil
}

func parseGitHubOwnerRepo(remoteURL string) string {
	// Handles git@github.com:owner/repo.git or https://github.com/owner/repo.git
	remoteURL = strings.TrimSuffix(remoteURL, ".git")
	if strings.HasPrefix(remoteURL, "git@github.com:") {
		return strings.TrimPrefix(remoteURL, "git@github.com:")
	}
	if strings.HasPrefix(remoteURL, "https://github.com/") {
		return strings.TrimPrefix(remoteURL, "https://github.com/")
	}
	return ""
}
