package color

func StyleError(text string) string   { return applyColorStyle(GetCurrentTheme().Error).Sprint(text) }
func StyleWarning(text string) string { return applyColorStyle(GetCurrentTheme().Warning).Sprint(text) }
func StyleSuccess(text string) string { return applyColorStyle(GetCurrentTheme().Success).Sprint(text) }
func StyleMuted(text string) string   { return applyColorStyle(GetCurrentTheme().Muted).Sprint(text) }
func StyleTitle(text string) string   { return applyColorStyle(GetCurrentTheme().Title).Sprint(text) }
