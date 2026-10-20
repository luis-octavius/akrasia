# Changelog

All notable changes to Akrasia are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and versions follow
[Semantic Versioning](https://semver.org/).

## [Unreleased]

### Changed

- The database opens in WAL mode and waits up to 5 seconds for another
  writer, so the CLI, the TUI and other processes can use it at the same
  time without "database is locked" errors. A `-wal` and a `-shm` file may
  appear next to `akrasia.db` while it is open.

## [2.4.0] - 2026-10-20

### Added

- **Subtasks**, one level deep: a task can be split into subtasks, and the
  parent shows how many are done (`2/5`).
  - Finishing the last subtask completes the parent; completing the parent
    with pending subtasks completes them too, after a confirmation
    (`update-status --yes` skips it). A new subtask reopens a finished
    parent, and deleting a parent deletes its subtasks.
  - A subtask follows its parent's daily mode and is due no later than it
    (by default, on the same day). For a daily parent, each subtask is done
    once a day and the parent closes the day, and its streak, when all are
    done.
  - CLI: `akrasia add <name> --parent <task>`; listings show subtasks
    indented under their parent, and `today --json` nests them in
    `Subtasks`.
  - TUI: subtasks under their parent with its progress, `s` adds one, `z`
    folds them, and `space` on a parent asks before completing them all.
  - Today's progress, `today`, `focus` and reminders count a parent as one
    task.

### Changed

- Task names only need to be unique among the subtasks of one parent (and
  among top-level tasks). A subtask can be named `"Parent/Subtask"`; on an
  exact name a top-level task comes first, and a name shared by subtasks
  of several parents is refused with their full names to choose from.

## [2.3.0] - 2026-10-20

### Added

- **Reminders in a Discord channel**, through a channel webhook (free, no
  bot).
  - Tasks due within 3 days and overdue ones at 10:00 and 19:00; the 19:00
    reminder also lists the daily tasks not done today.
  - Mentions you when `config notify mention <user-id>` is set; nobody else
    can be pinged from a task name.
  - Each time is sent once a day, a day with nothing due sends nothing, and
    a time missed while the computer was off is sent at the next check.
  - `akrasia config notify discord <url> | mention <id> | days <n> |
    times <HH:MM,...> | daily-at <HH:MM> | show | test | off`.
  - `akrasia config notify install` adds a crontab entry that runs the new
    `akrasia notify` every 15 minutes (`uninstall` or `off` removes it).

### Changed

- New versions apply their database migrations on their own; `akrasia
  init` is only needed once, to create the database.

### Fixed

- Reading the crontab no longer treats any error as an empty crontab,
  which could have overwritten the user's other jobs when adding one.

## [2.2.0] - 2026-10-19

### Added

- **Backups to a folder you already sync.** No account or server: point
  Akrasia at a Google Drive, Dropbox, OneDrive or Syncthing folder and the
  sync tool takes the copies off the machine.
  - `akrasia config backup dir <path>` turns it on; `keep <n>` (default
    14), `profile <name>` (default: the hostname), `show` and `off`.
  - Once a day, the first command (or opening the TUI) saves a checked,
    compacted copy and deletes the oldest ones beyond `keep`. A failure is
    only a warning.
  - Each machine has its own profile, so several can share one folder.
  - `akrasia backup` (alias `bk`) saves a copy now; `akrasia backup list`
    shows them.
  - `akrasia restore <file|latest>` checks the copy, saves the current
    database first, then replaces it. `--profile` restores another
    machine's copies; `--yes` skips the question.
  - The TUI status bar shows the date of the last backup.

## [2.1.0] - 2026-10-19

Merged to `main` in #55. New features on top of 2.0.0, with no breaking
changes: `today --json` only gains a field.

### Added

- **Today's progress bar**, in the theme's gradient colors: done today out
  of today's tasks (every daily task, plus one-off tasks that are overdue,
  due today or completed today).
  - It appears below the tabs in the TUI and refreshes after every change.
  - It opens `akrasia today` and `akrasia focus`.
  - New command: `akrasia progress` (alias `pg`).
  - `today --json` gains a `Progress` field (`Done`, `Total`).
