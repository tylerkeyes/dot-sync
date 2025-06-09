package internal

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tylerkeyes/dot-sync/internal/db"
)

func NewShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show the paths of all files currently tracked for syncing",
		Run:   showHandler,
	}
}

func showHandler(cmd *cobra.Command, args []string) {
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

	records, err := db.GetAllFilePaths(database)
	if err != nil {
		fmt.Println("Failed to retrieve file paths:", err)
		return
	}

	if len(records) == 0 {
		fmt.Println("No files currently tracked for syncing.")
		return
	}

	fmt.Printf("Files currently tracked for syncing (%d):\n", len(records))
	for _, record := range records {
		fmt.Printf("  %s\n", record.Path)
	}
}
