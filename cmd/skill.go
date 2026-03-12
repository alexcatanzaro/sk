package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/alexcatanzaro/sk/internal/config"
	"github.com/alexcatanzaro/sk/internal/registry"
	"github.com/alexcatanzaro/sk/internal/skill"
	"github.com/alexcatanzaro/sk/internal/syncer"
)

var skillCmd = &cobra.Command{
	Use:   "skill",
	Short: "Browse, install, remove, and author skills",
}

// ── skill list (TUI) ──────────────────────────────────────────────────────────

var skillListFrom string

// skillItem implements list.Item for bubbletea.
type skillItem struct {
	info skill.Info
}

func (s skillItem) Title() string       { return s.info.Name }
func (s skillItem) Description() string { return fmt.Sprintf("[%s] %s", s.info.Registry, s.info.Description) }
func (s skillItem) FilterValue() string { return s.info.Name + " " + s.info.Description }

var titleStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("62")).
	Bold(true).
	Padding(0, 1)

var skillListCmd = &cobra.Command{
	Use:   "list",
	Short: "Browse available skills from registered registries",
	Long: `Open a TUI browser showing skills available from registered registries.

Use arrow keys to navigate, type to filter. Press q or Ctrl+C to quit.
Use --from to limit to a specific registry.`,
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

		skills, err := buildCatalog(cfg, skillListFrom)
		if err != nil {
			return err
		}
		if len(skills) == 0 {
			if skillListFrom != "" {
				fmt.Printf("No skills found in registry %q.\n", skillListFrom)
			} else {
				fmt.Println("No skills found in any registered registry.")
			}
			return nil
		}

		items := make([]list.Item, len(skills))
		for i, s := range skills {
			items[i] = skillItem{info: s}
		}

		l := list.New(items, list.NewDefaultDelegate(), 80, 24)
		l.Title = "sk skill list"
		l.Styles.Title = titleStyle

		m := skillListModel{list: l}
		p := tea.NewProgram(m, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("TUI error: %w", err)
		}
		return nil
	},
}

// skillListModel is the bubbletea model for the skill list TUI.
type skillListModel struct {
	list list.Model
}

func (m skillListModel) Init() tea.Cmd { return nil }

func (m skillListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height-2)
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m skillListModel) View() string {
	return m.list.View()
}

// ── skill add ─────────────────────────────────────────────────────────────────

var skillAddFrom string

var skillAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Install a skill from a registered registry",
	Long: `Install a skill by name into the canonical store (~/.local/share/sk/installed/).

If multiple registries contain a skill with the same name, you must specify
--from <registry> to disambiguate.

After installation, installed skills are automatically synced to all enabled backends.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		// Already installed?
		if cfg.FindInstalledSkill(name) != nil {
			fmt.Printf("Skill %q is already installed.\n", name)
			return nil
		}

		skills, err := buildCatalog(cfg, skillAddFrom)
		if err != nil {
			return err
		}

		// Find matching skills.
		var matches []skill.Info
		for _, s := range skills {
			if s.Name == name {
				matches = append(matches, s)
			}
		}

		if len(matches) == 0 {
			return fmt.Errorf("skill %q not found in any registered registry\n  → run: sk skill list", name)
		}

		if len(matches) > 1 && skillAddFrom == "" {
			var registries []string
			for _, m := range matches {
				registries = append(registries, m.Registry)
			}
			return fmt.Errorf("skill %q exists in multiple registries: %s\n  → use --from <registry> to specify which one", name, strings.Join(registries, ", "))
		}

		chosen := matches[0]

		// Create installed/ dir and symlink.
		installedDir := config.InstalledDir()
		if err := os.MkdirAll(installedDir, 0o755); err != nil {
			return fmt.Errorf("creating installed dir: %w", err)
		}
		link := filepath.Join(installedDir, name)
		if err := os.Symlink(chosen.Dir, link); err != nil {
			return fmt.Errorf("creating symlink: %w", err)
		}

		cfg.InstalledSkills = append(cfg.InstalledSkills, config.InstalledSkill{
			Name:     name,
			Registry: chosen.Registry,
		})
		if err := cfg.Save(); err != nil {
			return err
		}
		fmt.Printf("Skill %q installed from %q.\n", name, chosen.Registry)

		// Trigger sync automatically.
		if len(cfg.EnabledBackends) > 0 {
			fmt.Println("Syncing to enabled backends…")
			if err := syncer.Run(cfg); err != nil {
				fmt.Fprintf(os.Stderr, "warning: sync failed: %v\n", err)
			}
		}
		return nil
	},
}

// ── skill remove ──────────────────────────────────────────────────────────────

var skillRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove an installed skill",
	Long: `Remove an installed skill from the canonical store.

