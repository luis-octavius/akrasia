package color

import (
	"encoding/json"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strings"

	"charm.land/lipgloss/v2"
)

// Theme defines the color configuration for the application
type Theme struct {
	Name    string     `json:"name"`
	Error   ColorStyle `json:"error"`
	Warning ColorStyle `json:"warning"`
	Success ColorStyle `json:"success"`
	Quote   ColorStyle `json:"quote"`
	Muted   ColorStyle `json:"muted"`
	Title   ColorStyle `json:"title"`

	// Accent marks what has focus: the active tab, the cursor, the app badge.
	Accent ColorStyle `json:"accent"`
	// Selection is the row under the cursor (usually only a background).
	Selection ColorStyle `json:"selection"`
	// Bar is the status bar at the bottom of the TUI.
	Bar ColorStyle `json:"bar"`
	// High, Medium and Low style the priority badges.
	High   ColorStyle `json:"high"`
	Medium ColorStyle `json:"medium"`
	Low    ColorStyle `json:"low"`
	// Border holds the gradient stops of focused borders; one stop draws a
	// solid border.
	Border []string `json:"border"`
	// BorderMuted colors borders without focus.
	BorderMuted string `json:"borderMuted"`
}

// ColorStyle defines foreground and background colors with attributes
type ColorStyle struct {
	Fg         string   `json:"fg"`
	Bg         string   `json:"bg"`
	Attributes []string `json:"attributes"`
}

// Available color names
const (
	Black     = "black"
	Red       = "red"
	Green     = "green"
	Yellow    = "yellow"
	Blue      = "blue"
	Magenta   = "magenta"
	Cyan      = "cyan"
	White     = "white"
	HiRed     = "hi_red"
	HiGreen   = "hi_green"
	HiYellow  = "hi_yellow"
	HiBlue    = "hi_blue"
	HiMagenta = "hi_magenta"
	HiCyan    = "hi_cyan"
	HiWhite   = "hi_white"
	HiBlack   = "hi_black"
)

// Available attributes
const (
	Bold       = "bold"
	Italic     = "italic"
	Underline  = "underline"
	CrossOut   = "crossed_out"
	BlinkSlow  = "blink_slow"
	BlinkRapid = "blink_rapid"
)

var (
	currentTheme *Theme
	configDir    string
)

var ansiCodes = map[string]string{
	Black: "0", Red: "1", Green: "2", Yellow: "3",
	Blue: "4", Magenta: "5", Cyan: "6", White: "7",
	HiBlack: "8", HiRed: "9", HiGreen: "10", HiYellow: "11",
	HiBlue: "12", HiMagenta: "13", HiCyan: "14", HiWhite: "15",
}

// colorValue resolves an ANSI color name or a "#rrggbb" hex color.
func colorValue(name string) (color.Color, bool) {
	if c, ok := ansiCodes[name]; ok {
		return lipgloss.Color(c), true
	}

	if strings.HasPrefix(name, "#") {
		return lipgloss.Color(name), true
	}

	return nil, false
}

// styleFrom creates a lipgloss Style based on
// a ColorStyle struct
func styleFrom(s ColorStyle) lipgloss.Style {
	st := lipgloss.NewStyle()
	if c, ok := colorValue(s.Fg); ok {
		st = st.Foreground(c)
	}
	if c, ok := colorValue(s.Bg); ok {
		st = st.Background(c)
	}

	for _, a := range s.Attributes {
		switch a {
		case Bold:
			st = st.Bold(true)
		case Italic:
			st = st.Italic(true)
		case Underline:
			st = st.Underline(true)
		case CrossOut:
			st = st.Strikethrough(true)
		case BlinkSlow, BlinkRapid:
			st = st.Blink(true)
		}
	}

	return st
}

func init() {
	// Initialize config directory
	configDir = getConfigDir()
	// Load theme on package initialization
	if err := loadTheme(); err != nil {
		// If loading fails, use default theme
		currentTheme = getDefaultTheme()
	}
}

