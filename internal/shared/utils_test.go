package shared

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestGetDotSyncDir(t *testing.T) {
	if GetDotSyncDir() != ".dot-sync" {
		t.Errorf("expected '.dot-sync', got %q", GetDotSyncDir())
	}
}

func TestGetDotSyncFilesDir(t *testing.T) {
	if GetDotSyncFilesDir() != ".dot-sync/files" {
		t.Errorf("expected '.dot-sync/files', got %q", GetDotSyncFilesDir())
	}
}

func TestGetGitRemoteURL(t *testing.T) {
	expected := "https://github.com/tylerkeyes/dot-sync-test.git"
	if GetGitRemoteURL() != expected {
		t.Errorf("expected %q, got %q", expected, GetGitRemoteURL())
	}
}

func TestGetStorageProviderKey(t *testing.T) {
	expected := contextKey("storageProvider")
	if key := GetStorageProviderKey(); key != expected {
		t.Errorf("expected %q, got %q", expected, key)
	}
}

func TestEnsureDir(t *testing.T) {
	temp := t.TempDir()
	testDir := filepath.Join(temp, "test", "nested")
	if err := EnsureDir(testDir); err != nil {
		t.Errorf("EnsureDir failed: %v", err)
	}
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		t.Error("directory was not created")
	}
}

func TestCopyFile(t *testing.T) {
	temp := t.TempDir()
	src := filepath.Join(temp, "src.txt")
	dst := filepath.Join(temp, "dst.txt")
	data := []byte("hello world")
	if err := os.WriteFile(src, data, 0600); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}
	if err := CopyFile(src, dst); err != nil {
		t.Errorf("CopyFile failed: %v", err)
	}
	if copied, err := os.ReadFile(dst); err != nil || !reflect.DeepEqual(copied, data) {
		t.Errorf("copied file content mismatch: %v, %q", err, copied)
	}
}

func TestCopyDir(t *testing.T) {
	temp := t.TempDir()
	src := filepath.Join(temp, "src")
	dst := filepath.Join(temp, "dst")
	os.MkdirAll(src, 0700)
	os.WriteFile(filepath.Join(src, "file.txt"), []byte("content"), 0600)
	os.MkdirAll(filepath.Join(src, "subdir"), 0700)
	os.WriteFile(filepath.Join(src, "subdir", "nested.txt"), []byte("nested"), 0600)
	if err := CopyDir(src, dst); err != nil {
		t.Errorf("CopyDir failed: %v", err)
	}
	if data, err := os.ReadFile(filepath.Join(dst, "file.txt")); err != nil || string(data) != "content" {
		t.Errorf("copied file content mismatch: %v, %q", err, data)
	}
	if data, err := os.ReadFile(filepath.Join(dst, "subdir", "nested.txt")); err != nil || string(data) != "nested" {
		t.Errorf("copied nested file content mismatch: %v, %q", err, data)
	}
}

func TestCopyToDotSyncFilesByID(t *testing.T) {
	temp := t.TempDir()
	dotSyncFiles := filepath.Join(temp, "dot-sync-files")
	os.MkdirAll(dotSyncFiles, 0700)
	// File
	file := filepath.Join(temp, "foo.txt")
	os.WriteFile(file, []byte("abc"), 0600)
	if err := CopyToDotSyncFilesByID(42, file, dotSyncFiles); err != nil {
		t.Errorf("CopyToDotSyncFilesByID (file) failed: %v", err)
	}
	copied := filepath.Join(dotSyncFiles, "42")
	if data, err := os.ReadFile(copied); err != nil || string(data) != "abc" {
		t.Errorf("copied file content mismatch: %v, %q", err, data)
	}
	// Dir
	dir := filepath.Join(temp, "adir")
	os.MkdirAll(dir, 0700)
	os.WriteFile(filepath.Join(dir, "bar.txt"), []byte("def"), 0600)
	if err := CopyToDotSyncFilesByID(99, dir, dotSyncFiles); err != nil {
		t.Errorf("CopyToDotSyncFilesByID (dir) failed: %v", err)
	}
	copiedDir := filepath.Join(dotSyncFiles, "99")
	if _, err := os.Stat(filepath.Join(copiedDir, "bar.txt")); err != nil {
		t.Errorf("copied dir file missing: %v", err)
	}
	// Non-existent path
	if err := CopyToDotSyncFilesByID(100, filepath.Join(temp, "nope.txt"), dotSyncFiles); err == nil {
		t.Error("expected error for non-existent path, got nil")
	}
}

