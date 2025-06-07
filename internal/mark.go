package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tylerkeyes/dot-sync/internal/shared"
)

func NewMarkCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "mark [files or directories]...",
		Short: "Mark a file or directory for syncing",
		Args:  cobra.MinimumNArgs(0),
		Run:   markHandler,
	}
}

func markHandler(cmd *cobra.Command, args []string) {
	// ctx := cmd.Context() // Only use if you need the storage provider
	db, err := OpenDotSyncDB()
	if err != nil {
		fmt.Println("Failed to open .dot-sync.db:", err)
		return
	}
	defer db.Close()

	if err := EnsureFilesTable(db); err != nil {
		fmt.Println("Failed to ensure files table:", err)
		return
	}

	if len(args) == 0 {
		fmt.Println("No changes.")
		return
	}

	// Add new entries from args
	absPaths := argsAsFullPaths(args)
	if err := InsertFiles(db, absPaths); err != nil {
		fmt.Printf("Failed to mark entries: %v\n", err)
	}
	fmt.Println("Marked entries for syncing:", absPaths)
}

func argsAsFullPaths(args []string) []string {
	cwd := ""
	if wd, err := os.Getwd(); err == nil {
		cwd = wd
	}
	var absPaths []string
	for _, input := range args {
		var absPath string
		if len(input) > 0 && input[0] == os.PathSeparator {
			// Absolute path
			absPath = input
		} else {
			// Relative path or single name
			baseDir := cwd
			if baseDir == "" {
				baseDir = shared.FindHomeDir()
			}
			if strings.Contains(input, string(os.PathSeparator)) {
				// Relative path: resolve to absolute path if possible
				if abs, err := filepath.Abs(input); err == nil {
					absPath = abs
				} else {
					absPath = baseDir + string(os.PathSeparator) + input
				}
			} else {
				// Single name: use baseDir + input
				absPath = baseDir + string(os.PathSeparator) + input
			}
		}
		absPaths = append(absPaths, absPath)
	}
	return absPaths
}
