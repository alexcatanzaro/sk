## Context

`sk` is a new Go CLI binary with no existing codebase. The agentskills.io specification defines the skill format (SKILL.md with YAML frontmatter in a named directory). 25+ tools follow the same path convention (`~/.<dirname>/skills/`), documented in the OpenSpec project's config.ts. Registries are plain git repositories; `sk` has no server-side component.

## Goals / Non-Goals

**Goals:**
- Manage agentskills.io-compliant skills across any registered git registry
- Sync installed skills to any configured tool adapter with a single command
- Support skill authoring and publishing via git, with zero custom auth infrastructure
- Keep the adapter contribution surface minimal (one line per new tool)

**Non-Goals:**
- Custom registry server or API — git repos only
- Skills format other than agentskills.io SKILL.md
- Plugin/extension system for adapters — all adapters compiled into the binary
- Version pinning or lockfile — HEAD of the registry, manual refresh

## Decisions

### D1: Canonical store with symlinks into sparse-checkout cache

**Decision:** `~/.local/share/sk/` holds the canonical state.
```
~/.local/share/sk/
├── cache/registries/<name>/   ← git sparse checkout (skills subpath only)
└── installed/                 ← symlinks into cache entries
```
`sk sync` copies (not symlinks) from `installed/` into each adapter's path.

**Rationale:** Separates "what I want" from "where it goes." When a new backend is enabled after installation, `sk sync` can populate it without re-fetching from the registry. Symlinks in the store mean `sk registry refresh` is reflected immediately without re-copying. Copies at the adapter level mean each tool has a durable, self-contained skill directory that survives sk cache operations.

**Alternatives considered:** Direct sync from registry → adapter paths (no store). Rejected because enabling a new backend after the fact would require re-fetching all skills.

### D2: Git sparse checkout for registry fetching

**Decision:** Use `git sparse-checkout` to fetch only the configured skills subpath from each registry repo.

**Rationale:** Users may register monorepos where the skills subpath is a small fraction of the repository. Fetching the full repo wastes bandwidth and time. Sparse checkout is well-supported in modern git and requires no external dependencies.

**Alternatives considered:** Full shallow clone (`--depth 1`). Simpler, but unsuitable for monorepos with a `--path` override.

### D3: Single StandardAdapter struct covering all tools

**Decision:**
```go
type StandardAdapter struct {
    name    string
    dirname string  // e.g., ".claude", ".codex", ".github"
}
```
All known adapters are entries in a `KnownAdapters` slice derived from OpenSpec's tool list. Adding a new tool is one line. `IsAvailable()` checks whether `~/.<dirname>/` exists.

**Rationale:** All 25+ agentskills.io-compliant tools follow the same path convention. No tool-specific logic is needed. Keeping adapters data-driven (not separate files) minimizes contribution friction.

**Alternatives considered:** Interface with `Transform(Skill) Skill` method. Rejected — since all supported tools use the same SKILL.md format, Transform is always identity. Can be added later if a non-conforming tool is added.

### D4: Availability gated at configuration, not sync time

**Decision:** `sk backend enable <name>` calls `IsAvailable()` and rejects if the tool isn't detected. `sk sync` trusts that configured backends are valid.

**Rationale:** Prevents silent failures at sync time. User intent is expressed at configuration; sync should be fast and quiet.

### D5: Registry index via filesystem walk, not manifest

**Decision:** `sk` discovers skills by walking the registry's skills root (default `skills/`, or `--path` override) one level deep, finding directories containing `SKILL.md`.

**Rationale:** No contract imposed on registry authors beyond placing skill directories in the right location. Works with any existing git repo that follows the agentskills.io directory convention.

**Alternatives considered:** Explicit `index.json` manifest. Better for large registries and richer metadata, but requires registry authors to maintain it. Deferred to a future enhancement.

### D6: Auth delegated entirely to git

**Decision:** `sk publish` shells out to `git push`. No credential management in sk.

**Rationale:** Users already have SSH keys or HTTPS tokens configured for git. Reinventing credential management adds complexity and security surface with no benefit.

### D7: Config format — TOML at ~/.config/sk/config.toml

**Decision:** Single TOML config file tracking registries, installed skills, and enabled backends.

**Rationale:** TOML is human-readable and has strong Go library support (`github.com/BurntSushi/toml`). Single file keeps state simple; no database needed.

## Risks / Trade-offs

- **Stale registry cache** → Mitigation: `sk registry refresh` is explicit and clearly documented. Future auto-refresh config flag.
- **git sparse checkout compatibility** → Mitigation: Require git ≥ 2.25 (sparse-checkout was stabilized then); print clear error if version is insufficient.
- **IsAvailable() false positive** → `~/.<dirname>/` may exist for reasons other than the tool being installed. Mitigation: Acceptable tradeoff; tool availability detection is best-effort.
- **No version pinning** → Skills always track HEAD of the registry. Mitigation: Documented as a limitation; lockfile support deferred.
- **Symlink breakage** → If `sk cache` is cleared, `installed/` symlinks break. Mitigation: `sk sync` detects broken symlinks and warns; `sk registry refresh` repairs them.

## Open Questions

- TUI library choice for `sk skill list` selection screens: bubbletea (charm.sh) vs. simpler promptui. Bubbletea preferred for richer UX but heavier dependency.
- Whether `sk skill add` without `--from` should prompt to select a registry when multiple are registered, or error and require explicit `--from`.
