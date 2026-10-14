package color

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
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

func colorValue(name string) (lipgloss.Color, bool) {
	if c, ok := ansiCodes[name]; ok {
		return lipgloss.Color(c), true
	}

	if strings.HasPrefix(name, "#") {
		return lipgloss.Color(name), true
	}

	return "", false
}

// styleFrom creates a lipgloss Style based on 
// a ColorStyle struct
func styleFrom(s ColorStyle) lipgloss.Style {
	st := lipgloss.NewStyle()
	if c, ok := colorValue(s.Fg); ok {
		st = st.Foreground(lipgloss.Color(c))
	}
	if c, ok := colorValue(s.Bg); ok {
		st = st.Background(lipgloss.Color(c))
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

// getDefaultTheme returns the default theme configuration
func getDefaultTheme() *Theme {
	return &Theme{
		Name: "default",
		Error: ColorStyle{
			Fg: Red,
		},
		Warning: ColorStyle{
			Fg: Yellow,
		},
		Success: ColorStyle{
			Fg: Blue,
		},
		Quote: ColorStyle{
			Bg:         White,
			Fg:         Black,
			Attributes: []string{Underline},
		},
		Muted: ColorStyle{
			Fg: HiBlack,
		},
		Title: ColorStyle{
			Attributes: []string{Bold},
		},
	}
}

// getHighContrastTheme returns a high-contrast accessibility-friendly theme
func getHighContrastTheme() *Theme {
	return &Theme{
		Name: "high-contrast",
		Error: ColorStyle{
			Fg:         HiRed,
			Attributes: []string{Bold},
		},
		Warning: ColorStyle{
			Fg:         HiYellow,
			Attributes: []string{Bold},
		},
		Success: ColorStyle{
			Fg:         HiCyan,
			Attributes: []string{Bold},
		},
		Quote: ColorStyle{
			Bg:         Black,
			Fg:         HiWhite,
			Attributes: []string{Bold, Underline},
		},
		Muted: ColorStyle{
			Fg: White,
		},
		Title: ColorStyle{
			Fg:         HiWhite,
			Attributes: []string{Bold},
		},
	}
}

func getPurpleTheme() *Theme {
	return &Theme{
		Name: "purple",
		Error: ColorStyle{
			Fg:         HiRed,
			Attributes: []string{Bold},
		},
		Warning: ColorStyle{
			Fg:         HiYellow,
			Attributes: []string{Bold},
		},
		Success: ColorStyle{
			Fg:         "#a78bfa",
			Attributes: []string{Bold},
		},
		Quote: ColorStyle{
			Bg:         "#4c1d95",
			Fg:         "#ede9fe",
			Attributes: []string{Italic},
		},
		Muted: ColorStyle{
			Fg: "#7c6f9f",
		},
		Title: ColorStyle{
			Fg:         HiMagenta,
			Attributes: []string{Bold},
		},
	}
}
// GetAvailableThemes returns a list of all available themes
func GetAvailableThemes() []string {
	return []string{"default", "high-contrast", "purple"}
}

// GetThemeByName returns a theme by its name
func GetThemeByName(name string) *Theme {
	switch name {
	case "default":
		return getDefaultTheme()
	case "high-contrast":
		return getHighContrastTheme()
	case "purple":
		return getPurpleTheme()
	default:
		return getDefaultTheme()
	}
}

// GetCurrentTheme returns the currently active theme
func GetCurrentTheme() *Theme {
	if currentTheme == nil {
		currentTheme = getDefaultTheme()
	}
	return currentTheme
}

// loadTheme loads the theme from config file
func loadTheme() error {
	themeFile := getThemeFile()

	data, err := os.ReadFile(themeFile)
	if err != nil {
		// File doesn't exist, use default
		return err
	}

	var theme Theme
	if err := json.Unmarshal(data, &theme); err != nil {
		return err
	}

	def := getDefaultTheme()
	if isEmpty(theme.Muted) {
		theme.Muted = def.Muted
	}

	if isEmpty(theme.Title) {
		theme.Title = def.Title
	}

	currentTheme = &theme
	return nil
}

func isEmpty(s ColorStyle) bool {
	return s.Fg == "" && s.Bg == "" && len(s.Attributes) == 0
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
