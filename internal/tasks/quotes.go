package tasks

import (
	"math/rand"

	"github.com/luis-octavius/akrasia/pkg/i18n"
)

var quotes = []string{
	i18n.T("quotes.quote1"),
	i18n.T("quotes.quote2"),
	i18n.T("quotes.quote3"),
	i18n.T("quotes.quote4"),
	i18n.T("quotes.quote5"),
	i18n.T("quotes.quote6"),
	i18n.T("quotes.quote7"),
	i18n.T("quotes.quote8"),
	i18n.T("quotes.quote9"),
	i18n.T("quotes.quote10"),
	i18n.T("quotes.quote11"),
	i18n.T("quotes.quote12"),
	i18n.T("quotes.quote13"),
}

// RandomQuote returns one of the motivational quotes, unwrapped.
func RandomQuote() string {
	return quotes[rand.Intn(len(quotes))]
}
