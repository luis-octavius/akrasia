package tasks

import (
	"fmt"
	"math/rand"
	"os"
	"strings"

	"github.com/luis-octavius/akrasia/pkg/color"
	"golang.org/x/term"
	"github.com/luis-octavius/akrasia/pkg/i18n"
)

var quotes = []string{
	i18n.T("quote1"),
	i18n.T("quote2"),
	i18n.T("quote3"),
	i18n.T("quote4"),
	i18n.T("quote5"),
	i18n.T("quote6"),
	i18n.T("quote7"),
	i18n.T("quote8"),
	i18n.T("quote9"),
	i18n.T("quote10"),
	i18n.T("quote11"),
	i18n.T("quote12"),
	i18n.T("quote13"),
}

// getTerminalWidth detects the current terminal width, with a safe minimum
func getTerminalWidth() int {
	if width, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && width > 0 {
		return width
	}
	// Default fallback width
	return 80
}

// wrapText wraps text to fit within terminal width with proper word boundaries
func wrapText(text string, maxWidth int) string {
	// Leave some margin for padding/formatting
	effectiveWidth := maxWidth - 4
	if effectiveWidth < 40 {
		effectiveWidth = 40
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return text
	}

	var lines []string
	var currentLine strings.Builder

	for _, word := range words {
		wordLen := len(word)
		currentLen := currentLine.Len()

		// If adding word would exceed width, start new line
		if currentLen > 0 && currentLen+1+wordLen > effectiveWidth {
			lines = append(lines, currentLine.String())
			currentLine.Reset()
		}

		// Add space if not first word on line
		if currentLen > 0 {
			currentLine.WriteString(" ")
		}

		currentLine.WriteString(word)
	}

	// Add remaining line
	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}

	return strings.Join(lines, "\n")
}

func generateRandomQuote() {
	quote := quotes[rand.Intn(len(quotes))]
	wrappedQuote := wrapText(quote, getTerminalWidth())
	formattedQuote := fmt.Sprintf("\n%s\n", wrappedQuote)
	color.MsgQuote(formattedQuote)
}
