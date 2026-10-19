package backup

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// Config is saved to backup.json, next to theme.json. Backups are off
// while Dir is empty.
type Config struct {
	Dir     string `json:"dir"`
	Keep    int    `json:"keep,omitempty"`
	Profile string `json:"profile,omitempty"`
}

// Enabled reports whether a backup folder is set.
func (c Config) Enabled() bool { return c.Dir != "" }

// KeepN is how many copies of this profile are kept.
func (c Config) KeepN() int {
	if c.Keep < 1 {
		return DefaultKeep
	}
	return c.Keep
}

// ProfileName is the configured profile, or this machine's hostname, so
// each machine sharing a folder keeps its own copies.
func (c Config) ProfileName() string {
	if c.Profile != "" {
		return c.Profile
	}

	host, err := os.Hostname()
	if err != nil {
		return "default"
	}

	// keep only what is safe in a file name
	host = strings.Map(func(r rune) rune {
		if ValidProfile(string(r)) {
			return r
		}
		return '-'
	}, host)
	if host == "" {
		return "default"
	}
	return host
}

// LoadConfig reads backup.json. A missing file means backups are off.
func LoadConfig() (Config, error) {
	var cfg Config

	data, err := os.ReadFile(configFile())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return cfg, err
	}

	err = json.Unmarshal(data, &cfg)
	return cfg, err
}

// SaveConfig writes backup.json.
func SaveConfig(cfg Config) error {
	if err := os.MkdirAll(filepath.Dir(configFile()), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configFile(), data, 0644)
}

// configFile lives next to theme.json and keymap.json.
func configFile() string {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return filepath.Join(".akrasia", "backup.json")
		}
		configHome = filepath.Join(home, ".config")
	}
	return filepath.Join(configHome, "akrasia", "backup.json")
}
