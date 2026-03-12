## 1. Project Scaffolding

- [ ] 1.1 Initialize Go module with `github.com/alexcatanzaro/sk`, set up `main.go` entry point
- [ ] 1.2 Add dependencies: CLI framework (cobra), TOML config (BurntSushi/toml), YAML frontmatter parser, TUI library (bubbletea)
- [ ] 1.3 Define top-level command structure with cobra: `registry`, `skill`, `backend`, `sync`, `publish` subcommands
- [ ] 1.4 Implement config file loading/saving at `~/.config/sk/config.toml` with typed Go structs

## 2. Registry Management

- [ ] 2.1 Implement `sk registry add <url>` — parse URL tail as default name, accept `--name` and `--path` flags, write to config
- [ ] 2.2 Implement git version check (≥ 2.25) with clear error output
- [ ] 2.3 Implement sparse checkout fetch: `git init`, `git sparse-checkout set <subpath>`, `git pull --depth 1` into `~/.local/share/sk/cache/registries/<name>/`
- [ ] 2.4 Implement `sk registry list` — tabular output of name, URL, path, last-refreshed
- [ ] 2.5 Implement `sk registry refresh [<name>]` — git fetch on single or all registry caches, update last-refreshed timestamp in config
- [ ] 2.6 Implement `sk registry remove <name>` — remove from config, delete cache dir, warn about installed skills with broken symlinks

## 3. Skill Discovery

- [ ] 3.1 Implement registry walker: scan `<cache>/<registry>/skills/` one level deep for directories containing `SKILL.md`
- [ ] 3.2 Implement SKILL.md frontmatter parser: extract `name` and `description` (and optional fields); handle malformed YAML gracefully
- [ ] 3.3 Implement skill catalog builder: aggregate skills across all (or specified) registries with source registry tracked

## 4. Skill Installation

- [ ] 4.1 Implement `sk skill list [--from <registry>]` — TUI selection screen showing name + description, sourced from skill catalog
- [ ] 4.2 Implement `sk skill add <name> [--from <registry>]` — disambiguate across registries, create symlink in `installed/`, update config
- [ ] 4.3 Implement `sk skill remove <name>` — remove symlink from `installed/`, update config

## 5. Backend Adapters

- [ ] 5.1 Define `Adapter` interface (`Name()`, `SkillsPath()`, `IsAvailable()`) and `StandardAdapter` struct
- [ ] 5.2 Populate `KnownAdapters` slice from OpenSpec tool list (claude, codex, copilot/github, opencode, cursor, windsurf, gemini, and remaining tools)
- [ ] 5.3 Implement `sk backend list` — tabular output: name, available (~/.<dirname>/ exists), enabled, skills path
- [ ] 5.4 Implement `sk backend enable <name>` — call `IsAvailable()`, reject if false, write to config, trigger sync
- [ ] 5.5 Implement `sk backend disable <name>` — remove from config, leave tool's skills directory untouched

## 6. Sync

- [ ] 6.1 Implement `sk sync` core: iterate `installed/`, detect broken symlinks (warn + skip), copy each skill directory into each enabled backend's skills path
- [ ] 6.2 Implement orphan cleanup: delete skills from backend paths that are no longer in `installed/`
- [ ] 6.3 Wire `sk sync` to trigger automatically after `sk skill add` and `sk backend enable`

## 7. Skill Authoring

- [ ] 7.1 Implement SKILL.md name validator: check character set, length, leading/trailing/consecutive hyphens, directory name match
- [ ] 7.2 Implement `sk skill new <name>` — validate name, scaffold `~/.local/share/sk/authored/<name>/SKILL.md` with frontmatter template, print path
- [ ] 7.3 Implement `sk publish <name> --to <registry>` — validate skill, copy into registry sparse checkout, `git add/commit/push`, surface git errors

## 8. Polish

- [ ] 8.1 Add `--help` descriptions to all commands and flags
- [ ] 8.2 Ensure all error messages are actionable (include remediation hint where applicable)
- [ ] 8.3 Add `sk version` command
