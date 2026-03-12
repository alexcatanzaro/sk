# skill-installation Specification

## Purpose
TBD - created by archiving change sk-package-manager. Update Purpose after archive.
## Requirements
### Requirement: Browse available skills from registered registries
The system SHALL provide a TUI selection interface for browsing skills available in registered registries. The interface SHALL display each skill's name and description. The user MAY filter by registry using `--from <name>`. If multiple registries are registered and no `--from` is given, skills from all registries SHALL be shown with their source registry indicated.

#### Scenario: Browse skills with single registry
- **WHEN** user runs `sk skill list`
- **THEN** system displays a TUI list of all skills from all registered registries with name and description

#### Scenario: Browse skills filtered by registry
- **WHEN** user runs `sk skill list --from myskills`
- **THEN** system displays only skills from the `myskills` registry

#### Scenario: No registries configured
- **WHEN** user runs `sk skill list` with no registries registered
- **THEN** system exits with a message directing the user to run `sk registry add`

### Requirement: Install a skill from a registry
The system SHALL install a skill by name from a registered registry into the canonical store at `~/.local/share/sk/installed/<skill-name>` as a symlink pointing into the registry cache. If multiple registries contain a skill with the same name, the user MUST specify `--from <registry>` to disambiguate.

#### Scenario: Install skill from single registry
- **WHEN** user runs `sk skill add pdf-tools`
- **THEN** system creates `~/.local/share/sk/installed/pdf-tools` as a symlink to the cached skill directory and records the installation in config

#### Scenario: Ambiguous skill name
- **WHEN** user runs `sk skill add pdf-tools` and the skill exists in multiple registries
- **THEN** system exits with an error listing the registries that contain it and prompting the user to use `--from`

#### Scenario: Skill not found
- **WHEN** user runs `sk skill add <name>` and no registered registry contains that skill
- **THEN** system exits with an error indicating the skill was not found

#### Scenario: Already installed
- **WHEN** user runs `sk skill add` for an already-installed skill
- **THEN** system exits with a message indicating the skill is already installed

### Requirement: Remove an installed skill
The system SHALL allow users to remove an installed skill by name. Removing a skill SHALL delete its symlink from `installed/` and remove it from config. The system SHALL NOT delete the skill from the registry cache.

#### Scenario: Remove installed skill
- **WHEN** user runs `sk skill remove pdf-tools`
- **THEN** system removes the `installed/pdf-tools` symlink and updates config

#### Scenario: Remove skill not installed
- **WHEN** user runs `sk skill remove <name>` for a skill not in installed/
- **THEN** system exits with an error indicating the skill is not installed

