// Package tui is the interactive terminal interface for Akrasia, built on
// Bubble Tea. It reads everything through the tasks service, the same one
// the CLI commands use.
package tui

import (
	"context"

	tea "charm.land/bubbletea/v2"
)

// Run starts the TUI in the alternate screen and blocks until the user
// quits. Queries still running when it returns are cancelled.
func Run(ctx context.Context, svc Service) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	p := tea.NewProgram(New(ctx, svc), tea.WithContext(ctx))
	_, err := p.Run()
	return err
}
