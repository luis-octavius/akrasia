package commands

import (
	"strings"
	"testing"
)

func TestWrapTextHasNoLeadingSpaces(t *testing.T) {
	text := strings.Repeat("word ", 30)

	for _, line := range strings.Split(wrapText(text, 44), "\n") {
		if strings.HasPrefix(line, " ") {
			t.Fatalf("line starts with a space: %q", line)
		}
		if len(line) > 40 {
			t.Fatalf("line longer than 40 chars: %q", line)
		}
	}
}
