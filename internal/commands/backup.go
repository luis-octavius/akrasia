package commands

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/luis-octavius/akrasia/internal/backup"
	"github.com/luis-octavius/akrasia/internal/db"
	"github.com/luis-octavius/akrasia/pkg/color"
	"github.com/luis-octavius/akrasia/pkg/i18n"
	"github.com/spf13/cobra"
)

var (
	backupProfile string
	restoreYes    bool
)

// backupCmd copies the database to the backup folder now.
var backupCmd = &cobra.Command{
	Use:     "backup",
	Short:   i18n.T("commands.commands.backupShort"),
	Aliases: []string{"bk"},
	Example: "akrasia backup\nakrasia backup list",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := enabledBackupConfig()
		if err != nil {
			return err
		}

		path, err := backup.Snapshot(cmd.Context(), db.GetDBPath(), cfg.Dir, cfg.ProfileName(), time.Now())
		if err != nil {
			return err
		}
		removed, err := backup.Prune(cfg.Dir, cfg.ProfileName(), cfg.KeepN())
		if err != nil {
			return err
		}

		color.MsgSuccess(fmt.Sprintf(i18n.T("commands.info.backupCreated"), path))
		if len(removed) > 0 {
			color.MsgMuted(fmt.Sprintf(i18n.T("commands.info.backupPruned"), len(removed)))
		}
		return nil
	},
}

// backupListCmd shows a profile's copies, newest first.
var backupListCmd = &cobra.Command{
	Use:     "list",
	Short:   i18n.T("commands.commands.backupListShort"),
	Aliases: []string{"ls"},
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := enabledBackupConfig()
		if err != nil {
			return err
		}

		profile := backupProfile
		if profile == "" {
			profile = cfg.ProfileName()
		}

		copies, err := backup.List(cfg.Dir, profile)
		if err != nil {
			return err
		}

		printBackups(cfg.Dir, profile, copies)
		return nil
	},
}

// restoreCmd replaces the database with a copy, after saving the current one.
var restoreCmd = &cobra.Command{
	Use:     i18n.T("commands.commands.restoreUse"),
	Short:   i18n.T("commands.commands.restoreShort"),
	Example: i18n.T("commands.commands.restoreExample"),
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := backup.LoadConfig()
		if err != nil {
			return err
		}

		src, err := restoreSource(cfg, args[0])
		if err != nil {
			return err
		}

		if !restoreYes {
			ok, err := confirm(cmd, fmt.Sprintf(i18n.T("commands.info.restoreConfirm"), db.GetDBPath(), src))
			if err != nil {
				return err
			}
			if !ok {
				color.MsgMuted(i18n.T("commands.info.restoreCancelled"))
				return nil
			}
		}

		safety, err := backup.Restore(cmd.Context(), db.GetDBPath(), src, cfg, time.Now())
		if err != nil {
			return err
		}

		// an older copy may predate the latest migrations
		conn, err := db.InitDB()
		if err != nil {
			return fmt.Errorf("%s%w", i18n.T("commands.error.openDatabase"), err)
		}
		conn.Close()

		color.MsgSuccess(fmt.Sprintf(i18n.T("commands.info.restored"), src))
		if safety != "" {
			color.MsgMuted(fmt.Sprintf(i18n.T("commands.info.restoreSafety"), safety))
		}
		return nil
	},
}

// restoreSource resolves `latest` to the newest copy of the profile, or
// returns the given file.
func restoreSource(cfg backup.Config, arg string) (string, error) {
	if arg != "latest" {
		return expandPath(arg)
	}

	if !cfg.Enabled() {
		return "", errors.New(i18n.T("commands.error.backupOff"))
	}

	profile := backupProfile
	if profile == "" {
		profile = cfg.ProfileName()
	}

	latest, ok, err := backup.Latest(cfg.Dir, profile)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", fmt.Errorf(i18n.T("commands.error.noBackups"), profile, cfg.Dir)
	}
	return latest.Path, nil
}

