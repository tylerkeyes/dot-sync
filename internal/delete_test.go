package internal

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/tylerkeyes/dot-sync/internal/db"
)

func TestNewDeleteCmd(t *testing.T) {
	cmd := NewDeleteCmd()
	if cmd == nil {
		t.Fatal("NewDeleteCmd returned nil")
	}
	if !strings.Contains(cmd.Use, "delete") {
		t.Errorf("expected Use to contain 'delete', got %q", cmd.Use)
	}
	if cmd.Run == nil {
		t.Error("expected Run to be set")
	}
	if !strings.Contains(cmd.Short, "Delete files from sync tracking") {
		t.Errorf("expected Short to contain description, got %q", cmd.Short)
	}
	if cmd.Args == nil {
		t.Error("expected Args to be set")
	}
}

func TestDeleteHandlerNoArgs(t *testing.T) {
	// This should be caught by cobra's MinimumNArgs, but test handler directly
	// Create temporary home directory for test database
	oldHome := os.Getenv("HOME")
	tempHome := t.TempDir()
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", oldHome)

	// Create .dot-sync directory
	os.MkdirAll(filepath.Join(tempHome, ".dot-sync"), 0700)

	// Capture stdout
	var buf bytes.Buffer
	orig := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := &cobra.Command{}
	deleteHandler(cmd, []string{})

	w.Close()
	os.Stdout = orig
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "No matching files found in tracking database.") {
		t.Errorf("expected output to contain 'No matching files found', got %q", output)
	}
}

func TestDeleteHandlerNoMatchingFiles(t *testing.T) {
	// Create temporary home directory for test database
	oldHome := os.Getenv("HOME")
	tempHome := t.TempDir()
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", oldHome)

	// Create .dot-sync directory
	os.MkdirAll(filepath.Join(tempHome, ".dot-sync"), 0700)

	// Capture stdout
	var buf bytes.Buffer
	orig := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := &cobra.Command{}
	deleteHandler(cmd, []string{"/nonexistent/path"})

	w.Close()
	os.Stdout = orig
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "No matching files found in tracking database.") {
		t.Errorf("expected output to contain 'No matching files found', got %q", output)
	}
}

