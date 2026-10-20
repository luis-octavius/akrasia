package notify

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"
)

// Defaults for a config that doesn't set them.
const (
	DefaultDays    = 3
	DefaultDailyAt = "19:00"
)

// DefaultTimes are the reminder times of a config that doesn't set them.
var DefaultTimes = []string{"10:00", "19:00"}

var (
	webhookURL = regexp.MustCompile(`^https://(ptb\.|canary\.)?discord(app)?\.com/api/webhooks/\d+/[\w-]+$`)
	userID     = regexp.MustCompile(`^\d{15,25}$`)
)

// Config is saved to notify.json, next to backup.json. Reminders are off
// while Webhook is empty. The file is only readable by its owner: anyone
// with the webhook URL can post to the channel.
type Config struct {
	Webhook string   `json:"webhook"`
	Mention string   `json:"mention,omitempty"`
	Days    *int     `json:"days,omitempty"`
	Times   []string `json:"times,omitempty"`
	DailyAt string   `json:"dailyAt,omitempty"`
}

// Enabled reports whether a webhook is set.
func (c Config) Enabled() bool { return c.Webhook != "" }

// DaysAhead is how many days ahead a due date is reminded (0 = today only).
func (c Config) DaysAhead() int {
	if c.Days == nil {
		return DefaultDays
	}
	return *c.Days
}

// Slots are the reminder times, sorted.
func (c Config) Slots() []string {
	if len(c.Times) == 0 {
		return DefaultTimes
	}
	return c.Times
}

// DailySlot is the reminder time that also lists the daily tasks not done
// yet. It is always one of Slots: the latest one when the saved time was
// removed from the list.
func (c Config) DailySlot() string {
	slots := c.Slots()
	daily := c.DailyAt
	if daily == "" {
		daily = DefaultDailyAt
	}
	if slices.Contains(slots, daily) {
		return daily
	}
	return slots[len(slots)-1]
}

// ValidWebhook reports whether url is a Discord webhook URL.
func ValidWebhook(url string) bool { return webhookURL.MatchString(url) }

// ValidUserID reports whether id looks like a Discord user ID.
func ValidUserID(id string) bool { return userID.MatchString(id) }

// ParseTime checks an HH:MM time and returns it zero-padded.
func ParseTime(s string) (string, error) {
	t, err := time.Parse("15:04", strings.TrimSpace(s))
	if err != nil {
		return "", fmt.Errorf("invalid time %q (use HH:MM, like 19:00)", s)
	}
	return t.Format("15:04"), nil
}

// ParseTimes parses a comma-separated list of HH:MM times, sorted and
// without repeats.
func ParseTimes(s string) ([]string, error) {
	var times []string
	for part := range strings.SplitSeq(s, ",") {
		t, err := ParseTime(part)
		if err != nil {
			return nil, err
		}
		times = append(times, t)
	}
	slices.Sort(times)
	return slices.Compact(times), nil
}

// LoadConfig reads notify.json. A missing file means reminders are off.
func LoadConfig() (Config, error) {
	var cfg Config

	data, err := os.ReadFile(configFile())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return cfg, err
	}

	err = json.Unmarshal(data, &cfg)
	return cfg, err
}

// SaveConfig writes notify.json, readable only by its owner.
func SaveConfig(cfg Config) error {
	if err := os.MkdirAll(filepath.Dir(configFile()), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(configFile(), data, 0600); err != nil {
		return err
	}
	// WriteFile keeps the mode of a file that already exists
	return os.Chmod(configFile(), 0600)
}

// configFile lives next to theme.json and backup.json.
func configFile() string {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return filepath.Join(".akrasia", "notify.json")
		}
		configHome = filepath.Join(home, ".config")
	}
	return filepath.Join(configHome, "akrasia", "notify.json")
}
