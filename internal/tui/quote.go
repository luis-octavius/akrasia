package tui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/luis-octavius/akrasia/internal/tasks"
	"github.com/luis-octavius/akrasia/pkg/color"
)

// openQuote shows a random quote in a pop-up. title names what was just
// done ("✔ X completed"); empty when the quote was asked for with Q.
func (m *Model) openQuote(title string) {
	m.mode, m.quote, m.quoteTitle = modeQuote, tasks.RandomQuote(), title
}

// handleQuoteKey closes the pop-up with esc, enter or q; other keys are
// ignored so a stray key doesn't act on the list behind it.
func (m Model) handleQuoteKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "enter", "q":
		m.mode = modeNormal
	}
	return m, nil
}

// renderQuoteModal wraps the quote to the pop-up width, under the title of
// what was completed.
func (m Model) renderQuoteModal(box lipgloss.Style, width int) string {
	quote := color.Quote().Width(width - 6).Render(m.quote)
	if m.quoteTitle == "" {
		return box.Render(quote)
	}
	return box.Render(color.Success().Bold(true).Render(m.quoteTitle) + "\n\n" + quote)
}
