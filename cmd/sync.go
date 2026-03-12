package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/alexcatanzaro/sk/internal/config"
	"github.com/alexcatanzaro/sk/internal/syncer"
)

// promptSyncBase asks the user where skills should be synced and returns the
// chosen base directory. Empty input or "1" → cwd; "2" → home dir.
func promptSyncBase() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getting current directory: %w", err)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home directory: %w", err)
	}

	fmt.Printf("\nWhere should skills be synced?\n")
	fmt.Printf("  [1] Project directory (%s)  (default)\n", cwd)
	fmt.Printf("  [2] Home directory (~)\n")
	fmt.Printf("\nEnter choice [1]: ")

	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("reading input: %w", err)
	}

	switch strings.TrimSpace(line) {
	case "", "1":
		return cwd, nil
	case "2":
		return home, nil
	default:
		return cwd, nil
	}
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync installed skills to all enabled backends",
	Long: `Copy all installed skills from ~/.local/share/sk/installed/ into
each enabled backend's skills directory.

Orphaned skills (removed from installed/) are deleted from backend paths.
Broken symlinks in installed/ are warned about and skipped.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		baseDir, err := promptSyncBase()
		if err != nil {
			return err
		}
		if err := syncer.Run(cfg, baseDir); err != nil {
			return err
		}
		fmt.Println("Sync complete.")
		return nil
	},
}
