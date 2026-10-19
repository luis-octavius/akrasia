package color

import (
	"charm.land/bubbles/v2/progress"
)

// ProgressBar draws a static bar of the given width (percentage included)
// filled with the theme's border gradient, for both the CLI and the TUI.
func ProgressBar(ratio float64, width int) string {
	bar := progress.New(
		progress.WithColors(BorderColors()...),
		progress.WithWidth(width),
	)
	bar.EmptyColor = BorderMutedColor()
	bar.PercentageStyle = Muted()
	return bar.ViewAs(ratio)
}
