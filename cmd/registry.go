package cmd

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/alexcatanzaro/sk/internal/config"
	"github.com/alexcatanzaro/sk/internal/gitutil"
)

var registryCmd = &cobra.Command{
	Use:   "registry",
	Short: "Manage skill registries",
	Long:  "Add, list, refresh, and remove remote git repositories as skill registries.",
}

// ── registry add ─────────────────────────────────────────────────────────────

var (
	registryAddName string
	registryAddPath string
)

var registryAddCmd = &cobra.Command{
	Use:   "add <url>",
	Short: "Register a remote git repository as a skill registry",
	Long: `Register a remote git repository as a named skill registry.

The registry name defaults to the last path segment of the URL.
Use --name to override. Use --path to specify the skills subpath inside
the repository (default: skills).`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		rawURL := args[0]

		// Derive default name from URL tail.
		name := registryAddName
		if name == "" {
			u, err := url.Parse(rawURL)
			if err != nil {
				return fmt.Errorf("invalid URL %q: %w", rawURL, err)
			}
			tail := filepath.Base(u.Path)
			tail = strings.TrimSuffix(tail, ".git")
			if tail == "" || tail == "." {
				return fmt.Errorf("cannot derive registry name from URL %q; use --name", rawURL)
			}
			name = tail
		}

		subpath := registryAddPath
		if subpath == "" {
			subpath = "skills"
		}

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if cfg.FindRegistry(name) != nil {
			return fmt.Errorf("registry %q already exists\n  → use a different name with --name, or remove the existing one with: sk registry remove %s", name, name)
		}

		// Prepare cache directory.
		cacheDir := config.RegistryCacheDir(name)
		if err := os.MkdirAll(cacheDir, 0o755); err != nil {
			return fmt.Errorf("creating cache dir: %w", err)
		}

		fmt.Printf("Fetching registry %q from %s (path: %s)…\n", name, rawURL, subpath)
		if err := gitutil.InitSparseCheckout(cacheDir, rawURL, subpath); err != nil {
			_ = os.RemoveAll(cacheDir)
			return err
		}

		cfg.Registries = append(cfg.Registries, config.Registry{
			Name:          name,
			URL:           rawURL,
			Path:          subpath,
			LastRefreshed: time.Now().UTC(),
		})
		if err := cfg.Save(); err != nil {
			return err
		}
		fmt.Printf("Registry %q added.\n", name)
		return nil
	},
}

// ── registry list ─────────────────────────────────────────────────────────────

var registryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List registered skill registries",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if len(cfg.Registries) == 0 {
			fmt.Println("No registries configured.")
			fmt.Println("  → run: sk registry add <url>")
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tURL\tPATH\tLAST REFRESHED")
		for _, r := range cfg.Registries {
			refreshed := "never"
			if !r.LastRefreshed.IsZero() {
				refreshed = r.LastRefreshed.Local().Format("2006-01-02 15:04:05")
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", r.Name, r.URL, r.Path, refreshed)
		}
		return w.Flush()
	},
}

// ── registry refresh ──────────────────────────────────────────────────────────

var registryRefreshCmd = &cobra.Command{
	Use:   "refresh [<name>]",
	Short: "Re-fetch registry content from remote",
	Long: `Re-fetch the skill index from a registry's remote.

If <name> is given, only that registry is refreshed.
If omitted, all registered registries are refreshed.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		var targets []config.Registry
		if len(args) == 1 {
			r := cfg.FindRegistry(args[0])
			if r == nil {
				return fmt.Errorf("registry %q not found\n  → run: sk registry list", args[0])
			}
			targets = []config.Registry{*r}
		} else {
			targets = cfg.Registries
		}

		for i, r := range targets {
			fmt.Printf("Refreshing %q…\n", r.Name)
			cacheDir := config.RegistryCacheDir(r.Name)
			if err := gitutil.Refresh(cacheDir); err != nil {
				return err
			}
			// Update last-refreshed in config.
			for j := range cfg.Registries {
				if cfg.Registries[j].Name == targets[i].Name {
					cfg.Registries[j].LastRefreshed = time.Now().UTC()
				}
			}
		}
		if err := cfg.Save(); err != nil {
			return err
		}
		fmt.Println("Done.")
		return nil
	},
}

// ── registry remove ───────────────────────────────────────────────────────────

var registryRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a registered registry",
	Long: `Remove a registry from config and delete its local cache.

Skills installed from this registry will have broken symlinks in the store
until the registry is re-added and refreshed.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if cfg.FindRegistry(name) == nil {
			return fmt.Errorf("registry %q not found\n  → run: sk registry list", name)
		}

		// Warn about installed skills that came from this registry.
		var affected []string
		for _, s := range cfg.InstalledSkills {
			if s.Registry == name {
				affected = append(affected, s.Name)
			}
		}
		if len(affected) > 0 {
			fmt.Fprintf(os.Stderr, "warning: the following installed skills will have broken symlinks after removal: %s\n", strings.Join(affected, ", "))
			fmt.Fprintln(os.Stderr, "  → re-add the registry and run: sk registry refresh")
		}

		cfg.RemoveRegistry(name)
		if err := cfg.Save(); err != nil {
			return err
		}

		cacheDir := config.RegistryCacheDir(name)
		if err := os.RemoveAll(cacheDir); err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not delete cache dir %s: %v\n", cacheDir, err)
		}

		fmt.Printf("Registry %q removed.\n", name)
		return nil
	},
}

func init() {
	// registry add flags
	registryAddCmd.Flags().StringVar(&registryAddName, "name", "", "Override the registry name (default: URL tail)")
	registryAddCmd.Flags().StringVar(&registryAddPath, "path", "", "Subpath within the repo where skills live (default: skills)")

	registryCmd.AddCommand(registryAddCmd)
	registryCmd.AddCommand(registryListCmd)
	registryCmd.AddCommand(registryRefreshCmd)
	registryCmd.AddCommand(registryRemoveCmd)
}
