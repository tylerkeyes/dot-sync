package internal

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewPullCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "pull",
		Short: "Pull dotfiles from remote storage",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Pulling dotfiles...")
		},
	}
}
