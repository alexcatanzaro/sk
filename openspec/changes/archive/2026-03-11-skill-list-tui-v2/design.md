## Context

`sk skill list` currently uses a flat bubbletea TUI that loads all skills from all registries at startup and presents them in a single `bubbles/list` component. Users cannot install from within the TUI — they must quit and run `sk skill add <name>`. The existing model is a simple struct wrapping one `list.Model` with no state machine.

All install logic lives in `internal/syncer` and `internal/config`, making it straightforward to invoke from inside a `tea.Cmd`. No new dependencies are required.

## Goals / Non-Goals

**Goals:**
- Two-level TUI: registry picker (multi-registry) → skill browser per registry
- Lazy skill loading (load only when a registry is selected)
- Inline install via Enter; multiple concurrent installs allowed
- Visual marking of installed skills (✓) and in-flight installs (spinner)
- Status line feedback for success and failure; TUI stays open on failure
- No behavioral changes to `sk skill add`, `sk skill remove`, or any `internal/` package

**Non-Goals:**
- Uninstall from within the TUI
- Pagination or streaming for very large registries
- Install progress beyond the spinner glyph (no byte-level progress)

## Decisions

### Decision: Custom delegate over item mutation

**Chosen:** A custom `skillDelegate` struct holds `installed map[string]bool`, `installing map[string]bool`, and `spinnerFrame int` as pointers/values from the model, and renders each item with the appropriate glyph on every frame.

**Alternatives considered:**
- Rebuilding the `[]list.Item` slice on each state change and calling `list.SetItems()` — loses filter/cursor state and is expensive.
- Calling `list.SetItem(index, item)` per-skill — requires O(n) index lookup and recreating item structs repeatedly.

The custom delegate is the idiomatic bubbletea approach: rendering is a pure function of model state, no item mutation needed.

### Decision: Lazy loading via tea.Cmd

**Chosen:** The registry picker is instantiated from `cfg.Registries` names alone (instant). On Enter, a `loadSkillsCmd` goroutine walks the registry cache dir and returns `skillsLoadedMsg`. A `stateLoading` transition provides a visual placeholder.

**Alternatives considered:**
- Pre-loading all registries at startup to show counts in the picker — requires scanning all caches upfront, adds latency, and the counts were explicitly out-of-scope for the picker level.

Lazy loading means the picker renders instantly and skill counts appear in the skill-list title only after selection.

### Decision: State machine with three states

```
stateRegPicker → (Enter) → stateLoading → (skillsLoadedMsg) → stateSkillList
                                                                      │
                                              (Esc, multi-registry) ──┘ → stateRegPicker
                                              (Esc, single registry) ──→ quit
```

`stateLoading` is a transient state between picker and skill list during the async walk. It prevents user input from being processed while the goroutine runs.

### Decision: Spinner via tea.Tick, not bubbles/spinner

**Chosen:** A `tickCmd()` function returns `tea.Tick(100ms, tickMsg{})`. The model increments `spinnerFrame` on each `tickMsg` only when `len(installing) > 0`. The delegate uses `spinnerFrames[spinnerFrame % len(spinnerFrames)]` to pick the glyph.

**Alternatives considered:**
- `bubbles/spinner` component — designed as a single global spinner, not per-item. Would require a separate model embedded in the parent, with its own tick lifecycle.

The manual tick approach is simpler, integrates naturally with the delegate pattern, and stops automatically when no installs are in flight.

### Decision: Stay in TUI on install failure

Install failures render a red status line at the bottom; the skill remains unmarked. The user can attempt another skill or quit. This matches the multi-install workflow where a failure on one skill shouldn't interrupt the session.

### Decision: --from bypasses picker

`sk skill list --from <registry>` populates `cfg` with only that registry and enters `stateSkillList` directly (or `stateLoading` while the cache walk runs). Esc in this mode quits rather than going back to a picker that was never shown.

## Risks / Trade-offs

- **Cache walk latency** → Skills load per-registry on selection. For very large registries this could be perceptible (~100ms). Mitigation: the `stateLoading` state shows a "Loading…" placeholder so the UI doesn't appear frozen.
- **Concurrent install race on config.Save()** → Multiple goroutines may call `cfg.Save()` concurrently if several Enter presses fire before any goroutine completes. Mitigation: each goroutine reloads config from disk, applies its change, and saves — the last write wins. In practice, goroutines resolve within milliseconds and the installed-skills list is append-only, making conflicts unlikely. A future change could add a mutex around Save().
- **Installed map diverges from disk state** → The model's `installed` map is initialized at TUI startup and updated on `skillInstalledMsg`. If another process installs a skill mid-session, the TUI won't reflect it. Acceptable for v1.
