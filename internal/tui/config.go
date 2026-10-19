package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/luis-octavius/akrasia/internal/backup"
	"github.com/luis-octavius/akrasia/pkg/i18n"
)

const (
	KeymapDefault = "default"
	KeymapVim     = "vim"
)

type keymapConfig struct {
	Keymap string `json:"keymap"`
}

// Keymaps lists the layouts accepted by SetKeymap.
func Keymaps() []string {
	return []string{KeymapDefault, KeymapVim}
}

// CurrentKeymap returns the saved layout, or the default one when nothing
// valid is saved.
func CurrentKeymap() string {
	data, err := os.ReadFile(keymapFile())
	if err != nil {
		return KeymapDefault
	}

	var cfg keymapConfig
	if err := json.Unmarshal(data, &cfg); err != nil || !slices.Contains(Keymaps(), cfg.Keymap) {
		return KeymapDefault
	}

	return cfg.Keymap
}

// SetKeymap saves the layout the TUI starts with.
func SetKeymap(name string) error {
	if !slices.Contains(Keymaps(), name) {
		return fmt.Errorf("unknown keymap %q", name)
	}

	if err := os.MkdirAll(filepath.Dir(keymapFile()), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(keymapConfig{Keymap: name}, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(keymapFile(), data, 0644)
}

// keymapFile lives next to theme.json and lang.json.
func keymapFile() string {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return filepath.Join(".akrasia", "keymap.json")
		}
		configHome = filepath.Join(home, ".config")
	}
	return filepath.Join(configHome, "akrasia", "keymap.json")
}

// lastBackup describes this machine's newest backup for the status bar,
// and reports whether backups are set up at all.
func lastBackup(now time.Time) (string, bool) {
	cfg, err := backup.LoadConfig()
	if err != nil || !cfg.Enabled() {
		return i18n.T("tui.status.noBackup"), false
	}

	latest, ok, err := backup.Latest(cfg.Dir, cfg.ProfileName())
	switch {
	case err != nil || !ok:
		return fmt.Sprintf(i18n.T("tui.status.backup"), i18n.T("commands.info.backupNever")), true
	case latest.Time.Format(time.DateOnly) == now.Format(time.DateOnly):
		return i18n.T("tui.status.backupToday"), true
	default:
		return fmt.Sprintf(i18n.T("tui.status.backup"), latest.Time.Format(time.DateOnly)), true
	}
}
