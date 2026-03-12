package skill

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Frontmatter holds the parsed SKILL.md YAML frontmatter fields.
type Frontmatter struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	// Additional optional fields are captured in Extra.
	Extra map[string]interface{} `yaml:",inline"`
}

// Info is a discovered skill with its source registry and path.
type Info struct {
	Frontmatter
	Registry string // source registry name
	Dir      string // absolute path to the skill directory
}

// ParseFrontmatter reads a SKILL.md file at path and extracts frontmatter.
// If the file has no YAML frontmatter delimiters, or the YAML is malformed,
// it returns an empty Frontmatter with a non-nil error (caller may warn and skip).
func ParseFrontmatter(path string) (Frontmatter, error) {
	f, err := os.Open(path)
	if err != nil {
		return Frontmatter{}, fmt.Errorf("opening %s: %w", path, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var lines []string
	inFrontmatter := false
	sawOpen := false

	for scanner.Scan() {
		line := scanner.Text()
		if !sawOpen && line == "---" {
			sawOpen = true
			inFrontmatter = true
			continue
		}
		if inFrontmatter && line == "---" {
			break
		}
		if inFrontmatter {
			lines = append(lines, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return Frontmatter{}, fmt.Errorf("reading %s: %w", path, err)
	}
	if !sawOpen {
		return Frontmatter{}, fmt.Errorf("%s: no YAML frontmatter found", path)
	}

	var fm Frontmatter
	raw := strings.Join(lines, "\n")
	if err := yaml.Unmarshal([]byte(raw), &fm); err != nil {
		return Frontmatter{}, fmt.Errorf("%s: malformed YAML frontmatter: %w", path, err)
	}
	return fm, nil
}

// ValidateName checks that name conforms to agentskills.io naming rules:
//   - lowercase alphanumeric and hyphens only
//   - 1–64 characters
//   - no leading, trailing, or consecutive hyphens
func ValidateName(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("name must not be empty")
	}
	if len(name) > 64 {
		return fmt.Errorf("name must be 64 characters or fewer (got %d)", len(name))
	}
	validChars := regexp.MustCompile(`^[a-z0-9-]+$`)
	if !validChars.MatchString(name) {
		return fmt.Errorf("name %q contains invalid characters; only lowercase letters, digits, and hyphens are allowed", name)
	}
	if strings.HasPrefix(name, "-") {
		return fmt.Errorf("name %q must not start with a hyphen", name)
	}
	if strings.HasSuffix(name, "-") {
		return fmt.Errorf("name %q must not end with a hyphen", name)
	}
	if strings.Contains(name, "--") {
		return fmt.Errorf("name %q must not contain consecutive hyphens", name)
	}
	return nil
}

// Validate checks a skill directory for publishability.
// dir is the path to the skill directory (e.g., authored/my-skill).
func Validate(dir string) error {
	dirName := filepath.Base(dir)
	skillFile := filepath.Join(dir, "SKILL.md")
	fm, err := ParseFrontmatter(skillFile)
	if err != nil {
		return err
	}

	if fm.Name == "" {
		return fmt.Errorf("name is required in SKILL.md frontmatter")
	}
	if err := ValidateName(fm.Name); err != nil {
		return err
	}
	if fm.Name != dirName {
		return fmt.Errorf("name field %q does not match directory name %q", fm.Name, dirName)
	}
	if strings.TrimSpace(fm.Description) == "" {
		return fmt.Errorf("description is required")
	}
	return nil
}

// SkillTemplate returns the content of a scaffolded SKILL.md for the given name.
func SkillTemplate(name string) string {
	return fmt.Sprintf(`---
name: %s
description: A short description of what this skill does.
---

## Overview

Describe the skill here.
`, name)
}
