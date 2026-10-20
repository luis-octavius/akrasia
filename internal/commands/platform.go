package commands

import (
	"fmt"
	"runtime"

	"github.com/luis-octavius/akrasia/pkg/i18n"
)

// goos is runtime.GOOS; tests replace it to check the Windows messages.
var goos = runtime.GOOS

// unixOnly stops a command that hands work to a Unix tool (systemd, cron)
// when running on Windows, pointing to WSL, where the tool exists. The
// rest of akrasia (CLI, TUI, serve, backups) works on Windows.
func unixOnly(command, tool string) error {
	if goos != "windows" {
		return nil
	}
	return fmt.Errorf(i18n.T("commands.error.windowsUnsupported"), command, tool)
}
