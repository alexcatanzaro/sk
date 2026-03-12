package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

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

// ── Types and constants ────────────────────────────────────────────────────────

type viewState int

const (
	stateRegPicker viewState = iota
	stateLoading
	stateSkillList
)

type skillsLoadedMsg struct {
	skills   []skill.Info
	registry string
	err      error
}

type skillInstalledMsg struct {
	name     string
	registry string
	backends int
	err      error
}

type tickMsg struct{}

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// ── Styles ────────────────────────────────────────────────────────────────────

var titleStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("62")).
	Bold(true).
	Padding(0, 1)

var (
	checkStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))  // green ✓
	spinnerStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("220")) // yellow spinner
	selectedTitle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	normalTitle   = lipgloss.NewStyle()
	itemDesc      = lipgloss.NewStyle().Foreground(lipgloss.Color("241")) // dim gray
	successStatus = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))  // green
	errorStatus   = lipgloss.NewStyle().Foreground(lipgloss.Color("196")) // red
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241")) // dim gray
)

// ── List item types ───────────────────────────────────────────────────────────

type skillItem struct{ info skill.Info }

func (s skillItem) Title() string       { return s.info.Name }
func (s skillItem) Description() string { return s.info.Description }
func (s skillItem) FilterValue() string { return s.info.Name + " " + s.info.Description }

type registryItem struct{ name string }

func (r registryItem) Title() string       { return r.name }
func (r registryItem) Description() string { return "" }
func (r registryItem) FilterValue() string { return r.name }

// ── Custom skill delegate ─────────────────────────────────────────────────────

// skillDelegate renders each skill with a ✓ (installed), spinner (installing),
// or blank glyph. installed and installing are maps shared with the model —
// since maps are reference types, mutations in the model are visible here
// without rebuilding the delegate. spinnerFrame must be updated via SetDelegate.
type skillDelegate struct {
	installed    map[string]bool
	installing   map[string]bool
	spinnerFrame int
}

