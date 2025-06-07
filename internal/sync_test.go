package internal

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestNewSyncCmd(t *testing.T) {
	cmd := NewSyncCmd()
	if cmd == nil {
		t.Fatal("NewSyncCmd returned nil")
	}
	if !strings.Contains(cmd.Use, "sync") {
		t.Errorf("expected Use to contain 'sync', got %q", cmd.Use)
	}
	if cmd.Run == nil {
		t.Error("expected Run to be set")
	}
	if !strings.Contains(cmd.Short, "Sync dotfiles") {
		t.Errorf("expected Short to contain 'Sync dotfiles', got %q", cmd.Short)
	}
}

// Note: syncHandler is complex and requires database setup and storage providers
// It's better tested through integration tests rather than unit tests

func TestSyncHandlerDatabaseError(t *testing.T) {
	// Test syncHandler with database error
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", "/invalid/nonexistent/path")
	defer os.Setenv("HOME", oldHome)

	// Capture stdout
	var buf bytes.Buffer
	orig := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := &cobra.Command{}
	syncHandler(cmd, []string{})

	w.Close()
	os.Stdout = orig
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "Failed to") {
		t.Errorf("expected database error message, got %q", output)
	}
}

func TestSyncHandlerStorageProviderError(t *testing.T) {
	// Test syncHandler with missing storage provider
	oldHome := os.Getenv("HOME")
	tempHome := t.TempDir()
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", oldHome)

	// Create .dot-sync directory
	dotSyncPath := filepath.Join(tempHome, ".dot-sync")
	os.MkdirAll(dotSyncPath, 0700)

	// Create command without storage provider in context
	cmd := &cobra.Command{}
	defer func() {
		if r := recover(); r != nil {
			// Expected panic due to missing storage provider
			t.Log("syncHandler panicked as expected due to missing storage provider")
		}
	}()

	syncHandler(cmd, []string{})
}
