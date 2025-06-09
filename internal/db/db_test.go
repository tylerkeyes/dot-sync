package db

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestOpenDotSyncDB(t *testing.T) {
	oldHome := os.Getenv("HOME")
	temp := t.TempDir()
	os.Setenv("HOME", temp)
	db, err := OpenDotSyncDB()
	if err != nil {
		t.Fatalf("OpenDotSyncDB failed: %v", err)
	}
	db.Close()
	os.Unsetenv("HOME")
	// Simulate missing home dir
	os.Setenv("HOME", "")
	_, err = OpenDotSyncDB()
	if err == nil {
		t.Error("expected error for missing home dir, got nil")
	}
	os.Setenv("HOME", oldHome)
}

func TestEnsureFilesTable(t *testing.T) {
	db, _ := sql.Open("sqlite3", ":memory:")
	defer db.Close()
	if err := EnsureFilesTable(db); err != nil {
		t.Errorf("EnsureFilesTable failed: %v", err)
	}
	// Error on bad DB
	badDB, _ := sql.Open("sqlite3", ":memory:")
	badDB.Close()
	if err := EnsureFilesTable(badDB); err == nil {
		t.Error("expected error for closed DB, got nil")
	}
}

func TestInsertFile(t *testing.T) {
	db, _ := sql.Open("sqlite3", ":memory:")
	defer db.Close()
	EnsureFilesTable(db)
	path := "/tmp/foo.txt"
	if err := InsertFile(db, path); err != nil {
		t.Errorf("InsertFile failed: %v", err)
	}
	// Duplicate insert
	if err := InsertFile(db, path); err != nil {
		t.Errorf("InsertFile duplicate failed: %v", err)
	}
	// Error on bad DB
	badDB, _ := sql.Open("sqlite3", ":memory:")
	badDB.Close()
	if err := InsertFile(badDB, path); err == nil {
		t.Error("expected error for closed DB, got nil")
	}
}

func TestInsertFiles(t *testing.T) {
	db, _ := sql.Open("sqlite3", ":memory:")
	defer db.Close()
	EnsureFilesTable(db)
	paths := []string{"/tmp/a.txt", "/tmp/b.txt"}
	if err := InsertFiles(db, paths); err != nil {
		t.Errorf("InsertFiles failed: %v", err)
	}
	// Empty input
	if err := InsertFiles(db, []string{}); err != nil {
		t.Errorf("InsertFiles with empty input failed: %v", err)
	}
	// Error on bad DB
	badDB, _ := sql.Open("sqlite3", ":memory:")
	badDB.Close()
	if err := InsertFiles(badDB, paths); err == nil {
		t.Error("expected error for closed DB, got nil")
	}
}

func TestGetAllFilePaths(t *testing.T) {
	db, _ := sql.Open("sqlite3", ":memory:")
	defer db.Close()
	EnsureFilesTable(db)
	// No files
	records, err := GetAllFilePaths(db)
	if err != nil {
		t.Errorf("GetAllFilePaths failed: %v", err)
	}
	if len(records) != 0 {
		t.Errorf("expected 0 records, got %v", records)
	}
	// With files
	InsertFile(db, "/tmp/1.txt")
	InsertFile(db, "/tmp/2.txt")
	records, err = GetAllFilePaths(db)
	if err != nil {
		t.Errorf("GetAllFilePaths failed: %v", err)
	}
	// Check that both paths are present and IDs are unique
	found := map[string]bool{"/tmp/1.txt": false, "/tmp/2.txt": false}
	ids := map[int]bool{}
	for _, rec := range records {
		if _, ok := found[rec.Path]; ok {
			found[rec.Path] = true
		}
		if ids[rec.ID] {
			t.Errorf("duplicate ID found: %d", rec.ID)
		}
		ids[rec.ID] = true
	}
	for path, ok := range found {
		if !ok {
			t.Errorf("expected path %q not found in records", path)
		}
	}
	// Error on bad DB
	badDB, _ := sql.Open("sqlite3", ":memory:")
	badDB.Close()
	_, err = GetAllFilePaths(badDB)
	if err == nil {
		t.Error("expected error for closed DB, got nil")
	}
}