func TestDeleteHandlerWithMatchingFiles(t *testing.T) {
	// Create temporary home directory for test database
	oldHome := os.Getenv("HOME")
	tempHome := t.TempDir()
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", oldHome)

	// Create .dot-sync directory and files directory
	dotSyncFilesPath := filepath.Join(tempHome, ".dot-sync", "files")
	os.MkdirAll(dotSyncFilesPath, 0700)

	// Open database and add test files
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

	// Get the record to find its ID
	records, err := db.GetFileRecordsByPaths(database, []string{testPath})
	if err != nil {
		t.Fatalf("Failed to get file records: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(records))
	}

	// Create the file in .dot-sync/files directory
	testFileInStorage := filepath.Join(dotSyncFilesPath, fmt.Sprintf("%d", records[0].ID))
	if err := os.WriteFile(testFileInStorage, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Capture stdout
	var buf bytes.Buffer
	orig := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := &cobra.Command{}
	deleteHandler(cmd, []string{testPath})

	w.Close()
	os.Stdout = orig
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "Successfully deleted 1 file(s) from tracking:") {
		t.Errorf("expected output to contain success message, got %q", output)
	}
	if !strings.Contains(output, testPath) {
		t.Errorf("expected output to contain path %s, got %q", testPath, output)
	}

	// Verify file was deleted from storage
	if _, err := os.Stat(testFileInStorage); !os.IsNotExist(err) {
		t.Error("expected file to be deleted from storage")
	}

	// Verify record was deleted from database
	remainingRecords, err := db.GetAllFilePaths(database)
	if err != nil {
		t.Fatalf("Failed to get remaining records: %v", err)
	}
	if len(remainingRecords) != 0 {
		t.Errorf("expected 0 remaining records, got %d", len(remainingRecords))
	}
}

func TestDeleteHandlerMultipleFiles(t *testing.T) {
	// Create temporary home directory for test database
	oldHome := os.Getenv("HOME")
	tempHome := t.TempDir()
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", oldHome)

	// Create .dot-sync directory and files directory
	dotSyncFilesPath := filepath.Join(tempHome, ".dot-sync", "files")
	os.MkdirAll(dotSyncFilesPath, 0700)

	// Open database and add test files
	database, err := db.OpenDotSyncDB()
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	if err := db.EnsureFilesTable(database); err != nil {
		t.Fatalf("Failed to ensure files table: %v", err)
	}

	testPaths := []string{"/home/user/.bashrc", "/home/user/.vimrc"}
	for _, path := range testPaths {
		if err := db.InsertFile(database, path); err != nil {
			t.Fatalf("Failed to insert file %s: %v", path, err)
		}
	}

	// Get the records to find their IDs
	records, err := db.GetFileRecordsByPaths(database, testPaths)
	if err != nil {
		t.Fatalf("Failed to get file records: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("Expected 2 records, got %d", len(records))
	}

	// Create the files in .dot-sync/files directory
	for _, record := range records {
		testFileInStorage := filepath.Join(dotSyncFilesPath, fmt.Sprintf("%d", record.ID))
		if err := os.WriteFile(testFileInStorage, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Capture stdout
	var buf bytes.Buffer
	orig := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := &cobra.Command{}
	deleteHandler(cmd, testPaths)

	w.Close()
	os.Stdout = orig
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "Successfully deleted 2 file(s) from tracking:") {
		t.Errorf("expected output to contain success message for 2 files, got %q", output)
	}

	// Verify all files were deleted from storage
	for _, record := range records {
		testFileInStorage := filepath.Join(dotSyncFilesPath, fmt.Sprintf("%d", record.ID))
		if _, err := os.Stat(testFileInStorage); !os.IsNotExist(err) {
			t.Errorf("expected file %d to be deleted from storage", record.ID)
		}
	}

	// Verify records were deleted from database
	remainingRecords, err := db.GetAllFilePaths(database)
	if err != nil {
		t.Fatalf("Failed to get remaining records: %v", err)
	}
	if len(remainingRecords) != 0 {
		t.Errorf("expected 0 remaining records, got %d", len(remainingRecords))
	}
}

func TestDeleteHandlerPartialMatch(t *testing.T) {
	// Create temporary home directory for test database
	oldHome := os.Getenv("HOME")
	tempHome := t.TempDir()
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", oldHome)

	// Create .dot-sync directory and files directory
	dotSyncFilesPath := filepath.Join(tempHome, ".dot-sync", "files")
	os.MkdirAll(dotSyncFilesPath, 0700)

	// Open database and add test files
	database, err := db.OpenDotSyncDB()
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	if err := db.EnsureFilesTable(database); err != nil {
		t.Fatalf("Failed to ensure files table: %v", err)
	}

	existingPath := "/home/user/.bashrc"
	nonExistentPath := "/home/user/nonexistent"

	if err := db.InsertFile(database, existingPath); err != nil {
		t.Fatalf("Failed to insert file: %v", err)
	}

	// Get the record to find its ID
	records, err := db.GetFileRecordsByPaths(database, []string{existingPath})
	if err != nil {
		t.Fatalf("Failed to get file records: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(records))
	}

	// Create the file in .dot-sync/files directory
	testFileInStorage := filepath.Join(dotSyncFilesPath, fmt.Sprintf("%d", records[0].ID))
	if err := os.WriteFile(testFileInStorage, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Capture stdout
	var buf bytes.Buffer
	orig := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := &cobra.Command{}
	deleteHandler(cmd, []string{existingPath, nonExistentPath})

	w.Close()
	os.Stdout = orig
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "Successfully deleted 1 file(s) from tracking:") {
		t.Errorf("expected output to contain success message for 1 file, got %q", output)
	}
	if !strings.Contains(output, existingPath) {
		t.Errorf("expected output to contain existing path %s, got %q", existingPath, output)
	}

	// Verify file was deleted from storage
	if _, err := os.Stat(testFileInStorage); !os.IsNotExist(err) {
		t.Error("expected file to be deleted from storage")
	}
}

func TestDeleteHandlerMissingStorageFile(t *testing.T) {
	// Create temporary home directory for test database
	oldHome := os.Getenv("HOME")
	tempHome := t.TempDir()
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", oldHome)

	// Create .dot-sync directory and files directory
	dotSyncFilesPath := filepath.Join(tempHome, ".dot-sync", "files")
	os.MkdirAll(dotSyncFilesPath, 0700)

	// Open database and add test files
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

	// Don't create the file in .dot-sync/files directory (simulate missing file)

	// Capture stdout
	var buf bytes.Buffer
	orig := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := &cobra.Command{}
	deleteHandler(cmd, []string{testPath})

	w.Close()
	os.Stdout = orig
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "Warning: File with ID") && !strings.Contains(output, "not found in storage") {
		t.Errorf("expected warning about missing storage file, got %q", output)
	}
	if !strings.Contains(output, "Successfully deleted 1 file(s) from tracking:") {
		t.Errorf("expected success message despite missing storage file, got %q", output)
	}

	// Verify record was still deleted from database
	remainingRecords, err := db.GetAllFilePaths(database)
	if err != nil {
		t.Fatalf("Failed to get remaining records: %v", err)
	}
	if len(remainingRecords) != 0 {
		t.Errorf("expected 0 remaining records, got %d", len(remainingRecords))
	}
}

func TestDeleteHandlerDatabaseError(t *testing.T) {
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
	deleteHandler(cmd, []string{"/test/path"})

	w.Close()
	os.Stdout = orig
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "Failed to") {
		t.Errorf("expected database error message, got %q", output)
	}
}
