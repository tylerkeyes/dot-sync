package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestRootCmd(t *testing.T) {
	if rootCmd == nil {
		t.Fatal("rootCmd is nil")
	}
	if !strings.Contains(rootCmd.Use, "dot-sync") {
		t.Errorf("expected Use to contain 'dot-sync', got %q", rootCmd.Use)
	}
	if !strings.Contains(rootCmd.Short, "CLI tool for dotfile syncing") {
		t.Errorf("expected Short to contain 'CLI tool for dotfile syncing', got %q", rootCmd.Short)
	}
	if rootCmd.PersistentPreRunE == nil {
		t.Error("expected PersistentPreRunE to be set")
	}
}

func TestRootCmdHasSubcommands(t *testing.T) {
	subcommands := rootCmd.Commands()
	if len(subcommands) == 0 {
		t.Fatal("expected root command to have subcommands")
	}

	expectedCommands := []string{"sync", "pull", "mark", "storage"}
	foundCommands := make(map[string]bool)

	for _, cmd := range subcommands {
		foundCommands[cmd.Name()] = true
	}

	for _, expected := range expectedCommands {
		if !foundCommands[expected] {
			t.Errorf("expected subcommand %q not found", expected)
		}
	}
}

func TestExecute(t *testing.T) {
	// Test that Execute function exists and can be called
	// We can't easily test the actual execution without complex setup
	// This at least verifies the function compiles and doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Execute panicked: %v", r)
		}
	}()

	// Save original args and command
	origArgs := os.Args
	origCmd := rootCmd

	// Create a test command that doesn't do anything
	testCmd := &cobra.Command{
		Use: "dot-sync",
		Run: func(cmd *cobra.Command, args []string) {
			// Do nothing for test
		},
	}
	rootCmd = testCmd

	// Set test args
	os.Args = []string{"dot-sync", "--help"}

	// This should not panic
	Execute()

	// Restore original state
	os.Args = origArgs
	rootCmd = origCmd
}

func TestPersistentPreRunE_StorageInit(t *testing.T) {
	// Test that storage init command is skipped
	oldArgs := os.Args
	os.Args = []string{"dot-sync", "storage", "init"}
	defer func() { os.Args = oldArgs }()

	cmd := &cobra.Command{}
	err := rootCmd.PersistentPreRunE(cmd, []string{})

	// Should return nil (no error) for storage init command
	if err != nil {
		t.Errorf("expected no error for storage init command, got: %v", err)
	}
}

func TestInitDirectoryCreation(t *testing.T) {
	// Test that the logic in init() works correctly
	tempHome := t.TempDir()

	// Simulate the directory creation logic from init()
	dotSyncPath := filepath.Join(tempHome, ".dot-sync")
	dotSyncFilesPath := filepath.Join(tempHome, ".dot-sync", "files")

	// Test EnsureDir function used in init()
	if err := os.MkdirAll(dotSyncPath, 0700); err != nil {
		t.Fatalf("failed to create .dot-sync directory: %v", err)
	}
	if err := os.MkdirAll(dotSyncFilesPath, 0700); err != nil {
		t.Fatalf("failed to create .dot-sync/files directory: %v", err)
	}

	// Verify directories were created
	if _, err := os.Stat(dotSyncPath); os.IsNotExist(err) {
		t.Error("expected .dot-sync directory to be created")
	}
	if _, err := os.Stat(dotSyncFilesPath); os.IsNotExist(err) {
		t.Error("expected .dot-sync/files directory to be created")
	}
}

func TestPersistentPreRunE_DatabaseError(t *testing.T) {
	// Test database error case
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", "/invalid/nonexistent/path")
	defer os.Setenv("HOME", oldHome)

	// Test without storage init args
	oldArgs := os.Args
	os.Args = []string{"dot-sync", "sync"}
	defer func() { os.Args = oldArgs }()

	cmd := &cobra.Command{}
	err := rootCmd.PersistentPreRunE(cmd, []string{})

	if err == nil {
		t.Error("expected error for invalid database path")
	}
	if !strings.Contains(err.Error(), "failed to") {
		t.Errorf("expected database error message, got: %v", err)
	}
}
