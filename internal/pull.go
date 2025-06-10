package internal

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/tylerkeyes/dot-sync/internal/db"
	"github.com/tylerkeyes/dot-sync/internal/shared"
	"github.com/tylerkeyes/dot-sync/internal/storage"
)

func NewPullCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "pull",
		Short: "Pull dotfiles from remote storage",
		Run:   pullHandler,
	}
}

func pullHandler(cmd *cobra.Command, args []string) {
	fmt.Println("Pulling dotfiles...")

	// Get the .dot-sync/files directory path
	dotSyncFilesPath := filepath.Join(shared.FindHomeDir(), shared.GetDotSyncFilesDir())
	dotSyncDir := filepath.Join(shared.FindHomeDir(), shared.GetDotSyncDir())

	// Ensure the .dot-sync/files directory exists
	if err := shared.EnsureDir(dotSyncFilesPath); err != nil {
		fmt.Printf("Failed to create .dot-sync/files directory: %v\n", err)
		return
	}

	// Get storage provider from context
	ctx := cmd.Context()
	sp, ok := ctx.Value(shared.GetStorageProviderKey()).(storage.StorageProvider)
	if !ok || sp == nil {
		fmt.Println("No storage provider configured. Please run 'dot-sync storage init' first.")
		return
	}

	// Pull from remote storage
	if err := sp.PullFromStorage(dotSyncDir); err != nil {
		fmt.Println("Failed to pull from storage:", err)
		return
	}

	// Open database to get file mappings
	database, err := db.OpenDotSyncDB()
	if err != nil {
		fmt.Println("Failed to open .dot-sync.db:", err)
		return
	}
	defer database.Close()

	// Get all file path records
	records, err := db.GetAllFilePaths(database)
	if err != nil {
		fmt.Println("Failed to read file paths from database:", err)
		return
	}

	if len(records) == 0 {
		fmt.Println("No files found in database. Nothing to pull.")
		return
	}

	// Copy files from .dot-sync/files/{id} back to their original locations
	for _, rec := range records {
		srcPath := filepath.Join(dotSyncFilesPath, fmt.Sprintf("%d", rec.ID))
		dstPath := rec.Path

		// Check if source file exists in .dot-sync/files
		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			fmt.Printf("Warning: File with ID %d not found in storage, skipping %s\n", rec.ID, rec.Path)
			continue
		}

		// Ensure destination directory exists
		if err := shared.EnsureDir(filepath.Dir(dstPath)); err != nil {
			fmt.Printf("Failed to create directory for %s: %v\n", dstPath, err)
			continue
		}

		// Determine if it's a file or directory and copy accordingly
		srcInfo, err := os.Lstat(srcPath)
		if err != nil {
			fmt.Printf("Failed to get info for %s: %v\n", srcPath, err)
			continue
		}

		if srcInfo.IsDir() {
			if err := shared.CopyDir(srcPath, dstPath); err != nil {
				fmt.Printf("Failed to copy directory %s to %s: %v\n", srcPath, dstPath, err)
				continue
			}
		} else {
			if err := shared.CopyFile(srcPath, dstPath); err != nil {
				fmt.Printf("Failed to copy file %s to %s: %v\n", srcPath, dstPath, err)
				continue
			}
		}

		fmt.Printf("✓ Restored: %s\n", rec.Path)
	}

	fmt.Println("Pull complete.")
}
