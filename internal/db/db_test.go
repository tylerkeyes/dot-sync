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
	storageType := "git"
	remote := "https://github.com/user/repo.git"
	InsertStorageProvider(db, storageType, remote)

	// Update
	newStorageType := "git"
	newRemote := "https://github.com/user/newrepo.git"
	if err := UpdateStorageProvider(db, newStorageType, newRemote); err != nil {
		t.Errorf("UpdateStorageProvider failed: %v", err)
	}

	// Verify update
	retrievedType, retrievedRemote, err := GetStorageProvider(db)
	if err != nil {
		t.Errorf("failed to retrieve updated storage provider: %v", err)
	}
	if retrievedType != newStorageType || retrievedRemote != newRemote {
		t.Errorf("expected %s/%s, got %s/%s", newStorageType, newRemote, retrievedType, retrievedRemote)
	}

	// Update with no existing record should not error but won't affect anything
	db2, _ := sql.Open("sqlite3", ":memory:")
	defer db2.Close()
	EnsureStorageTable(db2)
	if err := UpdateStorageProvider(db2, storageType, remote); err != nil {
		t.Errorf("UpdateStorageProvider with no existing record failed: %v", err)
	}

	// Error on bad DB
	badDB, _ := sql.Open("sqlite3", ":memory:")
	badDB.Close()
	if err := UpdateStorageProvider(badDB, storageType, remote); err == nil {
		t.Error("expected error for closed DB, got nil")
	}
}
