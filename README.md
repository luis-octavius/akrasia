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
- **Daily tasks and streaks**: track habits, see your current streak and its history, and backfill days you forgot to log.
- **Subtasks**: split a task into steps; finishing the last one completes it.
- **On your phone**: `akrasia serve` gives a web app you can install on the home screen, reached privately over Tailscale.
- **Backups and reminders**: daily copies to a folder you already sync, and reminders of what's due in a Discord channel.
- **Themes**: nine color themes (`catppuccin-mocha`, `tokyo-night`, `gruvbox`, `nord`, `rose-pine`, `cyberpunk`, …).
- **English and Portuguese** interface.

## Requirements

- [Go 1.27.1 or later](https://go.dev/doc/install)
- A terminal

SQLite is built in (pure-Go driver) and the database migrations ship inside the binary, so nothing else needs to be installed.

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
