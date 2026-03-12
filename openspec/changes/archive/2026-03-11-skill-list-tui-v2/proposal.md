## Why

`sk skill list` currently shows a flat, undifferentiated list of all skills from all registries with no way to install from within the TUI — users must quit and run `sk skill add <name>` separately. As registries grow, the flat list becomes harder to navigate, and the disconnected install flow creates unnecessary friction.

## What Changes

- `sk skill list` gains a two-level TUI: a registry picker (multi-registry) followed by a per-registry skill browser
- Already-installed skills are visually marked with a green ✓ in the skill list
- Pressing Enter on a skill installs it inline without leaving the TUI; multiple concurrent installs are supported
- In-flight installs display a spinner glyph; success and failure are shown as a status line at the bottom
- Single-registry users skip the picker and go directly to the skill list
- `sk skill list --from <registry>` continues to work, bypassing the picker

## Capabilities

### New Capabilities

- `skill-list-tui`: Two-level interactive TUI for browsing and installing skills — registry picker + skill browser with inline install, spinner feedback, and installed-state display

### Modified Capabilities

- `skill-installation`: The `sk skill list` command now supports installation directly from the TUI (Enter to install), expanding the ways a skill can be installed beyond `sk skill add <name>`

## Impact

- `cmd/skill.go`: Full rewrite of `skillListModel`, `skillDelegate`, and `skillListCmd`; new `tea.Cmd` functions for lazy skill loading and async installation
- No changes to `internal/` packages — install logic in `syncer` and `config` is reused as-is
- No breaking changes to `sk skill add`, `sk skill remove`, or any other commands
- No new dependencies (bubbletea, bubbles, lipgloss already present)
