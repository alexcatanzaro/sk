package adapter

import (
	"os"
	"path/filepath"
)

// Adapter describes a backend tool that can receive installed skills.
type Adapter interface {
	Name() string
	SkillsPath() string
	IsAvailable() bool
}

// StandardAdapter implements Adapter for any agentskills.io-compliant tool
// that stores skills at ~/.<dirname>/skills/.
type StandardAdapter struct {
	name    string
	dirname string // e.g. ".claude", ".codex"
}

func (a StandardAdapter) Name() string { return a.name }

func (a StandardAdapter) SkillsPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, a.dirname, "skills")
}

func (a StandardAdapter) IsAvailable() bool {
	home, _ := os.UserHomeDir()
	_, err := os.Stat(filepath.Join(home, a.dirname))
	return err == nil
}

// KnownAdapters is the authoritative list of supported agentskills.io-compliant tools.
// To add a new tool: one line, name + dirname. No other changes required.
var KnownAdapters = []StandardAdapter{
	{name: "claude-code", dirname: ".claude"},
	{name: "codex", dirname: ".codex"},
	{name: "opencode", dirname: ".opencode"},
	{name: "cursor", dirname: ".cursor"},
	{name: "windsurf", dirname: ".windsurf"},
	{name: "gemini", dirname: ".gemini"},
	{name: "github-copilot", dirname: ".github"},
	{name: "cline", dirname: ".cline"},
	{name: "continue", dirname: ".continue"},
	{name: "aider", dirname: ".aider"},
	{name: "goose", dirname: ".goose"},
	{name: "zed", dirname: ".zed"},
	{name: "amp", dirname: ".amp"},
	{name: "roo-cline", dirname: ".roo-cline"},
	{name: "plandex", dirname: ".plandex"},
	{name: "avante", dirname: ".avante"},
	{name: "supermaven", dirname: ".supermaven"},
	{name: "void", dirname: ".void"},
	{name: "kodu", dirname: ".kodu"},
	{name: "tabnine", dirname: ".tabnine"},
	{name: "mentat", dirname: ".mentat"},
	{name: "sweep", dirname: ".sweep"},
	{name: "gpt-pilot", dirname: ".gpt-pilot"},
	{name: "rift", dirname: ".rift"},
	{name: "bloop", dirname: ".bloop"},
}

// Find returns the adapter with the given name, or nil.
func Find(name string) *StandardAdapter {
	for i := range KnownAdapters {
		if KnownAdapters[i].Name() == name {
			return &KnownAdapters[i]
		}
	}
	return nil
}
