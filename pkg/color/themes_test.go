package color

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestEveryThemeHasEveryRole(t *testing.T) {
	for _, name := range GetAvailableThemes() {
		theme := GetThemeByName(name)
		if theme.Name != name {
			t.Errorf("GetThemeByName(%q) returned theme %q", name, theme.Name)
		}

		roles := map[string]ColorStyle{
			"error": theme.Error, "warning": theme.Warning, "success": theme.Success,
			"quote": theme.Quote, "muted": theme.Muted, "title": theme.Title,
			"accent": theme.Accent, "selection": theme.Selection, "bar": theme.Bar,
			"high": theme.High, "medium": theme.Medium, "low": theme.Low,
		}
		for role, style := range roles {
			if isEmpty(style) {
				t.Errorf("%s: role %s is empty", name, role)
			}
			for _, c := range []string{style.Fg, style.Bg} {
				if _, ok := colorValue(c); c != "" && !ok {
					t.Errorf("%s: role %s has invalid color %q", name, role, c)
				}
			}
		}

		if len(theme.Border) == 0 {
			t.Errorf("%s: no border colors", name)
		}
		for _, c := range append(theme.Border, theme.BorderMuted) {
			if _, ok := colorValue(c); !ok {
				t.Errorf("%s: invalid border color %q", name, c)
			}
		}
	}
}

func TestNextThemeWrapsAround(t *testing.T) {
	themes := GetAvailableThemes()

	seen := map[string]bool{}
	name := themes[0]
	for range themes {
		seen[name] = true
		name = NextTheme(name)
	}
	if name != themes[0] || len(seen) != len(themes) {
		t.Fatalf("NextTheme should visit all %d themes and come back, saw %d and ended at %q", len(themes), len(seen), name)
	}

	if NextTheme("unknown") != themes[0] {
		t.Errorf("unknown theme should restart at %q", themes[0])
	}
}

// A theme.json written before the TUI roles existed must still load, with
// the new roles taken from the built-in theme of the same name.
func TestLoadOldThemeFileFillsNewRoles(t *testing.T) {
	dir := t.TempDir()
	oldConfigDir, oldTheme := configDir, currentTheme
	configDir = dir
	t.Cleanup(func() { configDir, currentTheme = oldConfigDir, oldTheme })

	old := map[string]any{
		"name":    "nord",
		"error":   map[string]any{"fg": "#ff0000"},
		"warning": map[string]any{"fg": "yellow"},
		"success": map[string]any{"fg": "green"},
		"quote":   map[string]any{"fg": "white"},
	}
	data, _ := json.Marshal(old)
	if err := os.WriteFile(filepath.Join(dir, "theme.json"), data, 0644); err != nil {
		t.Fatal(err)
	}

	if err := loadTheme(); err != nil {
		t.Fatalf("loadTheme() error = %v", err)
	}

	got, nord := GetCurrentTheme(), GetThemeByName("nord")
	if got.Error.Fg != "#ff0000" {
		t.Errorf("custom error color should be kept, got %q", got.Error.Fg)
	}
	if got.Accent.Fg != nord.Accent.Fg || got.BorderMuted != nord.BorderMuted || len(got.Border) != len(nord.Border) {
		t.Errorf("missing roles should come from nord, got accent %+v border %v", got.Accent, got.Border)
	}
	if isEmpty(got.Muted) || isEmpty(got.Title) {
		t.Errorf("muted and title should be filled too")
	}
}
