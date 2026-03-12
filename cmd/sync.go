package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/alexcatanzaro/sk/internal/config"
	"github.com/alexcatanzaro/sk/internal/syncer"
)

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
		if err := syncer.Run(cfg); err != nil {
			return err
		}
		fmt.Println("Sync complete.")
		return nil
	},
}
