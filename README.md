# Akrasia

<img width="1376" height="768" alt="akrasia_logo" src="https://github.com/user-attachments/assets/09a4b50e-c9a4-4dd3-be84-88ca92d21a88" />

---

_Akrasía_ is a Greek word meaning "incontinence" or lack of self-control. This app helps you manage tasks you need to complete but tend to procrastinate on.     
As Plato wrote in _Laws_, humans are engaged in a never-ending internal war within their own souls — **a battle against pleasure-seeking**. Today, as we scroll through endless feeds of videos and posts, we chase instant gratification while neglecting the meaningful goals we should pursue.    
This app aims to help you regain self-control in your daily life.  

## Motivation

Yes, there are many applications for task management out there, but in my experience using them, I found that none truly met my needs. I have tried numerous productivity apps, each with different approaches to organizing tasks and sending reminders. Yet, none retained my engagement for more than a short period.

This gap between available tools and my personal workflow led me to develop my own application. I wanted a tool that was intuitive, aligned with how I think about productivity, and motivating enough to use consistently.

## Requirements

- [Go 1.25 or later](https://go.dev/doc/install)
- SQLite3
- [Goose](https://pressly.github.io/goose/installation/)
- A terminal

## Quick Start

1. Clone the repo

```bash
git clone git@github.com:luis-octavius/akrasia.git && cd akrasia
```

2. Install it:

```bash
go install .
```

3. Initialize database:

```bash
akrasia init
```

4. Create cron job to update daily tasks automatically:

```bash
akrasia create-cron # or cc
```

5. (Optional) Create an alias:

```bash
echo "akr='akrasia'" >> ~/.zshrc # or .bashrc
```

## ✨ Features

- **Simple Task Creation** — Use positional arguments for intuitive task creation
  - `akrasia add "Task name"` — Create task quickly
  - `akrasia add "Task name" "Description"` — Add description without flags
  
- **Colorful & Accessible** — Configurable color themes for better readability
  - Default theme with standard colors
  - High-contrast theme for improved accessibility
  - Manage themes with `akrasia config theme`

- **Smart Quotes** — Motivational quotes that adapt to your terminal width
  - Automatic word-wrapping for any terminal size
  - Never breaks words mid-line

- **Daily Task Management** — Built-in support for recurring daily tasks with streak tracking

- **Focus Dashboard** — See what matters now with the `today` and `focus` commands

## Commands

```
Usage:
  akrasia [command]

Available Commands:

  add                create a new task with optional description, priority, and expiration date
  backfill-history   backfill missing daily task history entries
  check-expired      show tasks that have passed their expiration date
  check-expiring     check todos that are expiring in 5 days
  completion         Generate the autocompletion script for the specified shell
  config             manage application settings (themes, preferences)
  create-cron        create the cron job to update daily tasks
  delete-concluded   delete all concluded todos
  delete-by-name     delete a todo by name
  focus              show top 1-3 priority items to focus on right now
  get-all            returns all the todos saved in storage
  get-by-name        search for a task by name (case-insensitive, fuzzy match)
  get-daily          show all daily tasks
  help               Help about any command
  history            get the streak history of provided todo
  init               initialize akrasia app
  streak             get the current streak of provided todo
  today              show a daily focus dashboard (overdue, due today, daily pending, expiring soon)
  update-daily       update daily todos (used by cron, not intended for manual use)
  update-status      mark a task as completed and record it in history with optional notes

Flags:
  -h, --help   help for akrasia

Use "akrasia [command] --help" for more information about a command.
```

## Usage

```bash
# add command with positional arguments (new style - simpler!)
akrasia add "Stendhal"
2026/01/04 11:38:31 Task Stendhal created successfully!

# add command with positional name and description
akrasia add "Stendhal" "Finish the book The Red and The Black" --priority high
2026/01/04 11:38:31 Task Stendhal created successfully!

# add command with flags (backward compatible)
akrasia add --name Stendhal --desc "Finish the book The Red and The Black" 13
2026/01/04 11:38:31 Todo Stendhal created successfully!

# add a daily task
akrasia add "Morning run" --daily --priority high
2026/03/08 08:15:20 Task Morning run created successfully!

# manage color themes
akrasia config theme list           # show available themes
Available themes:
  - default
  - high-contrast

akrasia config theme show           # show current theme
Current theme: default

akrasia config theme high-contrast  # switch to high-contrast theme for better accessibility
Theme set to: high-contrast

# mark task as done
akrasia update-status --name stendhal # case-insensitive, mark Todo as done
Stendhal | Finish the book The Red and The Black |
13 Feb 26 00:00 UTC | Done

# get all tasks with priority filter
akrasia get-all --priority high
Todos:

Stendhal | Finish the book The Red and The Black |
13 Feb 26 00:00 UTC | Done

# delete concluded tasks (requires confirmation)
akrasia delete-concluded --yes
Concluded Todos deleted successfully!

# backfill history for the last 30 days (default)
akrasia backfill-history
Backfilled 90 history entries for 3 daily task(s)

# backfill history for a specific task and 60 days back
akrasia backfill-history --task "Morning run" --days 60
Backfilled 60 history entries for task 'Morning run'

# see what needs attention now
akrasia today --only overdue --limit 5 --priority high
TODAY FOCUS

OVERDUE (1)
Pay electricity bill | ...

DUE TODAY (2)
Prepare weekly report | ...

DAILY PENDING (1)
Morning run | ...

# pick top items to execute now
akrasia focus --limit 3 --priority high
FOCUS (3)
Pay electricity bill | ...
Prepare weekly report | ...
Morning run | ...

# get streak for a daily task
akrasia streak --name "Morning run"
Your current streak with Morning run is: 5

# get streak history  
akrasia history --name "Morning run"
1. Started Date: 2026-02-15 | End Date: 2026-02-20 | Total Days: 6
2. Started Date: 2026-01-10 | End Date: 2026-01-15 | Total Days: 6
```

## Contributing

Contributions are welcome!
If you're willing to contribute, just fork the project and open a pull request at main.

## License

This project is licensed under the MIT License - see the [LICENSE](docs/LICENSE.md) file for details.

## Acknowledgments

- Inspired by ancient Greek philosophy on self-control
- Built with Go

[PT-BR Version](docs/README-pt.md)