// confirm asks a yes/no question on the terminal. Scripts must pass --yes.
func confirm(cmd *cobra.Command, question string) (bool, error) {
	if !isInteractive() {
		return false, errors.New(i18n.T("commands.error.destructiveAction"))
	}

	fmt.Fprint(cmd.OutOrStdout(), question+" [y/N] ")
	answer, err := bufio.NewReader(cmd.InOrStdin()).ReadString('\n')
	if err != nil && answer == "" {
		return false, nil
	}

	switch strings.ToLower(strings.TrimSpace(answer)) {
	case "y", "yes", "s", "sim":
		return true, nil
	}
	return false, nil
}

// configBackup sets the backup folder, how many copies to keep and the
// profile name.
var configBackup = &cobra.Command{
	Use:     i18n.T("commands.commands.configBackupUse"),
	Short:   i18n.T("commands.commands.configBackupShort"),
	Aliases: []string{"b"},
	Example: i18n.T("commands.commands.configBackupExample"),
	Args:    cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := backup.LoadConfig()
		if err != nil {
			return err
		}

		action, value := args[0], ""
		if len(args) > 1 {
			value = args[1]
		}

		switch action {
		case "show":
			printBackupConfig(cfg)
			return nil

		case "off":
			cfg.Dir = ""
			if err := backup.SaveConfig(cfg); err != nil {
				return err
			}
			color.MsgSuccess(i18n.T("commands.info.backupOff"))
			return nil
		}

		if value == "" {
			return fmt.Errorf(i18n.T("commands.error.needsValue"), action)
		}

		switch action {
		case "dir":
			dir, err := expandPath(value)
			if err != nil {
				return err
			}
			// a typo shows up now instead of at the first backup
			if err := os.MkdirAll(dir, 0755); err != nil {
				return err
			}
			cfg.Dir = dir

		case "keep":
			n, err := strconv.Atoi(value)
			if err != nil || n < 1 {
				return fmt.Errorf(i18n.T("commands.error.backupKeep"), value)
			}
			cfg.Keep = n

		case "profile":
			if !backup.ValidProfile(value) {
				return fmt.Errorf(i18n.T("commands.error.backupProfile"), value)
			}
			cfg.Profile = value

		default:
			return fmt.Errorf(i18n.T("commands.error.backupAction"), action)
		}

		if err := backup.SaveConfig(cfg); err != nil {
			return err
		}
		printBackupConfig(cfg)
		return nil
	},
}

// enabledBackupConfig loads the config and fails when no folder is set.
func enabledBackupConfig() (backup.Config, error) {
	cfg, err := backup.LoadConfig()
	if err != nil {
		return cfg, err
	}
	if !cfg.Enabled() {
		return cfg, errors.New(i18n.T("commands.error.backupOff"))
	}
	return cfg, nil
}

// expandPath resolves ~ and makes the path absolute, so the saved folder
// doesn't depend on where the command was run.
func expandPath(path string) (string, error) {
	if path == "~" || strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(home, strings.TrimPrefix(path, "~"))
	}
	return filepath.Abs(path)
}

// skipDailyBackup lists the commands that don't trigger the daily backup:
// they make their own copy, don't touch the database, or change the
// backup settings themselves.
var skipDailyBackup = map[string]bool{
	"backup": true, "restore": true, "config": true, "init": true,
	"help": true, "completion": true,
}

// runDailyBackup copies the database once a day when a backup folder is
// set. A failure is only a warning: the command still runs.
func runDailyBackup(cmd *cobra.Command) {
	top := cmd
	for top.HasParent() && top.Parent().HasParent() {
		top = top.Parent()
	}
	if top.HasParent() && skipDailyBackup[top.Name()] {
		return
	}

	if err := dailyBackup(cmd.Context()); err != nil {
		// stderr, so `today --json` stays valid JSON
		fmt.Fprintln(cmd.ErrOrStderr(), color.Warning().Render(fmt.Sprintf(i18n.T("commands.error.dailyBackup"), err)))
	}
}

// dailyBackup is a variable so tests never copy the developer's database.
var dailyBackup = func(ctx context.Context) error {
	cfg, err := backup.LoadConfig()
	if err != nil {
		return err
	}
	_, err = backup.Daily(ctx, db.GetDBPath(), cfg, time.Now())
	return err
}
