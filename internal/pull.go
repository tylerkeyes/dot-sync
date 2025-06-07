package internal

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewPullCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "pull",
		Short: "Pull dotfiles from remote storage",
		Run:   pullHandler,
	}
}

func pullHandler(cmd *cobra.Command, args []string) {
	// ctx := cmd.Context() // Only use if you need the storage provider
	fmt.Println("Pulling dotfiles...")
}
