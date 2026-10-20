package commands

import (
	"strings"
	"testing"
)

func TestUnixOnlyCommandsExplainWSLOnWindows(t *testing.T) {
	saved := goos
	t.Cleanup(func() { goos = saved })

	goos = "windows"
	if err := unixOnly("serve install", "systemd"); err == nil || !strings.Contains(err.Error(), "WSL") || !strings.Contains(err.Error(), "systemd") {
		t.Errorf("on Windows the error should name the tool and point to WSL, got %v", err)
	}
	if err := installNotifyJob(); err == nil || !strings.Contains(err.Error(), "cron") {
		t.Errorf("notify install should stop before touching cron on Windows, got %v", err)
	}

	goos = "linux"
	if err := unixOnly("serve install", "systemd"); err != nil {
		t.Errorf("on Linux nothing should be blocked, got %v", err)
	}
}