func (d skillDelegate) Height() int                               { return 2 }
func (d skillDelegate) Spacing() int                              { return 1 }
func (d skillDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd  { return nil }

func (d skillDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	s, ok := item.(skillItem)
	if !ok {
		return
	}

	var glyph string
	switch {
	case d.installed[s.info.Name]:
		glyph = checkStyle.Render("✓")
	case d.installing[s.info.Name]:
		frame := spinnerFrames[d.spinnerFrame%len(spinnerFrames)]
		glyph = spinnerStyle.Render(frame)
	default:
		glyph = " "
	}

	var titleStr string
	if index == m.Index() {
		titleStr = selectedTitle.Render(s.info.Name)
	} else {
		titleStr = normalTitle.Render(s.info.Name)
	}
	descStr := itemDesc.Render(s.info.Description)

	fmt.Fprintf(w, "%s %s\n  %s", glyph, titleStr, descStr)
}

// ── Model ─────────────────────────────────────────────────────────────────────

type skillListModel struct {
	state            viewState
	regList          list.Model
	skillList        list.Model
	cfg              *config.Config
	installed        map[string]bool
	installing       map[string]bool
	spinnerFrame     int
	status           string
	statusErr        bool
	multiRegistry    bool
	selectedRegistry string
	width            int
	height           int
}

func newSkillListModel(cfg *config.Config, fromRegistry string) skillListModel {
	installed := make(map[string]bool)
	for _, s := range cfg.InstalledSkills {
		installed[s.Name] = true
	}

	m := skillListModel{
		cfg:           cfg,
		installed:     installed,
		installing:    make(map[string]bool),
		multiRegistry: fromRegistry == "" && len(cfg.Registries) > 1,
	}

	if m.multiRegistry {
		// Tasks 4.2: build registry picker list
		items := make([]list.Item, len(cfg.Registries))
		for i, r := range cfg.Registries {
			items[i] = registryItem{name: r.Name}
		}
		d := list.NewDefaultDelegate()
		d.ShowDescription = false
		m.regList = list.New(items, d, 80, 20)
		m.regList.Title = "Select a registry"
		m.regList.Styles.Title = titleStyle
		m.state = stateRegPicker
	} else {
		// Task 4.3: single registry or --from: load immediately
		if fromRegistry != "" {
			m.selectedRegistry = fromRegistry
		} else {
			m.selectedRegistry = cfg.Registries[0].Name
		}
		m.state = stateLoading
	}

	return m
}

// Task 4.4
func (m skillListModel) Init() tea.Cmd {
	if m.state == stateLoading {
		return loadSkillsCmd(m.cfg, m.selectedRegistry)
	}
	return nil
}

func (m skillListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Task 5.1: window resize
	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		m.width, m.height = ws.Width, ws.Height
		switch m.state {
		case stateRegPicker:
			m.regList.SetSize(ws.Width, ws.Height-2)
		case stateSkillList:
			m.skillList.SetSize(ws.Width, ws.Height-3)
		}
		return m, nil
	}

	// Task 5.5: skills loaded
	if msg, ok := msg.(skillsLoadedMsg); ok {
		items := make([]list.Item, len(msg.skills))
		for i, s := range msg.skills {
			items[i] = skillItem{info: s}
		}
		delegate := skillDelegate{
			installed:    m.installed,
			installing:   m.installing,
			spinnerFrame: m.spinnerFrame,
		}
		h := m.height - 3
		if h < 5 {
			h = 20
		}
		m.skillList = list.New(items, delegate, m.width, h)
		m.skillList.SetShowHelp(false) // we render custom help in viewSkillList
		if msg.err != nil {
			m.skillList.Title = "Error"
			m.status = msg.err.Error()
			m.statusErr = true
		} else {
			m.selectedRegistry = msg.registry
			m.skillList.Title = fmt.Sprintf("%s · %d skills", msg.registry, len(msg.skills))
			m.skillList.Styles.Title = titleStyle
		}
		m.state = stateSkillList
		return m, nil
	}

	// Task 5.6: install result
	if msg, ok := msg.(skillInstalledMsg); ok {
		delete(m.installing, msg.name)
		if msg.err != nil {
			m.status = fmt.Sprintf("✗ %s: %s", msg.name, msg.err)
			m.statusErr = true
		} else {
			m.installed[msg.name] = true
			suffix := ""
			switch msg.backends {
			case 1:
				suffix = " · synced to 1 backend"
			default:
				if msg.backends > 1 {
					suffix = fmt.Sprintf(" · synced to %d backends", msg.backends)
				}
			}
			m.status = fmt.Sprintf("✓ %s installed%s", msg.name, suffix)
			m.statusErr = false
		}
		// Update delegate to reflect new installed/installing state
		m.skillList.SetDelegate(skillDelegate{
			installed:    m.installed,
			installing:   m.installing,
			spinnerFrame: m.spinnerFrame,
		})
		return m, nil
	}

	// Task 5.7: spinner tick
	if _, ok := msg.(tickMsg); ok {
		m.spinnerFrame++
		if len(m.installing) > 0 {
			m.skillList.SetDelegate(skillDelegate{
				installed:    m.installed,
				installing:   m.installing,
				spinnerFrame: m.spinnerFrame,
			})
			return m, tickCmd()
		}
		return m, nil
	}

	// Task 5.2-5.4: key handling
	if msg, ok := msg.(tea.KeyMsg); ok {
		// ctrl+c always quits
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		switch m.state {
		case stateRegPicker:
			// Task 5.2
			if m.regList.FilterState() != list.Filtering {
				switch msg.String() {
				case "q":
					return m, tea.Quit
				case "enter":
					if ri, ok := m.regList.SelectedItem().(registryItem); ok {
						m.selectedRegistry = ri.name
						m.state = stateLoading
						return m, loadSkillsCmd(m.cfg, ri.name)
					}
				}
			}
			var cmd tea.Cmd
			m.regList, cmd = m.regList.Update(msg)
			return m, cmd

		case stateLoading:
			// Task 5.3: ignore all input during load
			return m, nil

		case stateSkillList:
			// Task 5.4
			if m.skillList.FilterState() != list.Filtering {
				switch msg.String() {
				case "q":
					return m, tea.Quit
				case "esc":
					if m.multiRegistry {
						m.state = stateRegPicker
						m.status = ""
						return m, nil
					}
					return m, tea.Quit
				case "enter":
					si, ok := m.skillList.SelectedItem().(skillItem)
					if !ok {
						return m, nil
					}
					if m.installed[si.info.Name] {
						m.status = fmt.Sprintf("%s is already installed", si.info.Name)
						m.statusErr = false
						return m, nil
					}
					if m.installing[si.info.Name] {
						return m, nil
					}
					m.installing[si.info.Name] = true
					m.skillList.SetDelegate(skillDelegate{
						installed:    m.installed,
						installing:   m.installing,
						spinnerFrame: m.spinnerFrame,
					})
					installCmd := installSkillCmd(m.cfg, si.info)
					if len(m.installing) == 1 {
						// First install in flight: start the tick loop
						return m, tea.Batch(installCmd, tickCmd())
					}
					return m, installCmd
				}
			}
			var cmd tea.Cmd
			m.skillList, cmd = m.skillList.Update(msg)
			return m, cmd
		}
	}

	// Task 5.8: delegate other messages to the active list
	switch m.state {
	case stateRegPicker:
		var cmd tea.Cmd
		m.regList, cmd = m.regList.Update(msg)
		return m, cmd
	case stateSkillList:
		var cmd tea.Cmd
		m.skillList, cmd = m.skillList.Update(msg)
		return m, cmd
	}
	return m, nil
}

