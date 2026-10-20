package tasks

import (
	"fmt"
	"math/rand"

	"github.com/luis-octavius/akrasia/pkg/i18n"
)

// quoteCount is how many quotes the locales hold (quotes.quote1 to
// quotes.quoteN).
const quoteCount = 13

// RandomQuote returns one of the motivational quotes, unwrapped. It is
// looked up when asked for, not when the package loads, so a missing
// translation can never stop the program from starting.
func RandomQuote() string {
	return i18n.T(fmt.Sprintf("quotes.quote%d", rand.Intn(quoteCount)+1))
}
