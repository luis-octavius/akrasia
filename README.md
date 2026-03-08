# Akrasia

<img src="https://i.imgur.com/tWFhpXs.gif" />
    
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

## Commands

```
Usage:
  akrasia [command]

Available Commands:

  add                adds a todo in storage, description is optional
  backfill-history   backfill missing daily task history entries
  check-expired      check expired todos
  check-expiring     check todos that are expiring in 5 days
  completion         Generate the autocompletion script for the specified shell
  create-cron        create the cron job to update daily tasks
  delete-concluded   delete all concluded todos
  get-all            returns all the todos saved in storage
  get-by-name        gets a todo by name
  get-daily          show all daily tasks
  help               Help about any command
  history            get the streak history of provided todo
  delete-by-name     delete a todo by name
  init               initialize akrasia app
  streak             get the current streak of provided todo
  update-daily       update daily todos (used by cron, not intended for manual use)
  update-status      update concluded status to true

Flags:
  -h, --help   help for akrasia

Use "akrasia [command] --help" for more information about a command.
```

## Usage

```bash
# add command
akrasia add --name Stendhal --desc "Finish the book The Red and The Black" 13
2026/01/04 11:38:31 Todo Stendhal created successfully!

# add a daily task
akrasia add --name "Morning run" --daily
2026/03/08 08:15:20 Task Morning run created successfully!

akrasia update-status --name stendhal # case-insensitive, mark Todo as done
Stendhal | Finish the book The Red and The Black |
13 Feb 26 00:00 UTC | Done

akrasia get-all #
Todos:

Stendhal | Finish the book The Red and The Black |
13 Feb 26 00:00 UTC | Done

akrasia delete-concluded # self-explanatory
Concluded Todos deleted successfully!

# backfill history for the last 30 days (default)
akrasia backfill-history
Backfilled 90 history entries for 3 daily task(s)

# backfill history for a specific task and 60 days back
akrasia backfill-history --task "Morning run" --days 60
Backfilled 60 history entries for task 'Morning run'

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
