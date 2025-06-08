package internal

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/tylerkeyes/dot-sync/internal/db"
	"github.com/tylerkeyes/dot-sync/internal/shared"
	"github.com/tylerkeyes/dot-sync/internal/storage"
)

func NewSyncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Sync dotfiles to remote storage",
		Run:   syncHandler,
	}
}

func syncHandler(cmd *cobra.Command, args []string) {
	dotSyncFilesPath := filepath.Join(shared.FindHomeDir(), shared.GetDotSyncFilesDir())

	database, err := db.OpenDotSyncDB()
	if err != nil {
		fmt.Println("Failed to open .dot-sync.db:", err)
		return
	}
	defer database.Close()

	records, err := db.GetAllFilePaths(database)
	if err != nil {
		fmt.Println("Failed to read file paths from database:", err)
		return
	}

	for _, rec := range records {
		if err := shared.CopyToDotSyncFilesByID(rec.ID, rec.Path, dotSyncFilesPath); err != nil {
			fmt.Printf("Failed to copy %s: %v\n", rec.Path, err)
		}
	}

	ctx := cmd.Context()
	sp := ctx.Value(shared.GetStorageProviderKey()).(storage.StorageProvider)
	if err := sp.PushToStorage(dotSyncFilesPath); err != nil {
		fmt.Println("Failed to push to storage:", err)
		return
	}

	fmt.Println("Sync complete.")
}