- **Editing tasks.** Change a task's name, description, priority, daily mode
  or due date. Its history and streak are kept.
  - CLI: `akrasia edit <name>` (alias `e`) with `--name`, `--desc`,
    `--priority`, `--daily=true|false` and `--date`. Only the flags you
    pass are changed; `--desc ""` clears the description.
  - TUI: press `e` on a task. The form opens prefilled and only changed
    fields are saved, so an overdue task can be edited without touching its
    date.
  - A one-off task turned daily starts its streak window that day.
  - Clear errors when the new name is already taken or the priority is
    invalid.
- **Notes for a task.** See what was written on each completed day, newest
  first; useful as a journal of a daily streak.
  - CLI: `akrasia notes <name>` (alias `nt`, `--limit N`), with the current
    streak for daily tasks.
  - TUI: the details panel shows the last 3 notes, and `n` opens every note
    in a scrollable window (`esc` closes).
  - Backfilled days and days without notes are marked as such.
  - Rewriting today's note: completing a task again on the same day
    replaces its note (now stated in `update-status --help`). In the TUI,
    `space` on a task already done today opens today's note, ready to
    edit, instead of only saying "already done".
- **Quotes in the TUI.** Completing a task opens a pop-up with
  "✔ X completed" and a quote in the theme's quote style, the same moment
  the CLI prints one. `esc`, `enter` or `q` close it; other keys are
  ignored while it is open. `Q` shows a quote at any time.
- **Dump: delete entries and turn them into tasks.** `akrasia dump` now
  numbers the list.
  - `--delete N` removes an entry.
  - `--to-task N` creates a task from it (optionally with `--priority`,
    `--daily`, `--date`) and removes it from the dump. If the task can't be
    created (for example, the name is taken), the entry stays.
  - In the TUI, on the Dump tab: `x` deletes with confirmation, and `c`
    opens the new-task form prefilled with the entry.
- **Years in due dates.** `--date dd,mm,yyyy` in `add` and `edit`, and
  `dd/mm/yyyy` in the TUI form (a two-digit year means 20yy). The edit form
  prefills the full date.

### Fixed

- Dump dates are shown without a wrong "UTC" label; they are stored in
  local time.
