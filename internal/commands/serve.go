package commands

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/luis-octavius/akrasia/internal/backup"
	"github.com/luis-octavius/akrasia/internal/db"
	"github.com/luis-octavius/akrasia/internal/notify"
	"github.com/luis-octavius/akrasia/internal/web"
	"github.com/luis-octavius/akrasia/pkg/color"
	"github.com/luis-octavius/akrasia/pkg/i18n"
	"github.com/spf13/cobra"
)

var (
	serveAddr string
	tokenNew  bool
)

// serviceName is the systemd user unit `serve install` writes.
const serviceName = "akrasia.service"

// jobsEvery is how often serve runs the daily backup and the reminders;
// both only act when something is due.
const jobsEvery = 5 * time.Minute

// serveCmd runs the web interface, and while it runs, the daily backup
// and the Discord reminders.
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: i18n.T("commands.commands.serveShort"),
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		tkm, err := taskManagerFromContext(cmd.Context())
		if err != nil {
			return err
		}
		cfg, err := web.LoadConfig()
		if err != nil {
			return err
		}
		addr := cfg.ListenAddr()
		if serveAddr != "" {
			addr = serveAddr
		}

		release, err := web.Lock(db.GetDBPath())
		if errors.Is(err, web.ErrRunning) {
			return errors.New(i18n.T("commands.error.serveRunning"))
		}
		if err != nil {
			return err
		}
		defer release()

		handler, err := web.New(tkm, cfg.Token)
		if err != nil {
			return err
		}

		ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		listener, err := net.Listen("tcp", addr)
		if err != nil {
			return err
		}
		server := &http.Server{Handler: handler, ReadHeaderTimeout: 10 * time.Second}

		go runJobs(ctx, tkm)
		go func() {
			<-ctx.Done()
			shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = server.Shutdown(shutdown)
		}()

		fmt.Printf(i18n.T("commands.info.serving"), "http://"+listener.Addr().String(), port(addr))
		if err := server.Serve(listener); !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	},
}

// runJobs does what cron did before serve: the daily backup and the
// reminders, now and then every jobsEvery. Both read their config each
// time, so a change applies without a restart.
func runJobs(ctx context.Context, svc notify.Service) {
	run := func() {
		now := time.Now()
		if cfg, err := backup.LoadConfig(); err != nil {
			log.Printf("backup: %v", err)
		} else if _, err := backup.Daily(ctx, db.GetDBPath(), cfg, now); err != nil {
			log.Printf("backup: %v", err)
		}

		if cfg, err := notify.LoadConfig(); err != nil {
			log.Printf("notify: %v", err)
		} else if _, err := notify.Run(ctx, svc, cfg, now, notify.Discord(cfg)); err != nil {
			log.Printf("notify: %v", err)
		}
	}

	run()
	ticker := time.NewTicker(jobsEvery)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			run()
		}
	}
}

// port is the port of a host:port address, for the tailscale hint.
func port(addr string) string {
	if _, p, err := net.SplitHostPort(addr); err == nil {
		return p
	}
	return addr
}

// serveTokenCmd shows the token a device enters once, or replaces it.
var serveTokenCmd = &cobra.Command{
	Use:   "token",
	Short: i18n.T("commands.commands.serveTokenShort"),
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := web.LoadConfig()
		if err != nil {
			return err
		}

		if tokenNew {
			if cfg.Token, err = web.NewToken(); err != nil {
				return err
			}
			if err := web.SaveConfig(cfg); err != nil {
				return err
			}
			color.MsgWarning(i18n.T("commands.info.serveTokenNew"))
		}

		fmt.Println(cfg.Token)
		return nil
	},
}

// systemctl runs `systemctl --user`; tests replace it.
var systemctl = func(args ...string) error {
	out, err := exec.Command("systemctl", append([]string{"--user"}, args...)...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("systemctl --user %s: %v: %s", strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return nil
}

// serveInstallCmd makes serve a systemd user service: it starts with the
// computer and comes back if it stops. Serve then runs the reminders, so
// the crontab entry of `config notify install` is removed.
var serveInstallCmd = &cobra.Command{
	Use:   "install",
	Short: i18n.T("commands.commands.serveInstallShort"),
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := unixOnly("serve install", "systemd"); err != nil {
			return err
		}

		exe, err := os.Executable()
		if err != nil {
			return err
		}
		if exe, err = filepath.EvalSymlinks(exe); err != nil {
			return err
		}

		unit := unitFile()
		if err := os.MkdirAll(filepath.Dir(unit), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(unit, []byte(serviceUnit(exe)), 0644); err != nil {
			return err
		}

		if err := systemctl("daemon-reload"); err != nil {
			return err
		}
		if err := systemctl("enable", "--now", serviceName); err != nil {
			return err
		}
		if err := removeNotifyJob(); err != nil {
			return err
		}

		cfg, err := web.LoadConfig()
		if err != nil {
			return err
		}
		color.MsgSuccess(fmt.Sprintf(i18n.T("commands.info.serveInstalled"), unit))
		fmt.Printf(i18n.T("commands.info.serveNext"), port(cfg.ListenAddr()))
		return nil
	},
}

var serveUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: i18n.T("commands.commands.serveUninstallShort"),
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		unit := unitFile()
		if _, err := os.Stat(unit); errors.Is(err, os.ErrNotExist) {
			return nil
		}

		if err := systemctl("disable", "--now", serviceName); err != nil {
			return err
		}
		if err := os.Remove(unit); err != nil {
			return err
		}
		if err := systemctl("daemon-reload"); err != nil {
			return err
		}

		color.MsgSuccess(i18n.T("commands.info.serveUninstalled"))
		if cfg, err := notify.LoadConfig(); err == nil && cfg.Enabled() {
			color.MsgMuted(i18n.T("commands.info.serveNotifyHint"))
		}
		return nil
	},
}

// serviceUnit is the systemd unit running exe serve. Like the crontab
// entry, it carries the database and config paths in use now.
func serviceUnit(exe string) string {
	env := "Environment=" + systemdQuote("DB_PATH="+db.GetDBPath()) + "\n"
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		env += "Environment=" + systemdQuote("XDG_CONFIG_HOME="+xdg) + "\n"
	}

	return "[Unit]\n" +
		"Description=Akrasia web interface, daily backup and reminders\n" +
		"After=network-online.target\n\n" +
		"[Service]\n" +
		env +
		"ExecStart=" + systemdQuote(exe) + " serve\n" +
		"Restart=on-failure\n" +
		"RestartSec=5\n\n" +
		"[Install]\n" +
		"WantedBy=default.target\n"
}

// systemdQuote quotes a unit file value: backslashes and quotes are
// escaped, and % is doubled so systemd doesn't read it as a specifier.
func systemdQuote(s string) string {
	s = strings.NewReplacer(`\`, `\\`, `"`, `\"`, "%", "%%").Replace(s)
	return `"` + s + `"`
}

// unitFile is where systemd looks for the user's own units.
func unitFile() string {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		home, _ := os.UserHomeDir()
		configHome = filepath.Join(home, ".config")
	}
	return filepath.Join(configHome, "systemd", "user", serviceName)
}
