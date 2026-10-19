package color

// Built-in themes. The first three use the terminal's own ANSI palette;
// the others use hex colors from well-known palettes, which Lip Gloss
// downsamples on terminals without true color.

// getDefaultTheme returns the default theme configuration
func getDefaultTheme() *Theme {
	return &Theme{
		Name:        "default",
		Error:       ColorStyle{Fg: Red},
		Warning:     ColorStyle{Fg: Yellow},
		Success:     ColorStyle{Fg: Blue},
		Quote:       ColorStyle{Bg: White, Fg: Black, Attributes: []string{Underline}},
		Muted:       ColorStyle{Fg: HiBlack},
		Title:       ColorStyle{Attributes: []string{Bold}},
		Accent:      ColorStyle{Fg: Blue, Attributes: []string{Bold}},
		Selection:   ColorStyle{Bg: HiBlack},
		Bar:         ColorStyle{Fg: Black, Bg: Blue},
		High:        ColorStyle{Fg: Black, Bg: Red},
		Medium:      ColorStyle{Fg: Black, Bg: Yellow},
		Low:         ColorStyle{Fg: Black, Bg: Cyan},
		Border:      []string{Blue, Cyan},
		BorderMuted: HiBlack,
	}
}

// getHighContrastTheme returns a high-contrast accessibility-friendly theme
func getHighContrastTheme() *Theme {
	return &Theme{
		Name:        "high-contrast",
		Error:       ColorStyle{Fg: HiRed, Attributes: []string{Bold}},
		Warning:     ColorStyle{Fg: HiYellow, Attributes: []string{Bold}},
		Success:     ColorStyle{Fg: HiCyan, Attributes: []string{Bold}},
		Quote:       ColorStyle{Bg: Black, Fg: HiWhite, Attributes: []string{Bold, Underline}},
		Muted:       ColorStyle{Fg: White},
		Title:       ColorStyle{Fg: HiWhite, Attributes: []string{Bold}},
		Accent:      ColorStyle{Fg: HiYellow, Attributes: []string{Bold}},
		Selection:   ColorStyle{Fg: Black, Bg: HiWhite},
		Bar:         ColorStyle{Fg: Black, Bg: HiYellow, Attributes: []string{Bold}},
		High:        ColorStyle{Fg: Black, Bg: HiRed, Attributes: []string{Bold}},
		Medium:      ColorStyle{Fg: Black, Bg: HiYellow, Attributes: []string{Bold}},
		Low:         ColorStyle{Fg: Black, Bg: HiCyan, Attributes: []string{Bold}},
		Border:      []string{HiWhite},
		BorderMuted: White,
	}
}

func getPurpleTheme() *Theme {
	return &Theme{
		Name:        "purple",
		Error:       ColorStyle{Fg: HiRed, Attributes: []string{Bold}},
		Warning:     ColorStyle{Fg: HiYellow, Attributes: []string{Bold}},
		Success:     ColorStyle{Fg: "#a78bfa", Attributes: []string{Bold}},
		Quote:       ColorStyle{Bg: "#4c1d95", Fg: "#ede9fe", Attributes: []string{Italic}},
		Muted:       ColorStyle{Fg: "#7c6f9f"},
		Title:       ColorStyle{Fg: HiMagenta, Attributes: []string{Bold}},
		Accent:      ColorStyle{Fg: "#c4b5fd", Attributes: []string{Bold}},
		Selection:   ColorStyle{Fg: "#ede9fe", Bg: "#3b1f74"},
		Bar:         ColorStyle{Fg: "#ede9fe", Bg: "#4c1d95"},
		High:        ColorStyle{Fg: "#1e1b2e", Bg: "#f472b6"},
		Medium:      ColorStyle{Fg: "#1e1b2e", Bg: "#fbbf24"},
		Low:         ColorStyle{Fg: "#1e1b2e", Bg: "#818cf8"},
		Border:      []string{"#a78bfa", "#f0abfc", "#818cf8"},
		BorderMuted: "#4c3a75",
	}
}

