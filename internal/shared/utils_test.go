package shared

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetDotSyncDir(t *testing.T) {
	if dir := GetDotSyncDir(); dir != ".dot-sync" {
		t.Errorf("expected .dot-sync, got %q", dir)
	}
}

func TestGetDotSyncFilesDir(t *testing.T) {
	if dir := GetDotSyncFilesDir(); dir != ".dot-sync/files" {
		t.Errorf("expected .dot-sync/files, got %q", dir)
	}
}

func TestGetGitRemoteURL(t *testing.T) {
	expected := "https://github.com/tylerkeyes/dot-sync-test.git"
	if url := GetGitRemoteURL(); url != expected {
		t.Errorf("expected %q, got %q", expected, url)
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
	dir := filepath.Join(temp, "foo/bar/baz")
	if err := EnsureDir(dir); err != nil {
		t.Fatalf("failed to create nested dir: %v", err)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Errorf("dir not created: %v", err)
	}
	// Existing dir should not error
	if err := EnsureDir(dir); err != nil {
		t.Errorf("ensureDir on existing dir errored: %v", err)
	}
	// Invalid path (simulate by using a file as parent)
	file := filepath.Join(temp, "file")
	os.WriteFile(file, []byte("x"), 0600)
	badDir := filepath.Join(file, "bad")
	if err := EnsureDir(badDir); err == nil {
		t.Error("expected error for ensureDir with file as parent, got nil")
	}
}

func TestCopyFile(t *testing.T) {
	temp := t.TempDir()
	src := filepath.Join(temp, "src.txt")
	dst := filepath.Join(temp, "dst.txt")
	os.WriteFile(src, []byte("hello world"), 0600)
	if err := CopyFile(src, dst); err != nil {
		t.Fatalf("CopyFile failed: %v", err)
	}
	data, err := os.ReadFile(dst)
	if err != nil || string(data) != "hello world" {
		t.Errorf("CopyFile did not copy content correctly: %v, %q", err, data)
	}
	// Overwrite
	os.WriteFile(src, []byte("new content"), 0600)
	if err := CopyFile(src, dst); err != nil {
		t.Errorf("CopyFile overwrite failed: %v", err)
	}
	// Non-existent source
	if err := CopyFile(filepath.Join(temp, "nope.txt"), dst); err == nil {
		t.Error("expected error for non-existent source, got nil")
	}
}

func TestCopyDir(t *testing.T) {
	temp := t.TempDir()
	src := filepath.Join(temp, "srcdir")
	dst := filepath.Join(temp, "dstdir")
	os.MkdirAll(filepath.Join(src, "sub"), 0700)
	os.WriteFile(filepath.Join(src, "a.txt"), []byte("A"), 0600)
	os.WriteFile(filepath.Join(src, "sub", "b.txt"), []byte("B"), 0600)
	if err := CopyDir(src, dst); err != nil {
		t.Fatalf("CopyDir failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "a.txt")); err != nil {
		t.Errorf("file not copied: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "sub", "b.txt")); err != nil {
		t.Errorf("subdir file not copied: %v", err)
	}
	// Error: non-existent source
	if err := CopyDir(filepath.Join(temp, "nope"), dst); err == nil {
		t.Error("expected error for non-existent source dir, got nil")
	}
	// .git dir is skipped
	os.MkdirAll(filepath.Join(src, ".git", "inner"), 0700)
	os.WriteFile(filepath.Join(src, ".git", "inner", "skip.txt"), []byte("X"), 0600)
	if err := CopyDir(src, dst); err != nil {
		t.Errorf("CopyDir with .git dir failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, ".git", "inner", "skip.txt")); err == nil {
		t.Error(".git dir should be skipped, but file was copied")
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
