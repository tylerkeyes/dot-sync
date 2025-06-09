package internal

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/tylerkeyes/dot-sync/internal/db"
)

func TestNewShowCmd(t *testing.T) {
	cmd := NewShowCmd()
	if cmd == nil {
		t.Fatal("NewShowCmd returned nil")
	}
	if !strings.Contains(cmd.Use, "show") {
		t.Errorf("expected Use to contain 'show', got %q", cmd.Use)
	}
	if cmd.Run == nil {
		t.Error("expected Run to be set")
	}
	if !strings.Contains(cmd.Short, "Show the paths") {
		t.Errorf("expected Short to contain description, got %q", cmd.Short)
	}
}

func TestShowHandlerNoFiles(t *testing.T) {
	// Create temporary home directory for test database
	oldHome := os.Getenv("HOME")
	tempHome := t.TempDir()
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", oldHome)

	// Create .dot-sync directory
	dotSyncPath := filepath.Join(tempHome, ".dot-sync")
	os.MkdirAll(dotSyncPath, 0700)

	// Capture stdout
	var buf bytes.Buffer
	orig := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := &cobra.Command{}
	showHandler(cmd, []string{})

	w.Close()
	os.Stdout = orig
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "No files currently tracked for syncing.") {
		t.Errorf("expected output to contain 'No files currently tracked for syncing.', got %q", output)
	}
}

func TestShowHandlerWithFiles(t *testing.T) {
	// Create temporary home directory for test database
	oldHome := os.Getenv("HOME")
	tempHome := t.TempDir()
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", oldHome)

	// Create .dot-sync directory
	dotSyncPath := filepath.Join(tempHome, ".dot-sync")
	os.MkdirAll(dotSyncPath, 0700)

	// Open database and add test files
	database, err := db.OpenDotSyncDB()
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	if err := db.EnsureFilesTable(database); err != nil {
		t.Fatalf("Failed to ensure files table: %v", err)
	}

	testPaths := []string{
		"/home/user/.bashrc",
		"/home/user/.vimrc",
		"/home/user/.config/git/config",
	}

	for _, path := range testPaths {
		if err := db.InsertFile(database, path); err != nil {
			t.Fatalf("Failed to insert file %s: %v", path, err)
		}
	}

	// Capture stdout
	var buf bytes.Buffer
	orig := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := &cobra.Command{}
	showHandler(cmd, []string{})

	w.Close()
	os.Stdout = orig
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "Files currently tracked for syncing (3):") {
		t.Errorf("expected output to contain file count, got %q", output)
	}

	for _, path := range testPaths {
		if !strings.Contains(output, path) {
			t.Errorf("expected output to contain path %s, got %q", path, output)
		}
	}
}

func TestShowHandlerSingleFile(t *testing.T) {
	// Create temporary home directory for test database
	oldHome := os.Getenv("HOME")
	tempHome := t.TempDir()
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", oldHome)

	// Create .dot-sync directory
	dotSyncPath := filepath.Join(tempHome, ".dot-sync")
	os.MkdirAll(dotSyncPath, 0700)

	// Open database and add single test file
	database, err := db.OpenDotSyncDB()
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	if err := db.EnsureFilesTable(database); err != nil {
		t.Fatalf("Failed to ensure files table: %v", err)
	}

	testPath := "/home/user/.bashrc"
	if err := db.InsertFile(database, testPath); err != nil {
		t.Fatalf("Failed to insert file: %v", err)
	}

	// Capture stdout
	var buf bytes.Buffer
	orig := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := &cobra.Command{}
	showHandler(cmd, []string{})

	w.Close()
	os.Stdout = orig
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "Files currently tracked for syncing (1):") {
		t.Errorf("expected output to contain single file count, got %q", output)
	}
	if !strings.Contains(output, testPath) {
		t.Errorf("expected output to contain path %s, got %q", testPath, output)
	}
}

func TestShowHandlerDatabaseError(t *testing.T) {
	// Set invalid home directory to cause database error
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", "/invalid/nonexistent/path")
	defer os.Setenv("HOME", oldHome)

	// Capture stdout
	var buf bytes.Buffer
	orig := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := &cobra.Command{}
	showHandler(cmd, []string{})

	w.Close()
	os.Stdout = orig
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "Failed to") {
		t.Errorf("expected database error message, got %q", output)
	}
}

func TestShowHandlerWithDuplicateFiles(t *testing.T) {
	// Create temporary home directory for test database
	oldHome := os.Getenv("HOME")
	tempHome := t.TempDir()
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", oldHome)

	// Create .dot-sync directory
	dotSyncPath := filepath.Join(tempHome, ".dot-sync")
	os.MkdirAll(dotSyncPath, 0700)

	// Open database and add test files (including attempts at duplicates)
	database, err := db.OpenDotSyncDB()
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	if err := db.EnsureFilesTable(database); err != nil {
		t.Fatalf("Failed to ensure files table: %v", err)
	}

	testPath := "/home/user/.bashrc"
	// Try to insert the same file multiple times
	if err := db.InsertFile(database, testPath); err != nil {
		t.Fatalf("Failed to insert file: %v", err)
	}
	if err := db.InsertFile(database, testPath); err != nil {
		t.Fatalf("Failed to insert file (second time): %v", err)
	}

	// Capture stdout
	var buf bytes.Buffer
	orig := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := &cobra.Command{}
	showHandler(cmd, []string{})

	w.Close()
	os.Stdout = orig
	buf.ReadFrom(r)
	output := buf.String()

	// The database currently allows duplicates since there's no UNIQUE constraint on path
	// This test verifies that the show command displays all entries as they exist in the database
	if !strings.Contains(output, "Files currently tracked for syncing (2):") {
		t.Errorf("expected output to show duplicate file count, got %q", output)
	}
	// Count occurrences of the test path in output (should be 2 due to duplicates)
	pathCount := strings.Count(output, testPath)
	if pathCount != 2 {
		t.Errorf("expected path to appear exactly twice due to duplicates, but appeared %d times", pathCount)
	}
}
