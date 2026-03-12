# registry-management Specification

## Purpose
TBD - created by archiving change sk-package-manager. Update Purpose after archive.
## Requirements
### Requirement: Register a remote git repository as a skills registry
The system SHALL allow users to add a remote git repository as a named skills registry. The registry SHALL be identified by a short name defaulting to the last path segment of the repository URL. The user MAY override the name with a `--name` flag. The user MAY specify a subpath within the repository where skills are located using a `--path` flag; if omitted, the system SHALL default to `skills/`.

#### Scenario: Add registry with defaults
- **WHEN** user runs `sk registry add https://github.com/org/skills-repo`
- **THEN** system registers the repository under the name `skills-repo` with skills root `skills/`

#### Scenario: Add registry with custom name
- **WHEN** user runs `sk registry add https://github.com/org/skills-repo --name myskills`
- **THEN** system registers the repository under the name `myskills`

#### Scenario: Add registry with custom subpath
- **WHEN** user runs `sk registry add https://github.com/org/monorepo --path packages/skills`
- **THEN** system registers the repository with skills root `packages/skills`

#### Scenario: Duplicate registry name rejected
- **WHEN** user runs `sk registry add` with a name that already exists in config
- **THEN** system exits with an error message indicating the name is already in use

### Requirement: Fetch registry content via git sparse checkout
The system SHALL clone the registry using `git sparse-checkout` restricted to the configured skills subpath, storing the result in `~/.local/share/sk/cache/registries/<name>/`. The system SHALL perform a shallow fetch (`--depth 1`). The system SHALL require git version ≥ 2.25 and SHALL print a clear error if this requirement is not met.

#### Scenario: Successful sparse checkout
- **WHEN** a registry is added or refreshed
- **THEN** system fetches only the skills subpath content into the local cache

#### Scenario: Git version too old
- **WHEN** system detects git version < 2.25
- **THEN** system exits with an error: "sk requires git ≥ 2.25 for sparse checkout support"

#### Scenario: Network failure during fetch
- **WHEN** git fetch fails due to network error
- **THEN** system exits with the git error output and leaves existing cache intact

### Requirement: List registered registries
The system SHALL provide a command to list all registered registries, showing name, URL, skills subpath, and last refresh time.

#### Scenario: List with registries present
- **WHEN** user runs `sk registry list`
- **THEN** system prints each registry's name, URL, path, and last-refreshed timestamp

#### Scenario: List with no registries
- **WHEN** no registries are registered
- **THEN** system prints a message indicating no registries are configured

### Requirement: Refresh registry cache
The system SHALL provide a command to re-fetch a registry's content from the remote. The user MAY specify a registry name to refresh a single registry; if omitted, all registries SHALL be refreshed.

#### Scenario: Refresh single registry
- **WHEN** user runs `sk registry refresh myskills`
- **THEN** system performs a git fetch for the `myskills` registry cache

#### Scenario: Refresh all registries
- **WHEN** user runs `sk registry refresh` with no argument
- **THEN** system refreshes all registered registry caches in sequence

### Requirement: Remove a registry
The system SHALL allow users to remove a registered registry by name. Removing a registry SHALL NOT automatically uninstall skills that were sourced from it, but SHALL warn the user that installed skills from this registry will have broken cache symlinks until re-fetched.

#### Scenario: Remove existing registry
- **WHEN** user runs `sk registry remove myskills`
- **THEN** system removes `myskills` from config and deletes its cache directory, printing a warning about any installed skills from that registry

#### Scenario: Remove nonexistent registry
- **WHEN** user runs `sk registry remove` with a name not in config
- **THEN** system exits with an error indicating the registry was not found

