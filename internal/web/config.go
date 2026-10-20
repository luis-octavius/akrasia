package web

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// DefaultAddr only listens on this machine: `tailscale serve` publishes it
// to the user's devices over HTTPS.
const DefaultAddr = "127.0.0.1:8080"

// Config is saved to serve.json, next to notify.json, readable only by its
// owner: the token is what lets a device in.
type Config struct {
	Token string `json:"token"`
	Addr  string `json:"addr,omitempty"`
}

// ListenAddr is the configured address, or DefaultAddr.
func (c Config) ListenAddr() string {
	if c.Addr == "" {
		return DefaultAddr
	}
	return c.Addr
}

// LoadConfig reads serve.json, creating it with a new token the first time.
func LoadConfig() (Config, error) {
	var cfg Config

	data, err := os.ReadFile(configFile())
	switch {
	case errors.Is(err, os.ErrNotExist):
	case err != nil:
		return cfg, err
	default:
		if err := json.Unmarshal(data, &cfg); err != nil {
			return cfg, err
		}
	}

	if cfg.Token == "" {
		if cfg.Token, err = NewToken(); err != nil {
			return cfg, err
		}
		if err := SaveConfig(cfg); err != nil {
			return cfg, err
		}
	}
	return cfg, nil
}

// SaveConfig writes serve.json, readable only by its owner.
func SaveConfig(cfg Config) error {
	if err := os.MkdirAll(filepath.Dir(configFile()), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(configFile(), data, 0600); err != nil {
		return err
	}
	// WriteFile keeps the mode of a file that already exists
	return os.Chmod(configFile(), 0600)
}

// NewToken returns 128 random bits as hex: short enough to type on a
// phone once, far too long to guess.
func NewToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// configFile lives next to theme.json and notify.json.
func configFile() string {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return filepath.Join(".akrasia", "serve.json")
		}
		configHome = filepath.Join(home, ".config")
	}
	return filepath.Join(configHome, "akrasia", "serve.json")
}
