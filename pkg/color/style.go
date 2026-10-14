package color

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

func Success() lipgloss.Style { return styleFrom(GetCurrentTheme().Success) }
func Warning() lipgloss.Style { return styleFrom(GetCurrentTheme().Warning) }
func Muted() lipgloss.Style   { return styleFrom(GetCurrentTheme().Muted) }
func Title() lipgloss.Style   { return styleFrom(GetCurrentTheme().Title) }
func Error() lipgloss.Style   { return styleFrom(GetCurrentTheme().Error) }
func Quote() lipgloss.Style   { return styleFrom(GetCurrentTheme().Quote) }

func MsgSuccess(text string) { fmt.Println(Success().Render(text)) }
func MsgError(text string)   { fmt.Println(Error().Render(text)) }
func MsgMuted(text string)   { fmt.Println(Muted().Render(text)) }
func MsgWarning(text string) { fmt.Println(Warning().Render(text)) }
func MsgTitle(text string)   { fmt.Println(Title().Render(text)) }
func MsgQuote(text string)   { fmt.Println(Quote().Render(text)) }
