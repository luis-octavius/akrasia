package color

import (
	"charm.land/lipgloss/v2"
)

func Success() lipgloss.Style { return styleFrom(GetCurrentTheme().Success) }
func Warning() lipgloss.Style { return styleFrom(GetCurrentTheme().Warning) }
func Muted() lipgloss.Style   { return styleFrom(GetCurrentTheme().Muted) }
func Title() lipgloss.Style   { return styleFrom(GetCurrentTheme().Title) }
func Error() lipgloss.Style   { return styleFrom(GetCurrentTheme().Error) }
func Quote() lipgloss.Style   { return styleFrom(GetCurrentTheme().Quote) }

func MsgSuccess(text string) { lipgloss.Println(Success().Render(text)) }
func MsgError(text string)   { lipgloss.Println(Error().Render(text)) }
func MsgMuted(text string)   { lipgloss.Println(Muted().Render(text)) }
func MsgWarning(text string) { lipgloss.Println(Warning().Render(text)) }
func MsgTitle(text string)   { lipgloss.Println(Title().Render(text)) }
func MsgQuote(text string)   { lipgloss.Println(Quote().Render(text)) }
