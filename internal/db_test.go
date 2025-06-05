package internal

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestFindHomeDir(t *testing.T) {
	home := FindHomeDir()
	if home == "" {
		t.Error("FindHomeDir returned empty string")
	}
	if _, err := os.Stat(home); err != nil {
		t.Errorf("FindHomeDir returned non-existent directory: %v", err)
	}
}

func TestEnsureFilesTableAndInsertFile(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory sqlite db: %v", err)
	}
	defer db.Close()

	if err := EnsureFilesTable(db); err != nil {
		t.Fatalf("failed to ensure files table: %v", err)
	}

	path := "/tmp/testfile.txt"
	if err := InsertFile(db, path); err != nil {
		t.Fatalf("failed to insert file: %v", err)
	}

	var count int
	row := db.QueryRow("SELECT COUNT(*) FROM files WHERE path = ?", path)
	if err := row.Scan(&count); err != nil {
		t.Fatalf("failed to query inserted file: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 file inserted, got %d", count)
	}
}
