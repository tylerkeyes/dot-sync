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
