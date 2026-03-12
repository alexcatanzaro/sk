# skill-list-tui Specification

## Purpose
Defines the interactive TUI (terminal user interface) for browsing and installing skills from registered registries via `sk skill list`.

## Requirements

### Requirement: Registry picker for multi-registry environments
When multiple registries are configured, `sk skill list` SHALL display an interactive registry picker as the first screen. The picker SHALL list registry names only (no skill counts). Selecting a registry SHALL lazily load its skills and transition to the skill browser. When only one registry is configured, the picker SHALL be skipped and the skill browser SHALL open directly.

#### Scenario: Multiple registries - picker shown first
- **WHEN** user runs `sk skill list` with two or more registries configured
- **THEN** system displays a registry picker listing all configured registry names

#### Scenario: Single registry - picker skipped
- **WHEN** user runs `sk skill list` with exactly one registry configured
- **THEN** system skips the picker and opens the skill browser for that registry directly

#### Scenario: Registry selected from picker
- **WHEN** user presses Enter on a registry in the picker
- **THEN** system loads skills from that registry and transitions to the skill browser

#### Scenario: --from bypasses picker
- **WHEN** user runs `sk skill list --from <registry>`
- **THEN** system opens the skill browser for that registry directly without showing the picker

### Requirement: Skill browser with installed-state display
The skill browser SHALL display skills for the selected registry. Each skill entry SHALL show the skill name and description. Skills that are already installed SHALL be visually distinguished with a green ✓ glyph. The browser title SHALL display the registry name and total skill count (e.g., "acme · 12 skills").

#### Scenario: Installed skills marked
- **WHEN** the skill browser is open and a skill is already installed
- **THEN** that skill's entry displays a green ✓ glyph

#### Scenario: Uninstalled skills unmarked
- **WHEN** the skill browser is open and a skill is not installed
- **THEN** that skill's entry displays no installation glyph

#### Scenario: Browser title shows registry and count
- **WHEN** the skill browser is open
- **THEN** the title reads "<registry-name> · N skills" where N is the total number of skills in that registry

### Requirement: Inline skill installation from the TUI
In the skill browser, pressing Enter on an uninstalled skill SHALL install it without closing the TUI. The install SHALL run asynchronously so the UI remains responsive. Multiple skills MAY be installed concurrently by pressing Enter on each before the prior install completes.

#### Scenario: Install an uninstalled skill
- **WHEN** user presses Enter on an uninstalled skill in the skill browser
- **THEN** system begins installing the skill and shows a spinner glyph on that skill's entry

#### Scenario: Install completes successfully
- **WHEN** an in-flight install finishes without error
- **THEN** the skill's entry shows a green ✓ glyph and a success status message appears at the bottom of the TUI

#### Scenario: Install fails
- **WHEN** an in-flight install finishes with an error
- **THEN** the skill's entry remains unmarked and a red error status message appears at the bottom of the TUI; the TUI stays open

#### Scenario: Multiple concurrent installs
- **WHEN** user presses Enter on skill-a and then skill-b before skill-a's install finishes
- **THEN** both installs run concurrently, each showing a spinner; each resolves independently

#### Scenario: Enter on already-installed skill
- **WHEN** user presses Enter on a skill already marked ✓
- **THEN** system shows "already installed" in the status line and takes no further action

### Requirement: TUI navigation and exit
In the skill browser, Esc SHALL navigate back to the registry picker when multiple registries are configured, or quit the TUI when only one registry is configured (or when `--from` was used). Pressing q or Ctrl+C SHALL quit the TUI from any screen. In the registry picker, Esc or q SHALL quit the TUI.

#### Scenario: Esc in skill browser (multi-registry)
- **WHEN** user presses Esc in the skill browser and multiple registries were configured
- **THEN** system returns to the registry picker

#### Scenario: Esc in skill browser (single registry or --from)
- **WHEN** user presses Esc in the skill browser and only one registry was configured or --from was used
- **THEN** system quits the TUI

#### Scenario: q or Ctrl+C quits from any screen
- **WHEN** user presses q or Ctrl+C in any TUI screen
- **THEN** system quits the TUI
