package storage

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

type StorageProvider interface {
	InitializeStorage() error
	PushToStorage(filePath string) error
	PullFromStorage(filePath string) error
}

func NewStorageProviderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "storage",
		Short: "Manage dotfile storage backends",
	}
	cmd.AddCommand(newInitCmd())
	return cmd
}

func newInitCmd() *cobra.Command {
	var provider string
	var remoteURL string

	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize backend storage provider",
		RunE: func(cmd *cobra.Command, args []string) error {
			var sp StorageProvider
			switch provider {
			case "git":
				if remoteURL == "" {
					return errors.New("--remote-url is required for git provider")
				}
				sp = &GitStorage{RemoteURL: remoteURL}
			default:
				return fmt.Errorf("unsupported storage provider: %s", provider)
			}
			if err := sp.InitializeStorage(); err != nil {
				return fmt.Errorf("failed to initialize storage: %w", err)
			}
			fmt.Println("Storage initialized successfully.")
			return nil
		},
	}
	initCmd.Flags().StringVar(&provider, "provider", "git", "Storage provider to use (git)")
	initCmd.Flags().StringVar(&remoteURL, "remote-url", "", "Remote URL for git storage provider")
	initCmd.MarkFlagRequired("provider")
	return initCmd
}