func TestEnsureStorageTable(t *testing.T) {
	db, _ := sql.Open("sqlite3", ":memory:")
	defer db.Close()
	if err := EnsureStorageTable(db); err != nil {
		t.Errorf("EnsureStorageTable failed: %v", err)
	}
	// Verify table exists by querying it
	_, err := db.Query("SELECT id, storage_type, remote FROM storage_provider LIMIT 0")
	if err != nil {
		t.Errorf("storage_provider table not created properly: %v", err)
	}
	// Error on bad DB
	badDB, _ := sql.Open("sqlite3", ":memory:")
	badDB.Close()
	if err := EnsureStorageTable(badDB); err == nil {
		t.Error("expected error for closed DB, got nil")
	}
}

func TestInsertStorageProvider(t *testing.T) {
	db, _ := sql.Open("sqlite3", ":memory:")
	defer db.Close()
	EnsureStorageTable(db)

	storageType := "git"
	remote := "https://github.com/user/repo.git"
	if err := InsertStorageProvider(db, storageType, remote); err != nil {
		t.Errorf("InsertStorageProvider failed: %v", err)
	}

	// Verify insertion
	row := db.QueryRow("SELECT storage_type, remote FROM storage_provider WHERE storage_type = ? AND remote = ?", storageType, remote)
	var retrievedType, retrievedRemote string
	if err := row.Scan(&retrievedType, &retrievedRemote); err != nil {
		t.Errorf("failed to retrieve inserted storage provider: %v", err)
	}
	if retrievedType != storageType || retrievedRemote != remote {
		t.Errorf("expected %s/%s, got %s/%s", storageType, remote, retrievedType, retrievedRemote)
	}

	// Error on bad DB
	badDB, _ := sql.Open("sqlite3", ":memory:")
	badDB.Close()
	if err := InsertStorageProvider(badDB, storageType, remote); err == nil {
		t.Error("expected error for closed DB, got nil")
	}
}

func TestGetStorageProvider(t *testing.T) {
	db, _ := sql.Open("sqlite3", ":memory:")
	defer db.Close()
	EnsureStorageTable(db)

	// No storage provider exists
	_, _, err := GetStorageProvider(db)
	if err == nil {
		t.Error("expected error for missing storage provider, got nil")
	}

	// Insert and retrieve
	storageType := "git"
	remote := "https://github.com/user/repo.git"
	InsertStorageProvider(db, storageType, remote)

	retrievedType, retrievedRemote, err := GetStorageProvider(db)
	if err != nil {
		t.Errorf("GetStorageProvider failed: %v", err)
	}
	if retrievedType != storageType || retrievedRemote != remote {
		t.Errorf("expected %s/%s, got %s/%s", storageType, remote, retrievedType, retrievedRemote)
	}

	// Insert another and ensure we get the latest
	newRemote := "https://github.com/user/newrepo.git"
	InsertStorageProvider(db, storageType, newRemote)

	retrievedType, retrievedRemote, err = GetStorageProvider(db)
	if err != nil {
		t.Errorf("GetStorageProvider failed after second insert: %v", err)
	}
	if retrievedRemote != newRemote {
		t.Errorf("expected latest remote %s, got %s", newRemote, retrievedRemote)
	}

	// Error on bad DB
	badDB, _ := sql.Open("sqlite3", ":memory:")
	badDB.Close()
	_, _, err = GetStorageProvider(badDB)
	if err == nil {
		t.Error("expected error for closed DB, got nil")
	}
}

func TestUpdateStorageProvider(t *testing.T) {
	db, _ := sql.Open("sqlite3", ":memory:")
	defer db.Close()
	EnsureStorageTable(db)

	// Insert initial provider
	if err := InsertStorageProvider(db, "git", "initial-url"); err != nil {
		t.Fatalf("Failed to insert initial storage provider: %v", err)
	}

	// Update provider
	if err := UpdateStorageProvider(db, "git", "updated-url"); err != nil {
		t.Fatalf("Failed to update storage provider: %v", err)
	}

	// Verify update
	storageType, remote, err := GetStorageProvider(db)
	if err != nil {
		t.Fatalf("Failed to get updated storage provider: %v", err)
	}
	if storageType != "git" || remote != "updated-url" {
		t.Errorf("Expected git/updated-url, got %s/%s", storageType, remote)
	}
}

