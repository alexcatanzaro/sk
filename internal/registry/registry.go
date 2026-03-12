package registry

import (
	"os"
	"path/filepath"

	"github.com/alexcatanzaro/sk/internal/skill"
)

// WalkSkills scans the skills subpath of a registry cache directory one level
// deep, returning Info for each directory that contains a SKILL.md.
// Malformed SKILL.md files are warned about but skipped, not fatal.
func WalkSkills(cacheDir, registryName, subpath string) ([]skill.Info, error) {
	skillsRoot := filepath.Join(cacheDir, subpath)
	entries, err := os.ReadDir(skillsRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var skills []skill.Info
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		skillDir := filepath.Join(skillsRoot, e.Name())
		skillFile := filepath.Join(skillDir, "SKILL.md")
		if _, err := os.Stat(skillFile); os.IsNotExist(err) {
			continue
		}
		fm, err := skill.ParseFrontmatter(skillFile)
		if err != nil {
			// warn but don't fail
			continue
		}
		// Use directory name as canonical name if frontmatter name is missing.
		if fm.Name == "" {
			fm.Name = e.Name()
		}
		skills = append(skills, skill.Info{
			Frontmatter: fm,
			Registry:    registryName,
			Dir:         skillDir,
		})
	}
	return skills, nil
}