// Rosé Pine (main variant), https://rosepinetheme.com
func getRosePineTheme() *Theme {
	const (
		base    = "#191724"
		overlay = "#26233a"
		muted   = "#6e6a86"
		text    = "#e0def4"
		love    = "#eb6f92"
		gold    = "#f6c177"
		rose    = "#ebbcba"
		foam    = "#9ccfd8"
		iris    = "#c4a7e7"
		hlMed   = "#403d52"
		hlHigh  = "#524f67"
	)
	return &Theme{
		Name:        "rose-pine",
		Error:       ColorStyle{Fg: love, Attributes: []string{Bold}},
		Warning:     ColorStyle{Fg: gold},
		Success:     ColorStyle{Fg: foam},
		Quote:       ColorStyle{Fg: text, Bg: overlay, Attributes: []string{Italic}},
		Muted:       ColorStyle{Fg: muted},
		Title:       ColorStyle{Fg: rose, Attributes: []string{Bold}},
		Accent:      ColorStyle{Fg: iris, Attributes: []string{Bold}},
		Selection:   ColorStyle{Fg: text, Bg: hlMed},
		Bar:         ColorStyle{Fg: text, Bg: overlay},
		High:        ColorStyle{Fg: base, Bg: love},
		Medium:      ColorStyle{Fg: base, Bg: gold},
		Low:         ColorStyle{Fg: base, Bg: foam},
		Border:      []string{iris, rose, foam},
		BorderMuted: hlHigh,
	}
}

// Neon pink, purple and cyan on deep violet.
func getCyberpunkTheme() *Theme {
	const (
		base    = "#0d0221"
		surface = "#1a0b2e"
		select_ = "#2a0a4a"
		muted   = "#7a6f9b"
		text    = "#d1f7ff"
		pink    = "#ff2a6d"
		cyan    = "#05d9e8"
		yellow  = "#f9f002"
		purple  = "#d300c5"
		dim     = "#3d2a5c"
	)
	return &Theme{
		Name:        "cyberpunk",
		Error:       ColorStyle{Fg: pink, Attributes: []string{Bold}},
		Warning:     ColorStyle{Fg: yellow},
		Success:     ColorStyle{Fg: cyan},
		Quote:       ColorStyle{Fg: text, Bg: select_, Attributes: []string{Italic}},
		Muted:       ColorStyle{Fg: muted},
		Title:       ColorStyle{Fg: cyan, Attributes: []string{Bold}},
		Accent:      ColorStyle{Fg: pink, Attributes: []string{Bold}},
		Selection:   ColorStyle{Fg: text, Bg: select_},
		Bar:         ColorStyle{Fg: text, Bg: surface},
		High:        ColorStyle{Fg: base, Bg: pink},
		Medium:      ColorStyle{Fg: base, Bg: yellow},
		Low:         ColorStyle{Fg: base, Bg: cyan},
		Border:      []string{pink, purple, cyan},
		BorderMuted: dim,
	}
}

// Catppuccin Mocha, https://catppuccin.com
func getCatppuccinMochaTheme() *Theme {
	const (
		base     = "#1e1e2e"
		surface0 = "#313244"
		surface1 = "#45475a"
		overlay0 = "#6c7086"
		text     = "#cdd6f4"
		mauve    = "#cba6f7"
		pink     = "#f5c2e7"
		blue     = "#89b4fa"
		lavender = "#b4befe"
		green    = "#a6e3a1"
		yellow   = "#f9e2af"
		red      = "#f38ba8"
		peach    = "#fab387"
	)
	return &Theme{
		Name:        "catppuccin-mocha",
		Error:       ColorStyle{Fg: red, Attributes: []string{Bold}},
		Warning:     ColorStyle{Fg: peach},
		Success:     ColorStyle{Fg: green},
		Quote:       ColorStyle{Fg: text, Bg: surface0, Attributes: []string{Italic}},
		Muted:       ColorStyle{Fg: overlay0},
		Title:       ColorStyle{Fg: lavender, Attributes: []string{Bold}},
		Accent:      ColorStyle{Fg: mauve, Attributes: []string{Bold}},
		Selection:   ColorStyle{Fg: text, Bg: surface1},
		Bar:         ColorStyle{Fg: text, Bg: surface0},
		High:        ColorStyle{Fg: base, Bg: red},
		Medium:      ColorStyle{Fg: base, Bg: yellow},
		Low:         ColorStyle{Fg: base, Bg: blue},
		Border:      []string{mauve, pink, blue},
		BorderMuted: surface1,
	}
}

