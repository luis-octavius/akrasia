package i18n

import (
	"encoding/json"
	"path"
	"slices"
	"strings"
	"testing"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

func TestEveryLocaleLoads(t *testing.T) {
	if _, err := newBundle(); err != nil {
		t.Fatalf("loading the embedded locales failed: %v", err)
	}
	if err := LoadError(); err != nil {
		t.Fatalf("LoadError() = %v", err)
	}
}

// keys flattens a locale file into its dotted message ids.
func keys(t *testing.T, lang string) map[string]bool {
	t.Helper()
	data, err := localeFS.ReadFile(path.Join("locales", lang+".json"))
	if err != nil {
		t.Fatalf("reading %s: %v", lang, err)
	}
	var tree map[string]any
	if err := json.Unmarshal(data, &tree); err != nil {
		t.Fatalf("parsing %s: %v", lang, err)
	}

	out := map[string]bool{}
	var walk func(prefix string, node map[string]any)
	walk = func(prefix string, node map[string]any) {
		for k, v := range node {
			id := k
			if prefix != "" {
				id = prefix + "." + k
			}
			if child, ok := v.(map[string]any); ok {
				walk(id, child)
			} else {
				out[id] = true
			}
		}
	}
	walk("", tree)
	return out
}

func TestLocalesHaveTheSameKeys(t *testing.T) {
	en := keys(t, "en")
	for _, lang := range GetAvailableLanguages() {
		other := keys(t, lang)
		var missing, extra []string
		for k := range en {
			if !other[k] {
				missing = append(missing, k)
			}
		}
		for k := range other {
			if !en[k] {
				extra = append(extra, k)
			}
		}
		slices.Sort(missing)
		slices.Sort(extra)
		if len(missing) > 0 || len(extra) > 0 {
			t.Errorf("%s differs from en:\n  missing: %v\n  extra:   %v", lang, missing, extra)
		}
	}
}

// go-i18n reads these names inside a nested message as message fields, and
// then the whole file fails to load.
func TestNoReservedKeyNames(t *testing.T) {
	reserved := []string{"id", "description", "hash", "leftdelim", "rightdelim", "zero", "one", "two", "few", "many", "other"}
	for _, lang := range GetAvailableLanguages() {
		for id := range keys(t, lang) {
			for _, part := range strings.Split(id, ".") {
				if slices.Contains(reserved, strings.ToLower(part)) {
					t.Errorf("%s: %q uses the reserved name %q", lang, id, part)
				}
			}
		}
	}
}

func TestTNeverPanics(t *testing.T) {
	if got := T("no.such.message"); got != "no.such.message" {
		t.Errorf("an unknown id should come back as itself, got %q", got)
	}
	if got := T("tui.tabs.dump"); got == "" || got == "tui.tabs.dump" {
		t.Errorf("a known id should be translated, got %q", got)
	}
}

// With only English loaded, asking in Portuguese must still give the
// English text, not an error or the id.
func TestTFallsBackToEnglish(t *testing.T) {
	saved, savedEnglish := lcz, english
	t.Cleanup(func() { lcz, english = saved, savedEnglish })

	bnd := i18n.NewBundle(language.English)
	if _, err := bnd.LoadMessageFileFS(localeFS, "locales/en.json"); err != nil {
		t.Fatalf("loading en: %v", err)
	}
	lcz = i18n.NewLocalizer(bnd, "pt")
	english = i18n.NewLocalizer(bnd, "en")

	if got := T("tui.tabs.today"); got != "Today" {
		t.Errorf("T should fall back to English, got %q", got)
	}
}
