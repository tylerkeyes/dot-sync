package shared

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Constants

func GetDotSyncDir() string {
	return ".dot-sync"
}

func GetDotSyncFilesDir() string {
	return ".dot-sync/files"
}

func GetGitRemoteURL() string {
	return "https://github.com/tylerkeyes/dot-sync-test.git"
}

type contextKey string

func GetStorageProviderKey() contextKey {
	return contextKey("storageProvider")
}

// Utils

func EnsureDir(path string) error {
	return os.MkdirAll(path, 0700)
}

func FindHomeDir() string {
	baseDir := ""
	if h := os.Getenv("HOME"); h != "" {
		baseDir = h
	} else if u, err := os.UserHomeDir(); err == nil {
		baseDir = u
	}
	return baseDir
}

// Path conversion utilities for cross-device compatibility

// ToStoragePath converts an absolute path to a storage path by replacing
// the home directory with "HOME" placeholder if the path is within home directory
func ToStoragePath(absolutePath string) string {
	homeDir := FindHomeDir()
	if homeDir == "" {
		return absolutePath
	}

	// Clean both paths to ensure consistent separators
	cleanHome := filepath.Clean(homeDir)
	cleanPath := filepath.Clean(absolutePath)

	// Check if the path is within the home directory
	if strings.HasPrefix(cleanPath, cleanHome) {
		// Get the relative path from home
		relPath, err := filepath.Rel(cleanHome, cleanPath)
		if err != nil {
			return absolutePath
		}

		// If it's exactly the home directory
		if relPath == "." {
			return "HOME"
		}

		// Return with HOME placeholder
		return filepath.Join("HOME", relPath)
	}

	// If not in home directory, return as-is
	return absolutePath
}

// FromStoragePath converts a storage path back to an absolute path by replacing
// "HOME" placeholder with the current machine's home directory
func FromStoragePath(storagePath string) string {
	if !strings.HasPrefix(storagePath, "HOME") {
		return storagePath
	}

	homeDir := FindHomeDir()
	if homeDir == "" {
		return storagePath
	}

	// If it's exactly "HOME"
	if storagePath == "HOME" {
		return homeDir
	}

	// If it starts with "HOME/"
	if strings.HasPrefix(storagePath, "HOME"+string(filepath.Separator)) {
		relPath := storagePath[len("HOME"+string(filepath.Separator)):]
		return filepath.Join(homeDir, relPath)
	}

	return storagePath
}

// File copy utilities

func CopyToDotSyncFilesByID(id int, filePath string, dotSyncFilesPath string) error {
	info, err := os.Lstat(filePath)
	if err != nil {
		return err
	}
	dst := filepath.Join(dotSyncFilesPath, fmt.Sprintf("%d", id))
	if info.IsDir() {
		return CopyDir(filePath, dst)
	}
	return CopyFile(filePath, dst)
}

func CopyFile(src, dst string) error {
	if err := EnsureDir(filepath.Dir(dst)); err != nil {
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

func CopyDir(src, dst string) error {
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
			return EnsureDir(dstPath)
		}
		return CopyFile(path, dstPath)
	})
}

func RunCmd(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
