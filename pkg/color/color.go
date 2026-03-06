package color

import (
	"github.com/fatih/color"
)

func MsgError(text string) {
	c := color.New(color.FgRed)
	c.Println(text)
}

func MsgWarning(text string) {
	c := color.New(color.FgYellow)
	c.Println(text)
}

func MsgSuccess(text string) {
	c := color.New(color.FgBlue)
	c.Println(text)
}

func MsgQuote(text string) {
	c := color.New(color.BgWhite).Add(color.FgBlack, color.Underline)
	c.Println(text)
}
