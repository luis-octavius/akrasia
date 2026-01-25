package color

import (
	"fmt"

	"github.com/fatih/color"
)

var colors = map[string]*color.Color{
	"red":   color.New(color.FgRed),
	"green": color.New(color.FgGreen),
	"blue":  color.New(color.FgBlue),
	"cyan":  color.New(color.FgCyan),
}

func ColorizeOutput(colorName, text string) (string, error) {
	color, ok := colors[colorName]
	if !ok {
		return "", fmt.Errorf("Color does not exist on map")
	}

	colorizedOutput := color.Sprint(text)
	return colorizedOutput, nil
}
