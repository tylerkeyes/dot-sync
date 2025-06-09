package internal

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/tylerkeyes/dot-sync/internal/db"
	"github.com/tylerkeyes/dot-sync/internal/shared"
)

func NewDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete [files or directories]...",
		Short: "Delete files from sync tracking and remove them from storage",
		Args:  cobra.MinimumNArgs(1),
		Run:   deleteHandler,
	}
}

func deleteHandler(cmd *cobra.Command, args []string) {
	database, err := db.OpenDotSyncDB()
	if err != nil {
		fmt.Println("Failed to open .dot-sync.db:", err)
		return
	}
	defer database.Close()

	if err := db.EnsureFilesTable(database); err != nil {
		fmt.Println("Failed to ensure files table:", err)
		return
	}

	// Convert args to full paths for consistency
	absPaths := argsAsFullPaths(args)

	// Find records that match the provided paths
	records, err := db.GetFileRecordsByPaths(database, absPaths)
	if err != nil {
		fmt.Printf("Failed to query file records: %v\n", err)
		return
	}

	if len(records) == 0 {
		fmt.Println("No matching files found in tracking database.")
		return
	}

	// Get .dot-sync/files directory path
	dotSyncFilesPath := filepath.Join(shared.FindHomeDir(), shared.GetDotSyncFilesDir())

	// Delete files from .dot-sync/files directory
	var deletedPaths []string
	var failedPaths []string
	var deletedIDs []int

	for _, record := range records {
		filePath := filepath.Join(dotSyncFilesPath, fmt.Sprintf("%d", record.ID))

		// Check if file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			fmt.Printf("Warning: File with ID %d not found in storage: %s\n", record.ID, record.Path)
		} else {
			// Delete the file/directory
			if err := os.RemoveAll(filePath); err != nil {
				fmt.Printf("Failed to delete file from storage: %s (ID: %d): %v\n", record.Path, record.ID, err)
				failedPaths = append(failedPaths, record.Path)
				continue
			}
		}

		deletedPaths = append(deletedPaths, record.Path)
		deletedIDs = append(deletedIDs, record.ID)
	}

	// Delete records from database
	if len(deletedIDs) > 0 {
		if err := db.DeleteFilesByIDs(database, deletedIDs); err != nil {
			fmt.Printf("Failed to delete records from database: %v\n", err)
			return
		}
	}

	// Report results
	if len(deletedPaths) > 0 {
		fmt.Printf("Successfully deleted %d file(s) from tracking:\n", len(deletedPaths))
		for _, path := range deletedPaths {
			fmt.Printf("  ✓ %s\n", path)
		}
	}

	if len(failedPaths) > 0 {
		fmt.Printf("Failed to delete %d file(s):\n", len(failedPaths))
		for _, path := range failedPaths {
			fmt.Printf("  ✗ %s\n", path)
		}
	}
}
