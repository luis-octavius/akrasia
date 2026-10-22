# Akrasia

_Akrasía_ is a Greek word meaning "incontinence" or lack of self-control. This app helps you manage tasks you need to complete but tend to procrastinate on.
As Plato wrote in _Laws_, humans are engaged in a never-ending internal war within their own souls — **a battle against pleasure-seeking**. Today, as we scroll through endless feeds of videos and posts, we chase instant gratification while neglecting the meaningful goals we should pursue.
This app aims to help you regain self-control in your daily life.

## Motivation

Yes, there are many applications for task management out there, but in my experience using them, I found that none truly met my needs. I have tried numerous productivity apps, each with different approaches to organizing tasks and sending reminders. Yet, none retained my engagement for more than a short period.

This gap between available tools and my personal workflow led me to develop my own application. I wanted a tool that was intuitive, aligned with how I think about productivity, and motivating enough to use consistently.

## Features

- **Interactive TUI**: run `akrasia` to see today's tasks, complete, add, delete and search them, and follow your streaks, all from the keyboard (optional vim keys).
- **Fast CLI**: every action is also a single command, such as `akrasia add`, `akrasia today` or `akrasia update-status`, which makes it easy to script.
- **Daily, weekly and monthly tasks, and streaks**: track habits and routines (with an optional time of day), see your current streak and its history, and backfill days you forgot to log.
- **Subtasks**: split a task into steps; finishing the last one completes it.
- **On your phone**: `akrasia serve` gives a web app you can install on the home screen, reached privately over Tailscale.
- **Backups and reminders**: daily copies to a folder you already sync, and reminders of what's due in a Discord channel.
- **Themes**: nine color themes (`catppuccin-mocha`, `tokyo-night`, `gruvbox`, `nord`, `rose-pine`, `cyberpunk`, …).
- **English and Portuguese** interface.

## Requirements

- [Go 1.27.1 or later](https://go.dev/doc/install)
- A terminal

SQLite is built in (pure-Go driver) and the database migrations ship inside the binary, so nothing else needs to be installed.

## Platforms

Akrasia is made for **Linux**, where everything works.

On **Windows**, the recommended way is [WSL](https://learn.microsoft.com/windows/wsl/install) (Windows Subsystem for Linux): install a Linux distribution with `wsl --install`, then follow the [Quick Start](docs/QUICK-START.md) inside it. Everything works there, including `akrasia serve install` (enable systemd in WSL) and the reminders with cron.

Akrasia also runs natively on Windows (`go install .` in PowerShell), with these limits:

- The CLI, the TUI (use Windows Terminal), `akrasia serve`, backups and `akrasia notify` work.
- `akrasia serve install` (systemd) and `akrasia config notify install` (cron) don't: they stop with a message. Use the Task Scheduler to start `akrasia serve` at login, or to run `akrasia notify` every 15 minutes.
- Windows doesn't set `LANG`, so the interface starts in English. Run `akrasia config language pt` for Portuguese.
- `akrasia restore` can't tell whether `akrasia serve` is running, so stop it before restoring.

## Links

- [Quick Start](docs/QUICK-START.md)
- [Commands](docs/COMMANDS.md)
- [Usage](docs/USAGE.md)
- [Contributing](docs/CONTRIBUTING.md)
- [Changelog](CHANGELOG.md)

## License

This project is licensed under the MIT License - see the [LICENSE](docs/LICENSE.md) file for details.

## Acknowledgments

- Inspired by ancient Greek philosophy on self-control
- Built with Go and [Charm](https://charm.land) (Bubble Tea, Lip Gloss, Bubbles)