// Task 6.1
func (m skillListModel) View() string {
	switch m.state {
	case stateRegPicker:
		return m.viewRegPicker()
	case stateLoading:
		return m.viewLoading()
	case stateSkillList:
		return m.viewSkillList()
	}
	return ""
}

// Task 6.2
func (m skillListModel) viewRegPicker() string {
	return m.regList.View()
}

// Task 6.3
func (m skillListModel) viewLoading() string {
	return "\n\n  Loading…\n"
}

// Task 6.4
func (m skillListModel) viewSkillList() string {
	helpText := "  ↵ install  esc back  q quit"
	if !m.multiRegistry {
		helpText = "  ↵ install  q quit"
	}
	help := helpStyle.Render(helpText)

	status := ""
	if m.status != "" {
		if m.statusErr {
			status = "\n" + errorStatus.Render(m.status)
		} else {
			status = "\n" + successStatus.Render(m.status)
		}
	}
	return m.skillList.View() + "\n" + help + status
}

// ── tea.Cmd functions ─────────────────────────────────────────────────────────

// Task 3.1
func loadSkillsCmd(cfg *config.Config, registryName string) tea.Cmd {
	return func() tea.Msg {
		skills, err := buildCatalog(cfg, registryName)
		return skillsLoadedMsg{skills: skills, registry: registryName, err: err}
	}
}

// Task 3.2
func installSkillCmd(cfg *config.Config, info skill.Info) tea.Cmd {
	return func() tea.Msg {
		// Reload fresh config to safely handle concurrent installs.
		freshCfg, err := config.Load()
		if err != nil {
			return skillInstalledMsg{name: info.Name, err: err}
		}
		if freshCfg.FindInstalledSkill(info.Name) != nil {
			return skillInstalledMsg{name: info.Name, err: fmt.Errorf("already installed")}
		}
		installedDir := config.InstalledDir()
		if err := os.MkdirAll(installedDir, 0o755); err != nil {
			return skillInstalledMsg{name: info.Name, err: fmt.Errorf("creating installed dir: %w", err)}
		}
		link := filepath.Join(installedDir, info.Name)
		if err := os.Symlink(info.Dir, link); err != nil {
			return skillInstalledMsg{name: info.Name, err: fmt.Errorf("creating symlink: %w", err)}
		}
		freshCfg.InstalledSkills = append(freshCfg.InstalledSkills, config.InstalledSkill{
			Name:     info.Name,
			Registry: info.Registry,
		})
		if err := freshCfg.Save(); err != nil {
			return skillInstalledMsg{name: info.Name, err: fmt.Errorf("saving config: %w", err)}
		}
		backends := len(freshCfg.EnabledBackends)
		if backends > 0 {
			_ = syncer.Run(freshCfg)
		}
		return skillInstalledMsg{name: info.Name, registry: info.Registry, backends: backends}
	}
}

// Task 3.3
func tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(_ time.Time) tea.Msg {
		return tickMsg{}
	})
}

// ── skill list (TUI) ──────────────────────────────────────────────────────────

var skillListFrom string

var skillListCmd = &cobra.Command{
	Use:   "list",
	Short: "Browse available skills from registered registries",
	Long: `Open a TUI browser showing skills available from registered registries.

With multiple registries, first select a registry, then browse its skills.
Press Enter to install a skill. Press Esc to go back, q to quit.
Use --from to open a specific registry directly.`,
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
		if skillListFrom != "" && cfg.FindRegistry(skillListFrom) == nil {
			return fmt.Errorf("registry %q not found\n  → run: sk registry list", skillListFrom)
		}

		m := newSkillListModel(cfg, skillListFrom)
		p := tea.NewProgram(m, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("TUI error: %w", err)
		}
		return nil
	},
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
	skillListCmd.Flags().StringVar(&skillListFrom, "from", "", "Open a specific registry directly (bypasses picker)")
	skillAddCmd.Flags().StringVar(&skillAddFrom, "from", "", "Specify which registry to install from (required when skill exists in multiple)")

	skillCmd.AddCommand(skillListCmd)
	skillCmd.AddCommand(skillAddCmd)
	skillCmd.AddCommand(skillRemoveCmd)
	skillCmd.AddCommand(skillNewCmd)
}