// getConfigDir returns the akrasia config directory
func getConfigDir() string {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ".akrasia"
		}
		configHome = filepath.Join(home, ".config")
	}
	return filepath.Join(configHome, "akrasia")
}

// getThemeFile returns the path to the theme config file
func getThemeFile() string {
	return filepath.Join(configDir, "theme.json")
}

// builtinThemes maps each theme name to its constructor; themeOrder is the
// order shown by `config theme list` and cycled by the TUI.
var builtinThemes = map[string]func() *Theme{
	"default":          getDefaultTheme,
	"high-contrast":    getHighContrastTheme,
	"purple":           getPurpleTheme,
	"rose-pine":        getRosePineTheme,
	"cyberpunk":        getCyberpunkTheme,
	"catppuccin-mocha": getCatppuccinMochaTheme,
	"tokyo-night":      getTokyoNightTheme,
	"gruvbox":          getGruvboxTheme,
	"nord":             getNordTheme,
}

var themeOrder = []string{
	"default", "high-contrast", "purple",
	"rose-pine", "cyberpunk", "catppuccin-mocha", "tokyo-night", "gruvbox", "nord",
}

// GetAvailableThemes returns a list of all available themes
func GetAvailableThemes() []string {
	return append([]string(nil), themeOrder...)
}

// GetThemeByName returns a theme by its name, or the default theme when the
// name is unknown.
func GetThemeByName(name string) *Theme {
	if build, ok := builtinThemes[name]; ok {
		return build()
	}
	return getDefaultTheme()
}

// NextTheme returns the theme after current in the list, wrapping around.
// An unknown name starts from the first theme.
func NextTheme(current string) string {
	for i, name := range themeOrder {
		if name == current {
			return themeOrder[(i+1)%len(themeOrder)]
		}
	}
	return themeOrder[0]
}

// GetCurrentTheme returns the currently active theme
func GetCurrentTheme() *Theme {
	if currentTheme == nil {
		currentTheme = getDefaultTheme()
	}
	return currentTheme
}

// loadTheme loads the theme from config file. Fields missing from the file
// (an older or hand-edited theme.json) are taken from the built-in theme of
// the same name, or from the default one.
func loadTheme() error {
	data, err := os.ReadFile(getThemeFile())
	if err != nil {
		// File doesn't exist, use default
		return err
	}

	var theme Theme
	if err := json.Unmarshal(data, &theme); err != nil {
		return err
	}

	fillMissing(&theme, GetThemeByName(theme.Name))
	currentTheme = &theme
	return nil
}

// fillMissing copies every role that t leaves empty from base.
func fillMissing(t, base *Theme) {
	styles := []struct{ dst, src *ColorStyle }{
		{&t.Error, &base.Error}, {&t.Warning, &base.Warning},
		{&t.Success, &base.Success}, {&t.Quote, &base.Quote},
		{&t.Muted, &base.Muted}, {&t.Title, &base.Title},
		{&t.Accent, &base.Accent}, {&t.Selection, &base.Selection},
		{&t.Bar, &base.Bar}, {&t.High, &base.High},
		{&t.Medium, &base.Medium}, {&t.Low, &base.Low},
	}
	for _, s := range styles {
		if isEmpty(*s.dst) {
			*s.dst = *s.src
		}
	}

	if len(t.Border) == 0 {
		t.Border = base.Border
	}
	if t.BorderMuted == "" {
		t.BorderMuted = base.BorderMuted
	}
}

func isEmpty(s ColorStyle) bool {
	return s.Fg == "" && s.Bg == "" && len(s.Attributes) == 0
}

// SetTheme switches the active theme for this process without saving it.
func SetTheme(themeName string) {
	currentTheme = GetThemeByName(themeName)
}

// SaveTheme saves the theme to config file
func SaveTheme(themeName string) error {
	theme := GetThemeByName(themeName)

	// Ensure config directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(theme, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal theme: %w", err)
	}

	themeFile := getThemeFile()
	if err := os.WriteFile(themeFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write theme file: %w", err)
	}

	currentTheme = theme
	return nil
}
