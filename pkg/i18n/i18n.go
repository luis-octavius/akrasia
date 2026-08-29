package i18n

import (
	"fmt"
	"os"
	"path/filepath"
	"encoding/json"
	"embed"

	"github.com/nicksnyder/go-i18n/v2/i18n"	
	"golang.org/x/text/language"
	
)

// Language defines the language config for the application
type Language struct {
	Code string `json:"language"`
}

var (
	lcz *i18n.Localizer
	currentLanguage *Language
	configDir string
)

//go:embed locales
var localeFS embed.FS

func init() {
	// Get language configuration
	configDir = getConfigDir()
	if err := loadLanguage(); err != nil {
		// Try with environment
		loadLanguageEnv()
	}
	
	// Initialize parser
	bnd := i18n.NewBundle(language.English)
	lcz = i18n.NewLocalizer(bnd, currentLanguage.Code)

	// Load message files
	locales := GetAvailableLanguages()
	for _, lcl := range locales {
		file := filepath.Join("locales", lcl + ".json")
		bnd.LoadMessageFileFS(localeFS, file)
	}
}

// returns the terminal's environment language
func getEnvLanguage() language.Tag {
	env := os.Getenv("LANG")
	if env == "" {
		return language.English
	}
	return language.Make(env)
}

// getConfigDir returns the akrasia config directory
func getConfigDir() string {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ".akrasia"
		}
		configHome = filepath.Join(home, ".config")
	}
	return filepath.Join(configHome, "akrasia")
}

// getConfigFile returns the lang config file path
func getConfigFile() string {
	return filepath.Join(configDir, "lang.json")
}

// T returns a localized string from a key
func T(id string) string {
	return lcz.MustLocalize(&i18n.LocalizeConfig{
		MessageID: id,
	})
}

// GetAvailableLanguages returns the available languages to choose
func GetAvailableLanguages() []string {
	// would be better to parse the folder in real time
	return []string{"en", "pt"}
}

// GetCurrentLanguage returns the current language
func GetCurrentLanguage() string {
	return currentLanguage.Code
}

// SetLanguage sets the language and saves it
func SetLanguage(lang string) error {
	currentLanguage.Code = lang

	// Ensure config directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(currentLanguage, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal language config: %w", err)
	}

	configFile := getConfigFile()
	if err := os.WriteFile(configFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write language config file: %w", err)
	}

	return nil
}

// loadLanguage loads the language set in config
func loadLanguage() error {
	configFile := getConfigFile()

	data, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}

	var lang Language
	if err := json.Unmarshal(data, &lang); err != nil {
		return err
	}

	currentLanguage = &lang

	return nil

}

// loadLanguageEnv loads the language from environment
func loadLanguageEnv() {
	var lang Language
	lang.Code = getEnvLanguage().String()
	currentLanguage = &lang
}


