package syncer

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/alexcatanzaro/sk/internal/adapter"
	"github.com/alexcatanzaro/sk/internal/config"
)

// Run copies all installed skills from the canonical store into every enabled
// backend's skills path, and removes orphaned skills from backend paths.
//
// A broken symlink in installed/ is warned about and skipped.
// Returns an error only if no backends are enabled or a copy fails.
func Run(cfg *config.Config) error {
	if len(cfg.EnabledBackends) == 0 {
		return fmt.Errorf("no backends enabled\n  → run: sk backend enable <name>")
	}

	installedDir := config.InstalledDir()

	// Collect names of currently installed skills.
	installedNames := map[string]bool{}
	entries, err := os.ReadDir(installedDir)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("reading installed dir: %w", err)
	}

	for _, e := range entries {
		target, err := filepath.EvalSymlinks(filepath.Join(installedDir, e.Name()))
		if err != nil {
			// broken symlink
			fmt.Fprintf(os.Stderr, "warning: installed skill %q has a broken symlink (registry cache may have been cleared) — skipping\n", e.Name())
			continue
		}
		installedNames[e.Name()] = true
		// Copy to each enabled backend.
		for _, backendName := range cfg.EnabledBackends {
			a := adapter.Find(backendName)
			if a == nil {
				fmt.Fprintf(os.Stderr, "warning: unknown backend %q in config — skipping\n", backendName)
				continue
			}
			dest := filepath.Join(a.SkillsPath(), e.Name())
			if err := copyDir(target, dest); err != nil {
				return fmt.Errorf("syncing skill %q to backend %q: %w", e.Name(), backendName, err)
			}
		}
	}

	// Orphan cleanup: remove skills from backend paths no longer in installed/.
	for _, backendName := range cfg.EnabledBackends {
		a := adapter.Find(backendName)
		if a == nil {
			continue
		}
		backendSkillsDir := a.SkillsPath()
		bEntries, err := os.ReadDir(backendSkillsDir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("reading backend skills dir %s: %w", backendSkillsDir, err)
		}
		for _, be := range bEntries {
			if !installedNames[be.Name()] {
				orphan := filepath.Join(backendSkillsDir, be.Name())
				if err := os.RemoveAll(orphan); err != nil {
					fmt.Fprintf(os.Stderr, "warning: could not remove orphan %s: %v\n", orphan, err)
				}
			}
		}
	}
	return nil
}

// copyDir recursively copies src directory to dst, overwriting existing files.
func copyDir(src, dst string) error {
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
		return copyFile(path, target)
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	info, err := in.Stat()
	if err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
