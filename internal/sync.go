package internal

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewSyncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Sync dotfiles to remote storage",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Syncing dotfiles...")
		},
	}
}
