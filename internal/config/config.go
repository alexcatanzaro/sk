package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
)

// Registry represents a registered remote skill registry.
type Registry struct {
	Name          string    `toml:"name"`
	URL           string    `toml:"url"`
	Path          string    `toml:"path"`
	LastRefreshed time.Time `toml:"last_refreshed"`
}

// InstalledSkill tracks a skill installed into the canonical store.
type InstalledSkill struct {
	Name     string `toml:"name"`
	Registry string `toml:"registry"`
}

// Config is the top-level configuration structure for ~/.config/sk/config.toml.
type Config struct {
	Registries      []Registry       `toml:"registries"`
	InstalledSkills []InstalledSkill `toml:"installed_skills"`
	EnabledBackends []string         `toml:"enabled_backends"`
}

// Path returns the path to the config file.
func Path() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "sk", "config.toml")
}

// DataDir returns ~/.local/share/sk.
func DataDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "sk")
}

// CacheDir returns the registry cache root.
func CacheDir() string {
	return filepath.Join(DataDir(), "cache", "registries")
}

// InstalledDir returns the canonical installed skills directory.
func InstalledDir() string {
	return filepath.Join(DataDir(), "installed")
}

// AuthoredDir returns the authored skills directory.
func AuthoredDir() string {
	return filepath.Join(DataDir(), "authored")
}

// RegistryCacheDir returns the cache directory for a specific registry.
func RegistryCacheDir(name string) string {
	return filepath.Join(CacheDir(), name)
}

// Load reads the config file from disk. Returns an empty Config if the file
// does not exist yet.
func Load() (*Config, error) {
	cfg := &Config{}
	p := Path()
	if _, err := os.Stat(p); os.IsNotExist(err) {
		return cfg, nil
	}
	if _, err := toml.DecodeFile(p, cfg); err != nil {
		return nil, fmt.Errorf("reading config %s: %w", p, err)
	}
	return cfg, nil
}

// Save writes the config to disk, creating parent directories as needed.
func (c *Config) Save() error {
	p := Path()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}
	f, err := os.Create(p)
	if err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(c)
}

// FindRegistry returns the registry with the given name, or nil.
func (c *Config) FindRegistry(name string) *Registry {
	for i := range c.Registries {
		if c.Registries[i].Name == name {
			return &c.Registries[i]
		}
	}
	return nil
}

// RemoveRegistry removes the registry with the given name from config.
func (c *Config) RemoveRegistry(name string) {
	out := c.Registries[:0]
	for _, r := range c.Registries {
		if r.Name != name {
			out = append(out, r)
		}
	}
	c.Registries = out
}

// FindInstalledSkill returns the installed skill with the given name, or nil.
func (c *Config) FindInstalledSkill(name string) *InstalledSkill {
	for i := range c.InstalledSkills {
		if c.InstalledSkills[i].Name == name {
			return &c.InstalledSkills[i]
		}
	}
	return nil
}

// RemoveInstalledSkill removes the installed skill record with the given name.
func (c *Config) RemoveInstalledSkill(name string) {
	out := c.InstalledSkills[:0]
	for _, s := range c.InstalledSkills {
		if s.Name != name {
			out = append(out, s)
		}
	}
	c.InstalledSkills = out
}

// IsBackendEnabled reports whether the named backend is in EnabledBackends.
func (c *Config) IsBackendEnabled(name string) bool {
	for _, b := range c.EnabledBackends {
		if b == name {
			return true
		}
	}
	return false
}

// RemoveBackend removes the named backend from EnabledBackends.
func (c *Config) RemoveBackend(name string) {
	out := c.EnabledBackends[:0]
	for _, b := range c.EnabledBackends {
		if b != name {
			out = append(out, b)
		}
	}
	c.EnabledBackends = out
}
