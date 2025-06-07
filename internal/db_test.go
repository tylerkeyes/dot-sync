package internal

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
