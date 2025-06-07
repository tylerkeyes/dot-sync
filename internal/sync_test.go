package internal

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCopyFile(t *testing.T) {
	temp := t.TempDir()
	src := filepath.Join(temp, "src.txt")
	dst := filepath.Join(temp, "dst.txt")
	os.WriteFile(src, []byte("hello world"), 0600)
	if err := copyFile(src, dst); err != nil {
		t.Fatalf("copyFile failed: %v", err)
	}
	data, err := os.ReadFile(dst)
	if err != nil || string(data) != "hello world" {
		t.Errorf("copyFile did not copy content correctly: %v, %q", err, data)
	}
	// Overwrite
	os.WriteFile(src, []byte("new content"), 0600)
	if err := copyFile(src, dst); err != nil {
		t.Errorf("copyFile overwrite failed: %v", err)
	}
	// Non-existent source
	if err := copyFile(filepath.Join(temp, "nope.txt"), dst); err == nil {
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
	if err := copyDir(src, dst); err != nil {
		t.Fatalf("copyDir failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "a.txt")); err != nil {
		t.Errorf("file not copied: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "sub", "b.txt")); err != nil {
		t.Errorf("subdir file not copied: %v", err)
	}
	// Error: non-existent source
	if err := copyDir(filepath.Join(temp, "nope"), dst); err == nil {
		t.Error("expected error for non-existent source dir, got nil")
	}
	// .git dir is skipped
	os.MkdirAll(filepath.Join(src, ".git", "inner"), 0700)
	os.WriteFile(filepath.Join(src, ".git", "inner", "skip.txt"), []byte("X"), 0600)
	if err := copyDir(src, dst); err != nil {
		t.Errorf("copyDir with .git dir failed: %v", err)
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
	rec := FileRecord{ID: 42, Path: file}
	if err := copyToDotSyncFilesByID(rec, dotSyncFiles); err != nil {
		t.Errorf("copyToDotSyncFilesByID (file) failed: %v", err)
	}
	copied := filepath.Join(dotSyncFiles, fmt.Sprintf("%d", rec.ID))
	if data, err := os.ReadFile(copied); err != nil || string(data) != "abc" {
		t.Errorf("copied file content mismatch: %v, %q", err, data)
	}
	// Dir
	dir := filepath.Join(temp, "adir")
	os.MkdirAll(dir, 0700)
	os.WriteFile(filepath.Join(dir, "bar.txt"), []byte("def"), 0600)
	dirRec := FileRecord{ID: 99, Path: dir}
	if err := copyToDotSyncFilesByID(dirRec, dotSyncFiles); err != nil {
		t.Errorf("copyToDotSyncFilesByID (dir) failed: %v", err)
	}
	copiedDir := filepath.Join(dotSyncFiles, fmt.Sprintf("%d", dirRec.ID))
	if _, err := os.Stat(filepath.Join(copiedDir, "bar.txt")); err != nil {
		t.Errorf("copied dir file missing: %v", err)
	}
	// Non-existent path
	badRec := FileRecord{ID: 100, Path: filepath.Join(temp, "nope.txt")}
	if err := copyToDotSyncFilesByID(badRec, dotSyncFiles); err == nil {
		t.Error("expected error for non-existent path, got nil")
	}
}

func TestRunCmd(t *testing.T) {
	temp := t.TempDir()
	// Success
	if err := runCmd(temp, "echo", "hello"); err != nil {
		t.Errorf("runCmd echo failed: %v", err)
	}
	// Error
	if err := runCmd(temp, "nonexistent-cmd-xyz"); err == nil {
		t.Error("expected error for nonexistent command, got nil")
	}
}

// execCommandContext is a test helper for runCmd context simulation
func execCommandContext(ctx context.Context, dir, name string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	return cmd
}

// Note: ensureGitRepo and gitAddCommitPush are not tested here because they require a real git environment and remote.
// In a real project, use a mock or a test git repo, or abstract git commands for easier testing.
