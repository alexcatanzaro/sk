## Why

Developers using AI agents across multiple tools (Claude Code, Codex, Copilot, Opencode, Cursor, and 20+ others) have no standard way to discover, install, or share agentskills.io-compliant skills. Skills are manually copied between tool-specific paths, registries don't exist in a usable form, and authoring a skill has no workflow. `sk` solves this with a dedicated CLI package manager for the agentskills.io ecosystem.

## What Changes

- New CLI binary `sk` written in Go
- Registry management: add/list/refresh/remove remote git repositories as skill registries with configurable index subpaths
- Skill installation: browse and install skills from registered registries via TUI selection screens
- Backend sync: copy installed skills to any of 25+ agentskills.io-compliant tool paths via an opt-in adapter system
- Skill authoring: scaffold, validate, and publish new skills to remote registries using existing git credentials
- Config file at `~/.config/sk/config.toml` tracking registries, installed skills, and enabled backends

## Capabilities

### New Capabilities

- `registry-management`: Add, list, refresh, and remove remote git repositories as skill registries; sparse checkout of skills subpath only; registry named by URL tail with optional override
- `skill-installation`: Browse available skills from registered registries via TUI; install skills into sk's canonical store; remove installed skills
- `backend-sync`: Adapter system mapping installed skills to tool-specific paths (`~/.<dirname>/skills/`); opt-in configuration gated by tool availability detection; `sk sync` copies from store to all enabled backends
- `skill-authoring`: Scaffold new skills with `sk skill new`; validate SKILL.md frontmatter against agentskills.io spec; publish to a remote registry via git push using system credentials

### Modified Capabilities

## Impact

- New Go binary and module at `github.com/alexcatanzaro/sk`
- Reads/writes `~/.config/sk/config.toml`
- Canonical skill store at `~/.local/share/sk/`
- Shells out to `git` for registry operations and publishing; relies on user's existing SSH/HTTPS credentials
- No runtime dependencies on any AI tool; works alongside all agentskills.io-compliant clients
