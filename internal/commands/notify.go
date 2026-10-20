package commands

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/luis-octavius/akrasia/internal/db"
	"github.com/luis-octavius/akrasia/internal/notify"
	"github.com/luis-octavius/akrasia/pkg/color"
	"github.com/luis-octavius/akrasia/pkg/cron"
	"github.com/luis-octavius/akrasia/pkg/i18n"
	"github.com/spf13/cobra"
)

// notifyJob names the crontab entry `config notify install` writes.
const notifyJob = "akrasia-notify"

// notifyCmd sends the reminders that are due. It is meant to run every few
// minutes from the crontab entry `config notify install` adds.
var notifyCmd = &cobra.Command{
	Use:   "notify",
	Short: i18n.T("commands.commands.notifyShort"),
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := notify.LoadConfig()
		if err != nil {
			return err
		}
		if !cfg.Enabled() {
			return errors.New(i18n.T("commands.error.notifyOff"))
		}

		tkm, err := taskManagerFromContext(cmd.Context())
		if err != nil {
			return err
		}

		// quiet when nothing is due: it runs from cron every 15 minutes
		_, err = notify.Run(cmd.Context(), tkm, cfg, time.Now(), notify.Discord(cfg))
		return err
	},
}

// configNotify sets up the Discord reminders.
var configNotify = &cobra.Command{
	Use:     i18n.T("commands.commands.configNotifyUse"),
	Short:   i18n.T("commands.commands.configNotifyShort"),
	Aliases: []string{"n"},
	Example: i18n.T("commands.commands.configNotifyExample"),
	Args:    cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := notify.LoadConfig()
		if err != nil {
			return err
		}

		action, value := args[0], ""
		if len(args) > 1 {
			value = args[1]
		}

		switch action {
		case "show":
			printNotifyConfig(cfg)
			return nil

		case "off":
			cfg.Webhook = ""
			if err := notify.SaveConfig(cfg); err != nil {
				return err
			}
			color.MsgSuccess(i18n.T("commands.info.notifyOff"))
			return removeNotifyJob()

		case "test":
			if !cfg.Enabled() {
				return errors.New(i18n.T("commands.error.notifyOff"))
			}
			tkm, err := taskManagerFromContext(cmd.Context())
			if err != nil {
				return err
			}
			if err := notify.Test(cmd.Context(), tkm, cfg, time.Now(), notify.Discord(cfg)); err != nil {
				return err
			}
			color.MsgSuccess(i18n.T("commands.info.notifyTestSent"))
			return nil

		case "install":
			if !cfg.Enabled() {
				return errors.New(i18n.T("commands.error.notifyOff"))
			}
			return installNotifyJob()

		case "uninstall":
			return removeNotifyJob()
		}

		if value == "" {
			return fmt.Errorf(i18n.T("commands.error.needsValue"), action)
		}

		switch action {
		case "discord":
			if !notify.ValidWebhook(value) {
				return errors.New(i18n.T("commands.error.notifyWebhook"))
			}
			cfg.Webhook = value

		case "mention":
			switch {
			case value == "none":
				cfg.Mention = ""
			case notify.ValidUserID(value):
				cfg.Mention = value
			default:
				return fmt.Errorf(i18n.T("commands.error.notifyMention"), value)
			}

		case "days":
			n, err := strconv.Atoi(value)
			if err != nil || n < 0 || n > 60 {
				return fmt.Errorf(i18n.T("commands.error.notifyDays"), value)
			}
			cfg.Days = &n

		case "times":
			times, err := notify.ParseTimes(value)
			if err != nil {
				return err
			}
			cfg.Times = times

		case "daily-at":
			t, err := notify.ParseTime(value)
			if err != nil {
				return err
			}
			if !slices.Contains(cfg.Slots(), t) {
				return fmt.Errorf(i18n.T("commands.error.notifyDailyAt"), t, cfg.Slots())
			}
			cfg.DailyAt = t

		default:
			return fmt.Errorf(i18n.T("commands.error.notifyAction"), action)
		}

		if err := notify.SaveConfig(cfg); err != nil {
			return err
		}
		printNotifyConfig(cfg)
		return nil
	},
}

// installNotifyJob adds a crontab entry that runs `akrasia notify` every
// 15 minutes. Cron starts with an almost empty environment, so the
// database and config paths in use now are written into the entry.
func installNotifyJob() error {
	if err := unixOnly("config notify install", "cron"); err != nil {
		return err
	}

	exe, err := os.Executable()
	if err != nil {
		return err
	}
	if exe, err = filepath.EvalSymlinks(exe); err != nil {
		return err
	}

	if _, err := exec.LookPath("crontab"); err != nil {
		return errors.New(i18n.T("commands.error.noCrontab"))
	}

	command := fmt.Sprintf("DB_PATH=%s", shellQuote(db.GetDBPath()))
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		command += " XDG_CONFIG_HOME=" + shellQuote(xdg)
	}
	command += " " + shellQuote(exe) + " notify"

	m := cron.NewManager()
	// replace an older entry (another binary path, another DB_PATH)
	if err := m.RemoveJob(notifyJob, ""); err != nil {
		return err
	}
	if err := m.AddJob(notifyJob, "*/15 * * * *", command, "", "Akrasia: Discord reminders"); err != nil {
		return err
	}

	color.MsgSuccess(i18n.T("commands.info.notifyInstalled"))
	return nil
}

func removeNotifyJob() error {
	// without cron there is no entry to remove
	if _, err := exec.LookPath("crontab"); err != nil {
		return nil
	}

	m := cron.NewManager()
	installed, err := m.HasJob(notifyJob, "")
	if err != nil || !installed {
		return err
	}
	if err := m.RemoveJob(notifyJob, ""); err != nil {
		return err
	}
	color.MsgMuted(i18n.T("commands.info.notifyUninstalled"))
	return nil
}

// shellQuote wraps s in single quotes for the crontab's shell.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
