package internal

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

func OpenDotSyncDB() (*sql.DB, error) {
	homeDir := FindHomeDir()
	if homeDir == "" {
		return nil, fmt.Errorf("could not determine home directory")
	}
	dbPath := filepath.Join(homeDir, ".dot-sync.db")
	return sql.Open("sqlite3", dbPath)
}

func EnsureFilesTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS files (id INTEGER PRIMARY KEY AUTOINCREMENT, path TEXT)`)
	return err
}

func InsertFile(db *sql.DB, path string) error {
	_, err := db.Exec(`INSERT OR IGNORE INTO files (path) VALUES (?)`, path)
	return err
}

func InsertFiles(db *sql.DB, paths []string) error {
	if len(paths) == 0 {
		return nil
	}
	// Use a transaction for batch insert
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(`INSERT OR IGNORE INTO files (path) VALUES (?)`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()
	for _, path := range paths {
		if _, err := stmt.Exec(path); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func FindHomeDir() string {
	baseDir := ""
	if h := os.Getenv("HOME"); h != "" {
		baseDir = h
	} else if u, err := os.UserHomeDir(); err == nil {
		baseDir = u
	}
	return baseDir
}
