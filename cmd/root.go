package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tylerkeyes/dot-sync/internal"
)

var rootCmd = &cobra.Command{
	Use:   "dot-sync",
	Short: "A CLI tool for dotfile syncing",
	Long:  `dot-sync is a CLI tool for managing and syncing dotfiles.`,
}

func init() {
	rootCmd.AddCommand(internal.NewSyncCmd())
	rootCmd.AddCommand(internal.NewPullCmd())
	rootCmd.AddCommand(internal.NewMarkCmd())
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}

func main() {
	Execute()
}
