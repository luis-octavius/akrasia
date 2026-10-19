package color

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

func Success() lipgloss.Style   { return styleFrom(GetCurrentTheme().Success) }
func Warning() lipgloss.Style   { return styleFrom(GetCurrentTheme().Warning) }
func Muted() lipgloss.Style     { return styleFrom(GetCurrentTheme().Muted) }
func Title() lipgloss.Style     { return styleFrom(GetCurrentTheme().Title) }
func Error() lipgloss.Style     { return styleFrom(GetCurrentTheme().Error) }
func Quote() lipgloss.Style     { return styleFrom(GetCurrentTheme().Quote) }
func Accent() lipgloss.Style    { return styleFrom(GetCurrentTheme().Accent) }
func Selection() lipgloss.Style { return styleFrom(GetCurrentTheme().Selection) }
func Bar() lipgloss.Style       { return styleFrom(GetCurrentTheme().Bar) }

// Priority styles a priority badge; unknown priorities are muted.
func Priority(priority string) lipgloss.Style {
	theme := GetCurrentTheme()
	switch priority {
	case "high":
		return styleFrom(theme.High)
	case "medium":
		return styleFrom(theme.Medium)
	case "low":
		return styleFrom(theme.Low)
	default:
		return Muted()
	}
}

// BorderColors returns the gradient stops for focused borders, falling
// back to the accent color.
func BorderColors() []color.Color {
	var colors []color.Color
	for _, name := range GetCurrentTheme().Border {
		if c, ok := colorValue(name); ok {
			colors = append(colors, c)
		}
	}
	if len(colors) == 0 {
		colors = append(colors, fgOrNone(GetCurrentTheme().Accent))
	}
	return colors
}

// BorderMutedColor returns the color of borders without focus, falling back
// to the muted text color.
func BorderMutedColor() color.Color {
	if c, ok := colorValue(GetCurrentTheme().BorderMuted); ok {
		return c
	}
	return fgOrNone(GetCurrentTheme().Muted)
}

// AccentColor returns the accent foreground color, for borders and cursors.
func AccentColor() color.Color {
	return fgOrNone(GetCurrentTheme().Accent)
}

func fgOrNone(s ColorStyle) color.Color {
	if c, ok := colorValue(s.Fg); ok {
		return c
	}
	return lipgloss.NoColor{}
}

func MsgSuccess(text string) { lipgloss.Println(Success().Render(text)) }
func MsgError(text string)   { lipgloss.Println(Error().Render(text)) }
func MsgMuted(text string)   { lipgloss.Println(Muted().Render(text)) }
func MsgWarning(text string) { lipgloss.Println(Warning().Render(text)) }
func MsgTitle(text string)   { lipgloss.Println(Title().Render(text)) }
func MsgQuote(text string)   { lipgloss.Println(Quote().Render(text)) }