// Tokyo Night, https://github.com/folke/tokyonight.nvim
func getTokyoNightTheme() *Theme {
	const (
		base      = "#1a1b26"
		storm     = "#24283b"
		highlight = "#292e42"
		termBlack = "#414868"
		comment   = "#565f89"
		fg        = "#c0caf5"
		blue      = "#7aa2f7"
		cyan      = "#7dcfff"
		purple    = "#bb9af7"
		green     = "#9ece6a"
		yellow    = "#e0af68"
		red       = "#f7768e"
	)
	return &Theme{
		Name:        "tokyo-night",
		Error:       ColorStyle{Fg: red, Attributes: []string{Bold}},
		Warning:     ColorStyle{Fg: yellow},
		Success:     ColorStyle{Fg: green},
		Quote:       ColorStyle{Fg: fg, Bg: highlight, Attributes: []string{Italic}},
		Muted:       ColorStyle{Fg: comment},
		Title:       ColorStyle{Fg: blue, Attributes: []string{Bold}},
		Accent:      ColorStyle{Fg: purple, Attributes: []string{Bold}},
		Selection:   ColorStyle{Fg: fg, Bg: highlight},
		Bar:         ColorStyle{Fg: fg, Bg: storm},
		High:        ColorStyle{Fg: base, Bg: red},
		Medium:      ColorStyle{Fg: base, Bg: yellow},
		Low:         ColorStyle{Fg: base, Bg: cyan},
		Border:      []string{blue, purple, cyan},
		BorderMuted: termBlack,
	}
}

// Gruvbox dark, https://github.com/morhetz/gruvbox
func getGruvboxTheme() *Theme {
	const (
		bg     = "#282828"
		bg1    = "#3c3836"
		bg2    = "#504945"
		bg3    = "#665c54"
		fg     = "#ebdbb2"
		gray   = "#928374"
		red    = "#fb4934"
		green  = "#b8bb26"
		yellow = "#fabd2f"
		blue   = "#83a598"
		aqua   = "#8ec07c"
		orange = "#fe8019"
	)
	return &Theme{
		Name:        "gruvbox",
		Error:       ColorStyle{Fg: red, Attributes: []string{Bold}},
		Warning:     ColorStyle{Fg: yellow},
		Success:     ColorStyle{Fg: green},
		Quote:       ColorStyle{Fg: fg, Bg: bg1, Attributes: []string{Italic}},
		Muted:       ColorStyle{Fg: gray},
		Title:       ColorStyle{Fg: yellow, Attributes: []string{Bold}},
		Accent:      ColorStyle{Fg: orange, Attributes: []string{Bold}},
		Selection:   ColorStyle{Fg: fg, Bg: bg2},
		Bar:         ColorStyle{Fg: fg, Bg: bg1},
		High:        ColorStyle{Fg: bg, Bg: red},
		Medium:      ColorStyle{Fg: bg, Bg: yellow},
		Low:         ColorStyle{Fg: bg, Bg: blue},
		Border:      []string{orange, yellow, aqua},
		BorderMuted: bg3,
	}
}

// Nord, https://www.nordtheme.com
func getNordTheme() *Theme {
	const (
		night0 = "#2e3440"
		night1 = "#3b4252"
		night2 = "#434c5e"
		night3 = "#4c566a"
		muted  = "#616e88"
		snow   = "#eceff4"
		frost0 = "#8fbcbb"
		frost1 = "#88c0d0"
		frost2 = "#81a1c1"
		frost3 = "#5e81ac"
		red    = "#bf616a"
		orange = "#d08770"
		yellow = "#ebcb8b"
		green  = "#a3be8c"
	)
	return &Theme{
		Name:        "nord",
		Error:       ColorStyle{Fg: red, Attributes: []string{Bold}},
		Warning:     ColorStyle{Fg: orange},
		Success:     ColorStyle{Fg: green},
		Quote:       ColorStyle{Fg: snow, Bg: night1, Attributes: []string{Italic}},
		Muted:       ColorStyle{Fg: muted},
		Title:       ColorStyle{Fg: frost1, Attributes: []string{Bold}},
		Accent:      ColorStyle{Fg: frost1, Attributes: []string{Bold}},
		Selection:   ColorStyle{Fg: snow, Bg: night2},
		Bar:         ColorStyle{Fg: snow, Bg: night1},
		High:        ColorStyle{Fg: night0, Bg: red},
		Medium:      ColorStyle{Fg: night0, Bg: yellow},
		Low:         ColorStyle{Fg: night0, Bg: frost2},
		Border:      []string{frost0, frost1, frost3},
		BorderMuted: night3,
	}
}
