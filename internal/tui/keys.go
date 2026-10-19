package tui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/luis-octavius/akrasia/pkg/i18n"
)

// keyMap holds every binding the TUI reacts to. Each action can carry
// several keys, so an opt-in vim layout can add j/k on top of the arrows
// later without touching Update.
type keyMap struct {
	Up      key.Binding
	Down    key.Binding
	NextTab key.Binding
	PrevTab key.Binding
	Reload  key.Binding
	Quit    key.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
		Up: key.NewBinding(
			key.WithKeys("up"),
			key.WithHelp("↑", i18n.T("tui.help.up")),
		),
		Down: key.NewBinding(
			key.WithKeys("down"),
			key.WithHelp("↓", i18n.T("tui.help.down")),
		),
		NextTab: key.NewBinding(
			key.WithKeys("tab", "right"),
			key.WithHelp("tab/→", i18n.T("tui.help.nextTab")),
		),
		PrevTab: key.NewBinding(
			key.WithKeys("shift+tab", "left"),
			key.WithHelp("shift+tab/←", i18n.T("tui.help.prevTab")),
		),
		Reload: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", i18n.T("tui.help.reload")),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", i18n.T("tui.help.quit")),
		),
	}
}

// ShortHelp implements help.KeyMap.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.NextTab, k.Reload, k.Quit}
}

// FullHelp implements help.KeyMap.
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down},
		{k.NextTab, k.PrevTab},
		{k.Reload, k.Quit},
	}
}
