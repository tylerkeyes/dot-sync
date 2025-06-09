package db

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"

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
	// Convert absolute path to storage path before inserting
	storagePath := shared.ToStoragePath(path)
	_, err := db.Exec(`INSERT OR IGNORE INTO files (path) VALUES (?)`, storagePath)
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
		// Convert absolute path to storage path before inserting
		storagePath := shared.ToStoragePath(path)
		if _, err := stmt.Exec(storagePath); err != nil {
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
		// Convert storage path back to absolute path when retrieving
		rec.Path = shared.FromStoragePath(rec.Path)
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

func GetStorageProvider(db *sql.DB) (string, string, error) {
	row := db.QueryRow("SELECT storage_type, remote FROM storage_provider ORDER BY id DESC LIMIT 1")
	var storageType, remote string
	if err := row.Scan(&storageType, &remote); err != nil {
		return "", "", err
	}
	return storageType, remote, nil
}

func UpdateStorageProvider(db *sql.DB, storageType, remote string) error {
	_, err := db.Exec(`UPDATE storage_provider SET storage_type = ?, remote = ? WHERE id = (SELECT id FROM storage_provider ORDER BY id DESC LIMIT 1)`, storageType, remote)
	return err
}

func GetFileRecordsByPaths(db *sql.DB, paths []string) ([]FileRecord, error) {
	if len(paths) == 0 {
		return []FileRecord{}, nil
	}

	// Convert absolute paths to storage paths for querying
	storagePaths := make([]string, len(paths))
	for i, path := range paths {
		storagePaths[i] = shared.ToStoragePath(path)
	}

	// Build the query with placeholders for the IN clause
	query := "SELECT id, path FROM files WHERE path IN ("
	placeholders := make([]string, len(storagePaths))
	args := make([]interface{}, len(storagePaths))

	for i, storagePath := range storagePaths {
		placeholders[i] = "?"
		args[i] = storagePath
	}

	query += strings.Join(placeholders, ",") + ")"

	rows, err := db.Query(query, args...)
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
		// Convert storage path back to absolute path when retrieving
		rec.Path = shared.FromStoragePath(rec.Path)
		records = append(records, rec)
	}
	return records, rows.Err()
}

func DeleteFilesByIDs(db *sql.DB, ids []int) error {
	if len(ids) == 0 {
		return nil
	}

	// Use a transaction for batch delete
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("DELETE FROM files WHERE id = ?")
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, id := range ids {
		if _, err := stmt.Exec(id); err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
