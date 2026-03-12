## ADDED Requirements

### Requirement: List known backends and their availability
The system SHALL provide a command listing all built-in adapters, indicating for each whether it is available (tool detected), enabled (in config), and the resolved skills path.

#### Scenario: List backends
- **WHEN** user runs `sk backend list`
- **THEN** system prints each known adapter with columns: name, status (available/unavailable), enabled (yes/no), and skills path

### Requirement: Enable a backend
The system SHALL allow users to enable a backend adapter by name. Before enabling, the system SHALL call `IsAvailable()` for that adapter; if the tool is not detected (i.e., `~/.<dirname>/` does not exist), the system SHALL refuse with a clear error. If available, the adapter name SHALL be written to config and `sk sync` SHALL be run immediately to populate the newly enabled backend with all currently installed skills.

#### Scenario: Enable available backend
- **WHEN** user runs `sk backend enable claude-code` and `~/.claude/` exists
- **THEN** system adds `claude-code` to config and syncs all installed skills to `~/.claude/skills/`

#### Scenario: Enable unavailable backend
- **WHEN** user runs `sk backend enable codex` and `~/.codex/` does not exist
- **THEN** system exits with an error: "codex not detected (~/.codex/ not found)"

#### Scenario: Enable unknown backend
- **WHEN** user runs `sk backend enable unknown-tool`
- **THEN** system exits with an error listing valid backend names

#### Scenario: Enable already-enabled backend
- **WHEN** user runs `sk backend enable` for a backend already in config
- **THEN** system prints a message indicating it is already enabled and exits cleanly

### Requirement: Disable a backend
The system SHALL allow users to disable a backend adapter by name, removing it from config. Disabling SHALL NOT delete skills from the tool's skills directory.

#### Scenario: Disable enabled backend
- **WHEN** user runs `sk backend disable claude-code`
- **THEN** system removes `claude-code` from config; existing files in `~/.claude/skills/` are left in place

#### Scenario: Disable not-enabled backend
- **WHEN** user runs `sk backend disable <name>` for a backend not in config
- **THEN** system exits with an error indicating the backend is not enabled

### Requirement: Sync installed skills to all enabled backends
The system SHALL copy all installed skills from `~/.local/share/sk/installed/` into each enabled backend's skills path. Sync SHALL copy the full skill directory tree. Existing files at the destination SHALL be overwritten. Skills removed from `installed/` since the last sync SHALL be deleted from backend paths.

#### Scenario: Sync with installed skills and enabled backends
- **WHEN** user runs `sk sync`
- **THEN** system copies each installed skill directory into each enabled backend's `skills/` path

#### Scenario: Sync with broken symlink in installed/
- **WHEN** an entry in `installed/` is a broken symlink (registry cache was cleared)
- **THEN** system prints a warning for that skill and skips it, continuing with the rest

#### Scenario: Sync with no enabled backends
- **WHEN** user runs `sk sync` with no backends enabled
- **THEN** system exits with a message directing the user to run `sk backend enable`

#### Scenario: Sync removes orphaned skills
- **WHEN** a skill exists in a backend's skills path but is no longer in `installed/`
- **THEN** system deletes it from the backend path during sync

### Requirement: StandardAdapter covers all agentskills.io-compliant tools
The system SHALL implement a `StandardAdapter` type parameterized by tool name and directory name (e.g., `".claude"`, `".codex"`). All known adapters SHALL be entries in a `KnownAdapters` slice. `SkillsPath()` SHALL return `filepath.Join(home, dirname, "skills")`. `IsAvailable()` SHALL return true if `filepath.Join(home, dirname)` exists.

#### Scenario: New tool adapter contribution
- **WHEN** a contributor adds a new entry to `KnownAdapters` with the tool's name and dirname
- **THEN** the tool becomes selectable via `sk backend enable` with no other code changes required
