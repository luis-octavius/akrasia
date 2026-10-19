package tui

import (
	"charm.land/bubbles/v2/key"
	"github.com/luis-octavius/akrasia/pkg/i18n"
)

// keyMap holds every binding the TUI reacts to in normal mode. Each action
// can carry several keys, so the vim layout only adds keys on top of the
// default ones and Update never needs to know which layout is active.
type keyMap struct {
	Up       key.Binding
	Down     key.Binding
	Top      key.Binding
	Bottom   key.Binding
	NextTab  key.Binding
	PrevTab  key.Binding
	Complete key.Binding
	Add      key.Binding
	Delete   key.Binding
	Search   key.Binding
	Priority key.Binding
	Details  key.Binding
	Theme    key.Binding
	Reload   key.Binding
	Help     key.Binding
	Quit     key.Binding
}

func newKeyMap(layout string) keyMap {
	vim := layout == KeymapVim

	// keys returns the default keys, plus the vim ones when enabled.
	keys := func(def []string, vimKeys ...string) []string {
		if vim {
			return append(def, vimKeys...)
		}
		return def
	}
	// label shows the vim key first when enabled, since that is what a
	// vim user reaches for.
	label := func(def, vimKey string) string {
		if vim {
			return vimKey + "/" + def
		}
		return def
	}

	return keyMap{
		Up: key.NewBinding(
			key.WithKeys(keys([]string{"up"}, "k")...),
			key.WithHelp(label("↑", "k"), i18n.T("tui.help.up")),
		),
		Down: key.NewBinding(
			key.WithKeys(keys([]string{"down"}, "j")...),
			key.WithHelp(label("↓", "j"), i18n.T("tui.help.down")),
		),
		Top: key.NewBinding(
			key.WithKeys(keys([]string{"home"}, "g")...),
			key.WithHelp(label("home", "g"), i18n.T("tui.help.top")),
		),
		Bottom: key.NewBinding(
			key.WithKeys(keys([]string{"end"}, "G")...),
			key.WithHelp(label("end", "G"), i18n.T("tui.help.bottom")),
		),
		NextTab: key.NewBinding(
			key.WithKeys(keys([]string{"tab", "right"}, "l")...),
			key.WithHelp(label("tab/→", "l"), i18n.T("tui.help.nextTab")),
		),
		PrevTab: key.NewBinding(
			key.WithKeys(keys([]string{"shift+tab", "left"}, "h")...),
			key.WithHelp(label("shift+tab/←", "h"), i18n.T("tui.help.prevTab")),
		),
		Complete: key.NewBinding(
			key.WithKeys("space"),
			key.WithHelp("space", i18n.T("tui.help.complete")),
		),
		Add: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", i18n.T("tui.help.add")),
		),
		Delete: key.NewBinding(
			key.WithKeys("x", "delete"),
			key.WithHelp("x", i18n.T("tui.help.delete")),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", i18n.T("tui.help.search")),
		),
		Priority: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", i18n.T("tui.help.priority")),
		),
		Details: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", i18n.T("tui.help.details")),
		),
		Theme: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", i18n.T("tui.help.theme")),
		),
		Reload: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", i18n.T("tui.help.reload")),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", i18n.T("tui.help.help")),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", i18n.T("tui.help.quit")),
		),
	}
}

// ShortHelp implements help.KeyMap.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.NextTab, k.Complete, k.Add, k.Delete, k.Help, k.Quit}
}

// FullHelp implements help.KeyMap.
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Top, k.Bottom},
		{k.NextTab, k.PrevTab, k.Reload, k.Theme},
		{k.Complete, k.Add, k.Delete, k.Details},
		{k.Search, k.Priority, k.Help, k.Quit},
	}
}
