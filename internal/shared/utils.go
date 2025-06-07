package shared

import "os"

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
