package color

func MsgError(text string) {
	theme := GetCurrentTheme()
	c := applyColorStyle(theme.Error)
	c.Println(text)
}

func MsgWarning(text string) {
	theme := GetCurrentTheme()
	c := applyColorStyle(theme.Warning)
	c.Println(text)
}

func MsgSuccess(text string) {
	theme := GetCurrentTheme()
	c := applyColorStyle(theme.Success)
	c.Println(text)
}

func MsgQuote(text string) {
	theme := GetCurrentTheme()
	c := applyColorStyle(theme.Quote)
	c.Println(text)
}