func TestGetFileRecordsByPaths(t *testing.T) {
	db, _ := sql.Open("sqlite3", ":memory:")
	defer db.Close()
	EnsureFilesTable(db)

	// Insert test files
	testPaths := []string{"/home/user/.bashrc", "/home/user/.vimrc", "/home/user/.config/git/config"}
	for _, path := range testPaths {
		if err := InsertFile(db, path); err != nil {
			t.Fatalf("Failed to insert file %s: %v", path, err)
		}
	}

	// Test getting existing paths
	records, err := GetFileRecordsByPaths(db, []string{"/home/user/.bashrc", "/home/user/.vimrc"})
	if err != nil {
		t.Fatalf("Failed to get file records: %v", err)
	}
	if len(records) != 2 {
		t.Errorf("Expected 2 records, got %d", len(records))
	}

	// Verify the correct paths are returned
	pathMap := make(map[string]bool)
	for _, record := range records {
		pathMap[record.Path] = true
	}
	if !pathMap["/home/user/.bashrc"] || !pathMap["/home/user/.vimrc"] {
		t.Error("Expected paths not found in records")
	}

	// Test getting non-existent paths
	records, err = GetFileRecordsByPaths(db, []string{"/nonexistent/path"})
	if err != nil {
		t.Fatalf("Failed to get file records for non-existent path: %v", err)
	}
	if len(records) != 0 {
		t.Errorf("Expected 0 records for non-existent path, got %d", len(records))
	}

	// Test empty input
	records, err = GetFileRecordsByPaths(db, []string{})
	if err != nil {
		t.Fatalf("Failed to get file records for empty input: %v", err)
	}
	if len(records) != 0 {
		t.Errorf("Expected 0 records for empty input, got %d", len(records))
	}

	// Test mixed existing and non-existing paths
	records, err = GetFileRecordsByPaths(db, []string{"/home/user/.bashrc", "/nonexistent/path", "/home/user/.config/git/config"})
	if err != nil {
		t.Fatalf("Failed to get file records for mixed paths: %v", err)
	}
	if len(records) != 2 {
		t.Errorf("Expected 2 records for mixed paths, got %d", len(records))
	}
}

func TestDeleteFilesByIDs(t *testing.T) {
	db, _ := sql.Open("sqlite3", ":memory:")
	defer db.Close()
	EnsureFilesTable(db)

	// Insert test files
	testPaths := []string{"/home/user/.bashrc", "/home/user/.vimrc", "/home/user/.config/git/config"}
	for _, path := range testPaths {
		if err := InsertFile(db, path); err != nil {
			t.Fatalf("Failed to insert file %s: %v", path, err)
		}
	}

	// Get all records to find their IDs
	allRecords, err := GetAllFilePaths(db)
	if err != nil {
		t.Fatalf("Failed to get all file paths: %v", err)
	}
	if len(allRecords) != 3 {
		t.Fatalf("Expected 3 records, got %d", len(allRecords))
	}

	// Delete first two files by ID
	idsToDelete := []int{allRecords[0].ID, allRecords[1].ID}
	if err := DeleteFilesByIDs(db, idsToDelete); err != nil {
		t.Fatalf("Failed to delete files by IDs: %v", err)
	}

	// Verify only one record remains
	remainingRecords, err := GetAllFilePaths(db)
	if err != nil {
		t.Fatalf("Failed to get remaining file paths: %v", err)
	}
	if len(remainingRecords) != 1 {
		t.Errorf("Expected 1 remaining record, got %d", len(remainingRecords))
	}
	if remainingRecords[0].ID != allRecords[2].ID {
		t.Errorf("Expected remaining record to have ID %d, got %d", allRecords[2].ID, remainingRecords[0].ID)
	}

	// Test deleting with empty list
	if err := DeleteFilesByIDs(db, []int{}); err != nil {
		t.Fatalf("Failed to delete with empty list: %v", err)
	}

	// Verify record count unchanged
	recordsAfterEmpty, err := GetAllFilePaths(db)
	if err != nil {
		t.Fatalf("Failed to get file paths after empty delete: %v", err)
	}
	if len(recordsAfterEmpty) != 1 {
		t.Errorf("Expected 1 record after empty delete, got %d", len(recordsAfterEmpty))
	}

	// Test deleting non-existent ID
	if err := DeleteFilesByIDs(db, []int{99999}); err != nil {
		t.Fatalf("Failed to delete non-existent ID: %v", err)
	}

	// Verify record count unchanged
	recordsAfterNonExistent, err := GetAllFilePaths(db)
	if err != nil {
		t.Fatalf("Failed to get file paths after non-existent delete: %v", err)
	}
	if len(recordsAfterNonExistent) != 1 {
		t.Errorf("Expected 1 record after non-existent delete, got %d", len(recordsAfterNonExistent))
	}
}
