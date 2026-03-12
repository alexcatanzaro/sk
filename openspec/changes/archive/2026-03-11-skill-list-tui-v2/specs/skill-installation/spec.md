## MODIFIED Requirements

### Requirement: Browse available skills from registered registries
The system SHALL provide a TUI selection interface for browsing skills available in registered registries. When multiple registries are configured, the TUI SHALL first present a registry picker; when only one registry is configured, the TUI SHALL open the skill browser directly. The skill browser SHALL display each skill's name and description, and SHALL visually mark already-installed skills with a ✓ glyph. The user MAY filter by registry using `--from <name>`, which bypasses the registry picker. If `--from` is given, skills from only that registry SHALL be shown.

#### Scenario: Browse skills with single registry
- **WHEN** user runs `sk skill list` with one registry configured
- **THEN** system displays a TUI skill browser for that registry with name and description for each skill

#### Scenario: Browse skills with multiple registries
- **WHEN** user runs `sk skill list` with multiple registries configured
- **THEN** system first displays a registry picker, then on selection displays the skill browser for the chosen registry

#### Scenario: Browse skills filtered by registry
- **WHEN** user runs `sk skill list --from myskills`
- **THEN** system displays the skill browser for only the `myskills` registry, bypassing the picker

#### Scenario: No registries configured
- **WHEN** user runs `sk skill list` with no registries registered
- **THEN** system exits with a message directing the user to run `sk registry add`

### Requirement: Install a skill from a registry
The system SHALL install a skill by name from a registered registry into the canonical store at `~/.local/share/sk/installed/<skill-name>` as a symlink pointing into the registry cache. Skills MAY be installed via `sk skill add <name>` (CLI) or by pressing Enter on an uninstalled skill in the `sk skill list` TUI. If multiple registries contain a skill with the same name and `--from` is not specified, the user MUST be prompted to disambiguate via `--from <registry>`.

#### Scenario: Install skill from single registry via CLI
- **WHEN** user runs `sk skill add pdf-tools`
- **THEN** system creates `~/.local/share/sk/installed/pdf-tools` as a symlink to the cached skill directory and records the installation in config

#### Scenario: Install skill via TUI Enter key
- **WHEN** user presses Enter on an uninstalled skill in the skill browser TUI
- **THEN** system installs the skill asynchronously, shows a spinner during install, and marks it ✓ on success

#### Scenario: Ambiguous skill name
- **WHEN** user runs `sk skill add pdf-tools` and the skill exists in multiple registries
- **THEN** system exits with an error listing the registries that contain it and prompting the user to use `--from`

#### Scenario: Skill not found
- **WHEN** user runs `sk skill add <name>` and no registered registry contains that skill
- **THEN** system exits with an error indicating the skill was not found

#### Scenario: Already installed (CLI)
- **WHEN** user runs `sk skill add` for an already-installed skill
- **THEN** system exits with a message indicating the skill is already installed

#### Scenario: Already installed (TUI)
- **WHEN** user presses Enter on a skill already marked ✓ in the TUI
- **THEN** system shows "already installed" in the status line and takes no further action