func TestFindHomeDir(t *testing.T) {
	// Save original environment
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)

	// Test with HOME environment variable set
	testHome := "/tmp/testhome"
	os.Setenv("HOME", testHome)
	result := FindHomeDir()
	if result != testHome {
		t.Errorf("expected %q, got %q", testHome, result)
	}

	// Test with empty HOME (should fall back to os.UserHomeDir)
	os.Setenv("HOME", "")
	result = FindHomeDir()
	// May return empty string if os.UserHomeDir fails, which is acceptable
	// Just test that the function doesn't panic
	_ = result

	// Test with unset HOME
	os.Unsetenv("HOME")
	result = FindHomeDir()
	// May return empty string if os.UserHomeDir fails, which is acceptable
	// Just test that the function doesn't panic
	_ = result
}

func TestRunCmd(t *testing.T) {
	temp := t.TempDir()
	// Success
	if err := RunCmd(temp, "echo", "hello"); err != nil {
		t.Errorf("RunCmd echo failed: %v", err)
	}
	// Error
	if err := RunCmd(temp, "nonexistent-cmd-xyz"); err == nil {
		t.Error("expected error for nonexistent command, got nil")
	}
}

func TestToStoragePath(t *testing.T) {
	// Save original environment
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)

	// Set test home directory
	testHome := "/Users/testuser"
	os.Setenv("HOME", testHome)

	tests := []struct {
		name         string
		absolutePath string
		expected     string
	}{
		{
			name:         "path within home directory",
			absolutePath: "/Users/testuser/.bashrc",
			expected:     "HOME/.bashrc",
		},
		{
			name:         "nested path within home directory",
			absolutePath: "/Users/testuser/.config/git/config",
			expected:     "HOME/.config/git/config",
		},
		{
			name:         "exactly home directory",
			absolutePath: "/Users/testuser",
			expected:     "HOME",
		},
		{
			name:         "path outside home directory",
			absolutePath: "/etc/hosts",
			expected:     "/etc/hosts",
		},
		{
			name:         "root path",
			absolutePath: "/",
			expected:     "/",
		},
		{
			name:         "path with extra slashes",
			absolutePath: "/Users/testuser///.bashrc",
			expected:     "HOME/.bashrc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToStoragePath(tt.absolutePath)
			if result != tt.expected {
				t.Errorf("ToStoragePath(%q) = %q, expected %q", tt.absolutePath, result, tt.expected)
			}
		})
	}
}

func TestFromStoragePath(t *testing.T) {
	// Save original environment
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)

	// Set test home directory
	testHome := "/Users/testuser"
	os.Setenv("HOME", testHome)

	tests := []struct {
		name        string
		storagePath string
		expected    string
	}{
		{
			name:        "HOME placeholder with relative path",
			storagePath: "HOME/.bashrc",
			expected:    "/Users/testuser/.bashrc",
		},
		{
			name:        "nested HOME placeholder path",
			storagePath: "HOME/.config/git/config",
			expected:    "/Users/testuser/.config/git/config",
		},
		{
			name:        "exactly HOME placeholder",
			storagePath: "HOME",
			expected:    "/Users/testuser",
		},
		{
			name:        "absolute path without HOME",
			storagePath: "/etc/hosts",
			expected:    "/etc/hosts",
		},
		{
			name:        "relative path without HOME",
			storagePath: "some/relative/path",
			expected:    "some/relative/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FromStoragePath(tt.storagePath)
			if result != tt.expected {
				t.Errorf("FromStoragePath(%q) = %q, expected %q", tt.storagePath, result, tt.expected)
			}
		})
	}
}

func TestToStoragePathNoHome(t *testing.T) {
	// Save original environment
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)

	// Unset HOME to test behavior when home directory cannot be determined
	os.Unsetenv("HOME")

	absolutePath := "/some/absolute/path"
	result := ToStoragePath(absolutePath)
	if result != absolutePath {
		t.Errorf("ToStoragePath(%q) with no HOME = %q, expected %q", absolutePath, result, absolutePath)
	}
}

func TestFromStoragePathNoHome(t *testing.T) {
	// Save original environment
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)

	// Unset HOME to test behavior when home directory cannot be determined
	os.Unsetenv("HOME")

	storagePath := "HOME/.bashrc"
	result := FromStoragePath(storagePath)
	if result != storagePath {
		t.Errorf("FromStoragePath(%q) with no HOME = %q, expected %q", storagePath, result, storagePath)
	}
}

func TestPathConversionRoundTrip(t *testing.T) {
	// Save original environment
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)

	// Set test home directory
	testHome := "/Users/testuser"
	os.Setenv("HOME", testHome)

	testPaths := []string{
		"/Users/testuser/.bashrc",
		"/Users/testuser/.config/git/config",
		"/Users/testuser",
		"/etc/hosts",
		"/",
	}

	for _, originalPath := range testPaths {
		t.Run(originalPath, func(t *testing.T) {
			storagePath := ToStoragePath(originalPath)
			retrievedPath := FromStoragePath(storagePath)
			if retrievedPath != originalPath {
				t.Errorf("Round trip failed: %q -> %q -> %q", originalPath, storagePath, retrievedPath)
			}
		})
	}
}
