## ADDED Requirements

### Requirement: Scaffold a new skill
The system SHALL provide a command to scaffold a new skill directory at `~/.local/share/sk/authored/<name>/` containing a `SKILL.md` with pre-filled frontmatter (name, description placeholder) and an empty body section. The skill name SHALL be validated against the agentskills.io naming rules before scaffolding.

#### Scenario: Scaffold new skill
- **WHEN** user runs `sk skill new pdf-tools`
- **THEN** system creates `~/.local/share/sk/authored/pdf-tools/SKILL.md` with valid frontmatter and prints the path for the user to edit

#### Scenario: Invalid skill name
- **WHEN** user runs `sk skill new` with a name containing uppercase letters, consecutive hyphens, or leading/trailing hyphens
- **THEN** system exits with an error describing the naming constraint violated

#### Scenario: Skill name already exists in authored
- **WHEN** user runs `sk skill new <name>` and `authored/<name>/` already exists
- **THEN** system exits with an error indicating the skill already exists locally

### Requirement: Validate a skill before publishing
The system SHALL validate a skill's `SKILL.md` frontmatter against the agentskills.io specification before publishing. Validation SHALL check: `name` is present and matches the directory name; `description` is present and non-empty; `name` conforms to character and length constraints. Validation failures SHALL block publishing with a clear error message per violation.

#### Scenario: Valid skill passes validation
- **WHEN** a skill's SKILL.md has a valid name and non-empty description matching all spec constraints
- **THEN** system proceeds to publish

#### Scenario: Missing description
- **WHEN** SKILL.md frontmatter has no `description` field or an empty value
- **THEN** system exits with an error: "description is required"

#### Scenario: Name mismatch
- **WHEN** SKILL.md `name` field does not match the parent directory name
- **THEN** system exits with an error: "name field '<value>' does not match directory name '<dir>'"

#### Scenario: Invalid name characters
- **WHEN** SKILL.md `name` contains uppercase letters, consecutive hyphens, or starts/ends with a hyphen
- **THEN** system exits with an error describing the specific constraint violated

### Requirement: Publish a skill to a remote registry
The system SHALL copy the skill directory into the registry's sparse-checkout cache, then perform `git add`, `git commit`, and `git push` using the user's existing git credentials (SSH or HTTPS). The target registry MUST be specified with `--to <name>`. If only one registry is configured, it SHALL be the default. Publishing SHALL fail with the git error output if the push is rejected (e.g., no write access, conflicts).

#### Scenario: Publish to specified registry
- **WHEN** user runs `sk publish pdf-tools --to myskills`
- **THEN** system validates the skill, copies it into the registry cache, commits, and pushes to the remote

#### Scenario: Publish with single registry configured
- **WHEN** user runs `sk publish pdf-tools` with exactly one registry configured
- **THEN** system uses that registry without requiring `--to`

#### Scenario: Publish with multiple registries, no --to
- **WHEN** user runs `sk publish pdf-tools` with multiple registries configured and no `--to`
- **THEN** system exits with an error asking the user to specify `--to <registry>`

#### Scenario: Push rejected by remote
- **WHEN** git push fails (no write access or conflict)
- **THEN** system exits with git's error output and leaves the registry cache in its pre-publish state
