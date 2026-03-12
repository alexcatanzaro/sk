package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/alexcatanzaro/sk/internal/adapter"
	"github.com/alexcatanzaro/sk/internal/config"
	"github.com/alexcatanzaro/sk/internal/syncer"
)

var backendCmd = &cobra.Command{
	Use:   "backend",
	Short: "Manage tool backends for skill sync",
	Long:  "List, enable, and disable agentskills.io-compliant tool backends.",
}

// ── backend list ──────────────────────────────────────────────────────────────

var backendListCmd = &cobra.Command{
	Use:   "list",
	Short: "List known backends and their status",
	Long:  "Show all built-in adapters with their availability, enabled state, and skills path.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tAVAILABLE\tENABLED\tSKILLS PATH")

		adapters := adapter.KnownAdapters
		sort.Slice(adapters, func(i, j int) bool { return adapters[i].Name() < adapters[j].Name() })

		for _, a := range adapters {
			avail := "no"
			if a.IsAvailable() {
				avail = "yes"
			}
			enabled := "no"
			if cfg.IsBackendEnabled(a.Name()) {
				enabled = "yes"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", a.Name(), avail, enabled, a.SkillsPath())
		}
		return w.Flush()
	},
}

// ── backend enable ────────────────────────────────────────────────────────────

var backendEnableCmd = &cobra.Command{
	Use:   "enable <name>",
	Short: "Enable a backend and sync installed skills to it",
	Long: `Enable a backend adapter by name.

The tool must be detected (~/<dirname>/ must exist) before it can be enabled.
After enabling, installed skills are immediately synced to the backend.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		a := adapter.Find(name)
		if a == nil {
			var names []string
			for _, k := range adapter.KnownAdapters {
				names = append(names, k.Name())
			}
			sort.Strings(names)
			return fmt.Errorf("unknown backend %q\n  → valid backends: %s", name, strings.Join(names, ", "))
		}

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if cfg.IsBackendEnabled(name) {
			fmt.Printf("Backend %q is already enabled.\n", name)
			return nil
		}

		if !a.IsAvailable() {
			toolDir := filepath.Dir(a.SkillsPath())
			return fmt.Errorf("%s not detected (%s not found)\n  → install %s and retry", name, toolDir, name)
		}

		cfg.EnabledBackends = append(cfg.EnabledBackends, name)
		if err := cfg.Save(); err != nil {
			return err
		}
		fmt.Printf("Backend %q enabled.\n", name)

		// Sync immediately.
		fmt.Println("Syncing installed skills…")
		if err := syncer.Run(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "warning: sync failed: %v\n", err)
		}
		return nil
	},
}

// ── backend disable ───────────────────────────────────────────────────────────

var backendDisableCmd = &cobra.Command{
	Use:   "disable <name>",
	Short: "Disable a backend (skills remain in the tool's directory)",
	Long: `Remove a backend from the enabled list.

Existing skill files in the tool's skills directory are NOT deleted.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if !cfg.IsBackendEnabled(name) {
			return fmt.Errorf("backend %q is not enabled\n  → run: sk backend list", name)
		}

		cfg.RemoveBackend(name)
		if err := cfg.Save(); err != nil {
			return err
		}
		fmt.Printf("Backend %q disabled. Existing files in its skills directory are untouched.\n", name)
		return nil
	},
}

func init() {
	backendCmd.AddCommand(backendListCmd)
	backendCmd.AddCommand(backendEnableCmd)
	backendCmd.AddCommand(backendDisableCmd)
}
