package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/tylerkeyes/dot-sync/internal"
	"github.com/tylerkeyes/dot-sync/internal/db"
	"github.com/tylerkeyes/dot-sync/internal/shared"
	"github.com/tylerkeyes/dot-sync/internal/storage"
)

var rootCmd = &cobra.Command{
	Use:   "dot-sync",
	Short: "A CLI tool for dotfile syncing",
	Long:  `dot-sync is a CLI tool for managing and syncing dotfiles.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip check for 'storage init' command
		if len(os.Args) > 2 && os.Args[1] == "storage" && os.Args[2] == "init" {
			return nil
		}
		// Open DB and check for storage provider
		database, err := db.OpenDotSyncDB()
		if err != nil {
			return fmt.Errorf("failed to open .dot-sync.db: %v", err)
		}
		defer database.Close()
		if err := db.EnsureStorageTable(database); err != nil {
			return fmt.Errorf("failed to ensure storage_provider table: %w", err)
		}

		row := database.QueryRow("SELECT storage_type, remote FROM storage_provider ORDER BY id DESC LIMIT 1")
		var storageType, remote string
		if err := row.Scan(&storageType, &remote); err != nil {
			if err == sql.ErrNoRows {
				fmt.Println("No storage provider configured. Please run: dot-sync storage init")
				os.Exit(1)
			}
			return fmt.Errorf("failed to read storage provider: %v", err)
		}
		// Only support git for now
		var sp storage.StorageProvider
		if storageType == "git" {
			sp = &storage.GitStorage{RemoteURL: remote}
			if err := sp.InitializeStorage(); err != nil {
				return fmt.Errorf("failed to initialize storage provider: %v", err)
			}
		} else {
			return fmt.Errorf("unsupported storage provider: %s", storageType)
		}
		ctx := context.WithValue(cmd.Context(), shared.GetStorageProviderKey(), sp)
		cmd.SetContext(ctx)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(internal.NewSyncCmd())
	rootCmd.AddCommand(internal.NewPullCmd())
	rootCmd.AddCommand(internal.NewMarkCmd())
	rootCmd.AddCommand(internal.NewShowCmd())
	rootCmd.AddCommand(internal.NewDeleteCmd())
	rootCmd.AddCommand(storage.NewStorageProviderCmd())

	home := shared.FindHomeDir()
	if home == "" {
		fmt.Println("Could not determine home directory")
		return
	}
	dotSyncPath := filepath.Join(home, shared.GetDotSyncDir())
	dotSyncFilesPath := filepath.Join(home, shared.GetDotSyncFilesDir())

	if err := shared.EnsureDir(dotSyncPath); err != nil {
		fmt.Println("Failed to create .dot-sync directory:", err)
		return
	}
	if err := shared.EnsureDir(dotSyncFilesPath); err != nil {
		fmt.Println("Failed to create .dot-sync/files directory:", err)
		return
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}

func main() {
	Execute()
}
