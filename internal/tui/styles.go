package tui

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/luis-octavius/akrasia/pkg/color"
)

// Every style is built from the current theme at render time, so switching
// themes with `t` restyles the whole screen on the next frame.

// panelStyle is a rounded box; focused panels get the theme's gradient.
func panelStyle(focused bool) lipgloss.Style {
	st := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1)
	if focused {
		return st.BorderForegroundBlend(color.BorderColors()...)
	}
	return st.BorderForeground(color.BorderMutedColor())
}

// modalStyle frames prompts and the form floating over the list.
func modalStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForegroundBlend(color.BorderColors()...).
		Padding(1, 2)
}

// tabStyle draws one tab as a small rounded box.
func tabStyle(active bool) lipgloss.Style {
	st := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1)
	if active {
		return st.BorderForegroundBlend(color.BorderColors()...).
			Inherit(color.Accent())
	}
	return st.BorderForeground(color.BorderMutedColor()).
		Inherit(color.Muted())
}

// badgeStyle is the "akrasia" label: the accent color reversed.
func badgeStyle() lipgloss.Style {
	return color.Accent().Reverse(true).Bold(true).Padding(0, 1)
}

// pillStyle shows the current mode at the left of the status bar.
func pillStyle() lipgloss.Style {
	return color.Accent().Reverse(true).Bold(true).Padding(0, 1)
}

func priorityBadge(priority string) string {
	if priority == "" {
		return ""
	}
	return color.Priority(priority).Bold(true).Padding(0, 1).Render(priority)
}

func helpStyles() help.Styles {
	key := color.Accent().UnsetBold()
	desc := color.Muted()
	sep := color.Muted()
	return help.Styles{
		Ellipsis:       sep,
		ShortKey:       key,
		ShortDesc:      desc,
		ShortSeparator: sep,
		FullKey:        key,
		FullDesc:       desc,
		FullSeparator:  sep,
	}
}

// themedInput returns a copy of in painted with the current theme.
func themedInput(in textinput.Model) textinput.Model {
	state := textinput.StyleState{
		Text:        lipgloss.NewStyle(),
		Placeholder: color.Muted(),
		Suggestion:  color.Muted(),
		Prompt:      color.Accent(),
	}
	in.SetStyles(textinput.Styles{
		Focused: state,
		Blurred: state,
		Cursor: textinput.CursorStyle{
			Color: color.AccentColor(),
			Shape: tea.CursorBlock,
			Blink: true,
		},
	})
	return in
}