- Adding a task whose name is taken says so ("there is already a task
  called X") instead of a raw database error.
- Due dates are set at local midnight instead of UTC midnight. Before, a
  task due on the 21st showed up as due on the 20th in timezones behind
  UTC (like São Paulo).
- A due date of today is accepted; before, it was rejected as "in the
  past".
- Days that don't exist (31/02, 29/02 outside leap years) are rejected
  instead of silently becoming a day in the next month.
- TUI: a details panel taller than the screen is cut instead of pushing
  the status bar and help out of view.

## [2.0.0] - 2026-10-19

The headline of this release is an interactive terminal interface (TUI). It
is a major version because two behaviors that scripts may rely on changed:
`akrasia` with no arguments and the shape of `today --json` (see
[Changed](#changed)).

### Added

- **Interactive TUI.** Run `akrasia` with no arguments, or `akrasia tui`.
  - Four tabs, each with an item count:
    - **Today**: overdue, due today, pending daily and expiring soon;
    - **All**;
    - **Daily**;
    - **Dump**.
  - Complete the selected task with `space`. A prompt asks for optional
    notes.
  - Create a task with `a` in a form (name, description, priority, daily,
    due date). On the Dump tab, `a` adds a dump entry instead.
  - Delete the selected task with `x`, after a y/n confirmation (`s` is
    accepted for "sim").
  - Filter by name with `/` and cycle a priority filter with `p`. `esc`
    clears the filters.
  - Open a details panel with `enter`. It shows due date, priority and
    status, plus the current streak and streak history for daily tasks.
  - Switch to the next theme with `t`; the choice is saved. Reload a tab with
    `r`, list every key with `?`, quit with `q`.
  - Prompts and the new-task form open as centered pop-up windows over the
    list. A status bar shows the current mode, the last message and the
    active theme.
- **Optional vim keys.** `akrasia config keymap vim` adds `j` `k` `g` `G` `h`
  `l` on top of the arrow keys. `akrasia config keymap default` removes them;
  `list` and `show` are also available.
- **Six new themes:** `rose-pine`, `cyberpunk`, `catppuccin-mocha`,
  `tokyo-night`, `gruvbox` and `nord`, next to `default`, `high-contrast`
  and `purple`. Themes now also color the TUI: accent, selection, status
  bar, priority badges and gradient borders.
- `docs/COMMANDS.md` documents the TUI keys, keymaps and themes.

### Documentation

- The README, quick start, usage, commands and contributing guides were
  rewritten to match the current app:
  - Go 1.27.1;
  - no separate SQLite or Goose install (both are built in);
  - `create-cron` and `delete-concluded` removed from the docs;
  - real command output;
  - Docker instructions;
  - the service/render/TUI architecture for contributors.
- The Portuguese docs (`docs/pt`) were removed; the app itself is still
  translated to Portuguese.

### Changed

- **`akrasia` with no arguments opens the TUI** instead of printing the help.
  `akrasia help` and `akrasia --help` still print the help. When stdin or
  stdout is not a terminal (pipes, scripts, cron), `akrasia` prints the help
  as before.
- **`today --json` output changed.** Each task now has the fields `ID`,
  `Name`, `Description`, `Concluded`, `ExpiresAt`, `Priority`, `IsDaily` and
  `DoneToday`. `CreatedAt`, `UpdatedAt` and `HistorySince` were removed, and
  `DoneToday` is new.
- Colored CLI output adapts to the terminal: hex theme colors are converted
  on terminals without true color, and colors are removed when the output
  goes to a pipe or a file.
- `check-expiring` prints its "Expiring..." header once instead of once per
  task.
- Updated to Bubble Tea v2, Lip Gloss v2 and Bubbles v2
  (`charm.land/*/v2`).
- Internal: the `tasks` package no longer prints anything. Its methods take
  a `context.Context` and return data, the CLI output lives in
  `internal/commands/render.go`, and the TUI reuses the same methods.

### Fixed

- Daily tasks no longer show their creation date as a due date (#52).
- Daily tasks show whether they were done **today**; before, they always
  showed "Not done".
- Invalid dates or months in `add --date` return an error instead of
  stopping the program abruptly (`log.Fatal`).
- Lookups by name report a clear "no task matches" message. Errors keep
  their underlying cause instead of discarding it.
- Wrapped quotes no longer start their continuation lines with a space.

### Upgrade notes

- Scripts that read `today --json` must use the new task fields listed
  above.
- An existing `~/.config/akrasia/theme.json` keeps working: the colors it
  lacks for the new TUI roles are taken from the built-in theme of the same
  name.

## [1.1.1] and earlier

See the [GitHub releases and tags](https://github.com/luis-octavius/akrasia/tags).

[Unreleased]: https://github.com/luis-octavius/akrasia/compare/v.2.4.0...dev
[2.4.0]: https://github.com/luis-octavius/akrasia/compare/v.2.3.0...v.2.4.0
[2.3.0]: https://github.com/luis-octavius/akrasia/compare/v.2.2.0...v.2.3.0
[2.2.0]: https://github.com/luis-octavius/akrasia/compare/v.2.1.0...v.2.2.0
[2.1.0]: https://github.com/luis-octavius/akrasia/compare/v.2.0.0...v.2.1.0
[2.0.0]: https://github.com/luis-octavius/akrasia/compare/v.1.1.1...v.2.0.0
[1.1.1]: https://github.com/luis-octavius/akrasia/releases/tag/v.1.1.1
