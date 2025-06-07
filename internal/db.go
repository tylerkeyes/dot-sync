package internal

import (
	"database/sql"
	"fmt"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"github.com/tylerkeyes/dot-sync/internal/shared"
)

const dotSyncDBName = ".dot-sync/state.db"

type FileRecord struct {
	ID   int
	Path string
}

func OpenDotSyncDB() (*sql.DB, error) {
	homeDir := shared.FindHomeDir()
	if homeDir == "" {
		return nil, fmt.Errorf("could not determine home directory")
	}
	dbPath := filepath.Join(homeDir, dotSyncDBName)
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

func GetAllFilePaths(db *sql.DB) ([]FileRecord, error) {
	rows, err := db.Query("SELECT id, path FROM files")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var records []FileRecord
	for rows.Next() {
		var rec FileRecord
		if err := rows.Scan(&rec.ID, &rec.Path); err != nil {
			return nil, err
		}
		records = append(records, rec)
	}
	return records, rows.Err()
}

func EnsureStorageTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS storage_provider (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		storage_type TEXT,
		remote TEXT
	)`)
	return err
}

func InsertStorageProvider(db *sql.DB, storageType, remote string) error {
	_, err := db.Exec(`INSERT INTO storage_provider (storage_type, remote) VALUES (?, ?)`, storageType, remote)
	return err
}