This deletes the symlink from installed/ and removes the record from config.
The skill is NOT deleted from the registry cache.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if cfg.FindInstalledSkill(name) == nil {
			return fmt.Errorf("skill %q is not installed\n  → run: sk skill list --installed (or check ~/.local/share/sk/installed/)", name)
		}

		link := filepath.Join(config.InstalledDir(), name)
		if err := os.Remove(link); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("removing symlink: %w", err)
		}

		cfg.RemoveInstalledSkill(name)
		if err := cfg.Save(); err != nil {
			return err
		}
		fmt.Printf("Skill %q removed.\n", name)
		return nil
	},
}

// ── skill new ─────────────────────────────────────────────────────────────────

var skillNewCmd = &cobra.Command{
	Use:   "new <name>",
	Short: "Scaffold a new skill for authoring",
	Long: `Scaffold a new skill directory at ~/.local/share/sk/authored/<name>/.

The name must be lowercase alphanumeric with hyphens, 1-64 characters,
no leading/trailing/consecutive hyphens.

After scaffolding, edit the SKILL.md and publish with: sk publish <name> --to <registry>`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if err := skill.ValidateName(name); err != nil {
			return err
		}

		authoredDir := filepath.Join(config.AuthoredDir(), name)
		if _, err := os.Stat(authoredDir); err == nil {
			return fmt.Errorf("skill %q already exists at %s", name, authoredDir)
		}

		if err := os.MkdirAll(authoredDir, 0o755); err != nil {
			return fmt.Errorf("creating skill dir: %w", err)
		}

		skillFile := filepath.Join(authoredDir, "SKILL.md")
		if err := os.WriteFile(skillFile, []byte(skill.SkillTemplate(name)), 0o644); err != nil {
			return fmt.Errorf("writing SKILL.md: %w", err)
		}

		fmt.Printf("Skill %q scaffolded at:\n  %s\n\nEdit the file, then publish with:\n  sk publish %s --to <registry>\n", name, skillFile, name)
		return nil
	},
}

// ── catalog builder ───────────────────────────────────────────────────────────

// buildCatalog aggregates skills from all registries (or just fromRegistry).
func buildCatalog(cfg *config.Config, fromRegistry string) ([]skill.Info, error) {
	var registries []config.Registry
	if fromRegistry != "" {
		r := cfg.FindRegistry(fromRegistry)
		if r == nil {
			return nil, fmt.Errorf("registry %q not found\n  → run: sk registry list", fromRegistry)
		}
		registries = []config.Registry{*r}
	} else {
		registries = cfg.Registries
	}

	var all []skill.Info
	for _, r := range registries {
		cacheDir := config.RegistryCacheDir(r.Name)
		skills, err := registry.WalkSkills(cacheDir, r.Name, r.Path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not read registry %q cache: %v\n", r.Name, err)
			continue
		}
		all = append(all, skills...)
	}
	return all, nil
}

func init() {
	skillListCmd.Flags().StringVar(&skillListFrom, "from", "", "Limit to a specific registry")
	skillAddCmd.Flags().StringVar(&skillAddFrom, "from", "", "Specify which registry to install from (required when skill exists in multiple)")

	skillCmd.AddCommand(skillListCmd)
	skillCmd.AddCommand(skillAddCmd)
	skillCmd.AddCommand(skillRemoveCmd)
	skillCmd.AddCommand(skillNewCmd)
}
