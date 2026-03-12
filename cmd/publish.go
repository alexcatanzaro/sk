package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/alexcatanzaro/sk/internal/config"
	"github.com/alexcatanzaro/sk/internal/gitutil"
	"github.com/alexcatanzaro/sk/internal/skill"
)

var publishTo string

var publishCmd = &cobra.Command{
	Use:   "publish <name>",
	Short: "Publish an authored skill to a remote registry",
	Long: `Validate and publish an authored skill to a remote registry via git.

The skill must be in ~/.local/share/sk/authored/<name>/.
Your existing git credentials (SSH or HTTPS) are used for the push.

If only one registry is configured, --to is optional.
With multiple registries, --to <registry> is required.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		// Resolve target registry.
		var target *config.Registry
		if publishTo != "" {
			target = cfg.FindRegistry(publishTo)
			if target == nil {
				return fmt.Errorf("registry %q not found\n  → run: sk registry list", publishTo)
			}
		} else {
			switch len(cfg.Registries) {
			case 0:
				return fmt.Errorf("no registries configured\n  → run: sk registry add <url>")
			case 1:
				target = &cfg.Registries[0]
			default:
				return fmt.Errorf("multiple registries configured; specify one with --to <registry>")
			}
		}

		// Validate the skill.
		authoredDir := filepath.Join(config.AuthoredDir(), name)
		if _, err := os.Stat(authoredDir); os.IsNotExist(err) {
			return fmt.Errorf("skill %q not found in authored dir (%s)\n  → scaffold it with: sk skill new %s", name, authoredDir, name)
		}
		if err := skill.Validate(authoredDir); err != nil {
			return fmt.Errorf("skill validation failed: %w", err)
		}

		// Copy skill into registry sparse checkout.
		cacheDir := config.RegistryCacheDir(target.Name)
		destSkillsDir := filepath.Join(cacheDir, target.Path)
		dest := filepath.Join(destSkillsDir, name)

		if err := copyDirPublish(authoredDir, dest); err != nil {
			return fmt.Errorf("copying skill to registry cache: %w", err)
		}

		// git add + commit + push.
		if err := gitutil.RunIn(cacheDir, "add", filepath.Join(target.Path, name)); err != nil {
			return err
		}
		msg := fmt.Sprintf("publish skill: %s", name)
		if err := gitutil.RunIn(cacheDir, "commit", "-m", msg); err != nil {
			return err
		}
		if err := gitutil.RunIn(cacheDir, "push", "origin"); err != nil {
			// Undo the commit to leave cache in pre-publish state.
			_ = gitutil.RunIn(cacheDir, "reset", "--soft", "HEAD~1")
			return err
		}

		fmt.Printf("Skill %q published to registry %q.\n", name, target.Name)
		return nil
	},
}

// copyDirPublish recursively copies src to dst for the publish operation.
func copyDirPublish(src, dst string) error {
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		in, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		info, _ := d.Info()
		return os.WriteFile(target, in, info.Mode())
	})
}

func init() {
	publishCmd.Flags().StringVar(&publishTo, "to", "", "Target registry name (required when multiple registries are configured)")
}
