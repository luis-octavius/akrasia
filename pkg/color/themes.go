package color

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
)

// Theme defines the color configuration for the application
type Theme struct {
	Name    string     `json:"name"`
	Error   ColorStyle `json:"error"`
	Warning ColorStyle `json:"warning"`
	Success ColorStyle `json:"success"`
	Quote   ColorStyle `json:"quote"`
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
)

// Available attributes
const (
	Bold       = "bold"
	Italic     = "italic"
	Underline  = "underline"
	BlinkSlow  = "blink_slow"
	BlinkRapid = "blink_rapid"
	Concealed  = "concealed"
	CrossOut   = "crossed_out"
)

var (
	currentTheme *Theme
	configDir    string
)

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
	}
}

// GetAvailableThemes returns a list of all available themes
func GetAvailableThemes() []string {
	return []string{"default", "high-contrast"}
}

// GetThemeByName returns a theme by its name
func GetThemeByName(name string) *Theme {
	switch name {
	case "default":
		return getDefaultTheme()
	case "high-contrast":
		return getHighContrastTheme()
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

	currentTheme = &theme
	return nil
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

// colorNameToValue converts color name string to fatih/color constant
func colorNameToValue(name string) color.Attribute {
	switch name {
	case Red:
		return color.FgRed
	case Green:
		return color.FgGreen
	case Yellow:
		return color.FgYellow
	case Blue:
		return color.FgBlue
	case Magenta:
		return color.FgMagenta
	case Cyan:
		return color.FgCyan
	case White:
		return color.FgWhite
	case Black:
		return color.FgBlack
	case HiRed:
		return color.FgHiRed
	case HiGreen:
		return color.FgHiGreen
	case HiYellow:
		return color.FgHiYellow
	case HiBlue:
		return color.FgHiBlue
	case HiMagenta:
		return color.FgHiMagenta
	case HiCyan:
		return color.FgHiCyan
	case HiWhite:
		return color.FgHiWhite
	default:
		return color.FgWhite
	}
}

// bgColorNameToValue converts background color name string to fatih/color constant
func bgColorNameToValue(name string) color.Attribute {
	switch name {
	case Red:
		return color.BgRed
	case Green:
		return color.BgGreen
	case Yellow:
		return color.BgYellow
	case Blue:
		return color.BgBlue
	case Magenta:
		return color.BgMagenta
	case Cyan:
		return color.BgCyan
	case White:
		return color.BgWhite
	case Black:
		return color.BgBlack
	case HiRed:
		return color.BgHiRed
	case HiGreen:
		return color.BgHiGreen
	case HiYellow:
		return color.BgHiYellow
	case HiBlue:
		return color.BgHiBlue
	case HiMagenta:
		return color.BgHiMagenta
	case HiCyan:
		return color.BgHiCyan
	case HiWhite:
		return color.BgHiWhite
	default:
		return color.BgBlack
	}
}

// attributeNameToValue converts attribute name string to fatih/color constant
func attributeNameToValue(name string) color.Attribute {
	switch name {
	case Bold:
		return color.Bold
	case Italic:
		return color.Italic
	case Underline:
		return color.Underline
	case BlinkSlow:
		return color.BlinkSlow
	case BlinkRapid:
		return color.BlinkRapid
	case Concealed:
		return color.Concealed
	case CrossOut:
		return color.CrossedOut
	default:
		return color.Reset
	}
}

// applyColorStyle returns a color.Color with the specified style applied
func applyColorStyle(style ColorStyle) *color.Color {
	var attrs []color.Attribute

	// Add foreground color
	if style.Fg != "" {
		attrs = append(attrs, colorNameToValue(style.Fg))
	}

	// Add background color
	if style.Bg != "" {
		attrs = append(attrs, bgColorNameToValue(style.Bg))
	}

	// Add attributes
	for _, attr := range style.Attributes {
		attrs = append(attrs, attributeNameToValue(attr))
	}

	if len(attrs) == 0 {
		return color.New()
	}

	return color.New(attrs...)
}
