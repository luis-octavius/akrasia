package tasks

import (
	"fmt"
	"math/rand"
	"os"
	"strings"

	"github.com/luis-octavius/akrasia/pkg/color"
	"golang.org/x/term"
)

var quotes = []string{
	`We suffer more often in imagination than in reality. And yet, we put off what matters to the calm, indefinite future, as if time were ours to command." - Seneca, Moral Letters to Lucilius (Letter 13).`,
	`"The present is the only time that is truly ours. To allow the mind to wander in past regret or future anxiety is the greatest theft of the self." - Marcus Aurelius, Meditations (Adapted from Book 3, 10; Book 8, 36).`,
	`"One must not think that a task is easy simply because it is not urgent. It is the non-urgent tasks that most often define the quality of a life." - Inspired by Epictetus's doctrine of prohairesis (moral character).`,
	`"The soul is dyed by the color of its thoughts. Guard, therefore, your thinking, for it becomes your action. Guard your action, for it becomes your habit." - Marcus Aurelius, Meditations (Book 5, 16).`,
	`"We must keep constant guard over our perceptions, for they are the gatekeepers of the mind." - Epictetus, Discourses (Book 4, 12).`,
	`"The mind distracted is the mind diminished. It is not a single flame but scattered embers, giving neither light nor heat." - Inspired by Plutarch's "On the Control of Anger" and the Stoic doctrine of unity of mind.`,
	`"To postpone is to choose. When you defer a good action, you have actively chosen the worse state of your soul." - Inspired by Socrates in Plato's Gorgias.`,
	`"The cause of human inaction is not that the task is large, but that the will is small. Begin, and the mind will be heated; continue, and the task will be completed." - Adapted from Democritus (Fragment 84) and the Stoic principle of "begin."`,
	`"No one errs willingly; the failure to act rightly is a failure of knowledge, not of will." - Socrates, in Plato's Protagoras`,
	`"Self-control (enkrateia) is the foundation of virtue, and the first step is to master your pleasures and appetites." - Aristotle, Nicomachean Ethics`,
	`"I see and approve the better course, but I follow the worse." - Ovid, Metamorphoses`,
	`"The wise man is commanded by reason, the fool by passion, and the madman by whim." - Zeno of Citium`,
	`"How long will you wait before you demand the best for yourself?" - Epictetus, Discourses`,
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
