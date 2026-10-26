# Changelog

All notable changes to Akrasia are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and versions follow
[Semantic Versioning](https://semver.org/).

## [Unreleased]

## [2.8.0] - 2026-10-26

### Added

- **The phone app works without the computer.** When `akrasia serve` can't
  be reached, the app opens from the last copy of Today, All and Dump, and
  what you add (a task, a subtask, a dump entry) or complete waits in a
  queue on the phone, saved across closing the app.
  - Queued items show under "Waiting to send" as *pending*; a completed task
    shows as done; the bar says "Offline" and how many are waiting.
  - They are sent in order when the computer answers again (on opening the
    app, when the phone is back online and every 30 seconds). A refused one
    (name taken, date passed...) stays as *failed* with the reason, to
    discard; an expired token asks to sign in again, and nothing is lost.
  - Each request has an id the server remembers (`web_queue`, migration
    012, 30 days), so one whose answer was lost is not applied twice.
  - A completion made offline keeps its day (`done_on`, up to 7 days back):
    the history, the streak and the week or month of a recurring task follow
    the day it was done, not the day it was sent. A queued completion must
    match the task's name exactly, so a task renamed on the computer is
    never replaced by a similar one.
  - Adding with an empty name while offline says "name is required" at
    once instead of queuing an item that could only fail.
  - Writes answer with `X-Akrasia-Result: ok|error`.
  - Not offline: editing, deleting or converting a dump entry, and sending
    while the app is closed.
  - New files: `sw.js` (cache, named by a hash of the build so a new
    version replaces the old one) and `queue.js`. Its logic is tested with
    `node --test internal/web/jstest/queue.test.cjs`, which CI runs.

## [2.7.1] - 2026-10-24

### Changed

- **TUI: the All tab is grouped by kind of task**: DAILY, WEEKLY, MONTHLY
  and WITH A DATE (the one-off tasks, by due date). A group without tasks
  has no header. Daily tasks come by time of day, weekly ones by weekday
  and monthly ones by day of the month, timed tasks first.

## [2.7.0] - 2026-10-22

### Added

- **Weekly and monthly tasks**, next to daily ones.
  - `add`, `edit` and `dump --to-task` take `--weekly` or `--monthly`
    (only one of the three kinds at a time), and `--on` for the day: a weekday
    (`mon`..`sun`, `seg`..`dom` or 1-7, Monday = 1) for a weekly task, a day
    of the month (1-31) for a monthly one. Without `--on` the day comes from
    `--date`, or is today's. A monthly task on the 29th-31st falls on the
    last day of shorter months.
  - A weekly task is done for the whole week (Monday to Sunday) once you
    complete it, a monthly one for the whole month. It shows in `today`,
    `focus`, the progress line and the reminders from its day until it is
    done or the period ends; before that it is only in `get-all` and
    `get-daily` (and the Daily tab). When there is one, the section is
    titled "RECURRING PENDING".
  - Streaks count weeks or months in a row, and the period in progress
    does not break them. `streak`, `his`, `notes` and the TUI details say
    which. `backfill-history` stays daily-only.
  - `edit --daily/--weekly/--monthly` change the kind, `--on` moves the day,
    and `--daily=false` (or the other two) makes it a one-off task again.
  - Subtasks follow their parent's kind and day.
  - TUI: the form's "Repeat" field takes `n`, `d`, `w fri`, `m 5`; the list
    and details show how a task repeats. Web: a repeat select and a day
    input replace the daily checkbox (a form that only sends `daily` still
    works).
  - `--json` gains `Recurrence`, `RepeatOn` and `DonePeriod`.
  - Migration 011 adds `recurrence` and `repeat_on`; `is_daily` now means
    "recurs", and existing daily tasks become `recurrence = 'daily'`.

- **Optional time of day on a task.** `add`, `edit` and `dump --to-task`
  take `--time HH:MM` (24 hours); `edit --time ""` removes it.
  - It is shown on the task in the command line (`today`, `get-all`,
    subtasks, `--json` as `DueTime`), the TUI (list, details and the new
    Time field of the form) and the web app (a time input next to the date
    in the add form and the dump dialog).
  - Within the same priority, timed tasks come first, by time.
  - Stored in a new `due_time` column (migration 010); existing tasks have
    no time.

- **Discord reminder at a task's time.** A task with a time of day
  (`--time`) is announced when it comes, on its due day (daily tasks every
  day; weekly and monthly ones on their day), once and only if it is not
  done. The check runs every 5 minutes under `serve` (15 with the crontab
  entry); a reminder more than 2 hours late is dropped. The task's time
  also shows in the fixed reminders.

### Changed

- The command line and the TUI list show a one-off task's date without the
  hour it was created at (an undated task used to show the creation time as
  if it were a time of day); only a time you set is shown.

## [2.6.0] - 2026-10-20

### Added

- **Dump tab in the web app.** The dump entries saved from the phone (or
  anywhere else) can now be seen there, with their count in the tab bar.
  - `→` opens a dialog to turn an entry into a task: the name can be
    edited, and date, priority, daily and parent can be set. The entry
    leaves the dump once the task is created, and stays if it can't be
    (for example, the name is taken).
  - `×` deletes an entry after the browser's confirmation.

### Fixed

- **Akrasia crashed on start on Windows** ("message quotes.quote1 not
  found"). The translations are embedded files, whose paths always use
  `/`, but they were built with the OS separator (`\` on Windows), so
  nothing loaded. They load on every OS now.
- A missing translation no longer crashes the app on any OS: it falls
  back to English, then to the message id, and a load failure is reported
  on stderr. Quotes are looked up when shown, not at startup.
- On Windows, `serve install` (systemd) and `config notify install` (cron)
  stop with a message that points to WSL or the Task Scheduler.
- Web app: tapping a button with an empty name did nothing on phones (the
  browser's own check shows no message there). The app now shows "name is
  required".
- Web app: messages (status and errors) appear in the add bar at the
  bottom, so they are visible however far the list is scrolled.

### Documentation

- README: new "Platforms" section, covering Linux, Windows through WSL
  (recommended) and native Windows with its limits.
- CONTRIBUTING: steps to use the CD versioning.

### Development

- **Continuous integration** (GitHub Actions, `.github/workflows/ci.yml`)
  on every pull request and on pushes to `main` and `dev`:
  - Linux: gofmt, `go mod tidy`, `go vet` and the tests with the race
    detector;
  - Windows: build and tests.
- `.github` was in `.gitignore`, which would have kept any workflow out of
  the repository; it no longer is.
- Tests of Unix-only features skip themselves on Windows, with the reason.
- **Continuous delivery** (`.github/workflows/release.yml`): pushing a
  version tag (`v2.6.0`) builds the binaries for Linux, Windows and macOS
  (amd64 and arm64; Windows amd64 only), with a `checksums.txt`, and
  publishes a GitHub release whose notes are that version's section of this
  changelog. The release fails if the section is missing.

## [2.5.0] - 2026-10-20

### Added

- **`akrasia serve`: a web app for your phone.** Today and All tabs with
  subtasks and today's progress, a bar to add a task, a subtask or a dump
  entry, and a tap to complete a task with notes (and the confirmation for
  a parent with pending subtasks). Made to install on the home screen
  (manifest, icon and service worker), in light and dark mode, in English
  or Portuguese.
  - Reached over Tailscale: it listens on `127.0.0.1:8080` and
    `tailscale serve` publishes it in HTTPS to your devices only.
  - Every page needs an access token, entered once per device and kept in
    an HttpOnly, SameSite=Strict cookie (`akrasia serve token`, `--new`
    replaces it). Writes only accept htmx requests, which closes CSRF.
  - `akrasia serve install` makes it a systemd user service; while it
    runs, it also makes the daily backup and sends the reminders (the
    crontab entry is removed). `serve uninstall` removes it.
  - Built with Go templates and htmx (vendored, 0BSD): no Node, nothing
    to build besides `go build`.

### Changed

- The database opens in WAL mode and waits up to 5 seconds for another
  writer, so the CLI, the TUI and other processes can use it at the same
  time without "database is locked" errors. A `-wal` and a `-shm` file may
  appear next to `akrasia.db` while it is open.
- `restore` refuses to run while `akrasia serve` uses the database.

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

[Unreleased]: https://github.com/luis-octavius/akrasia/compare/dev...v.2.8.0
[2.8.0]: https://github.com/luis-octavius/akrasia/compare/2.8.0...v.2.7.1
[2.7.1]: https://github.com/luis-octavius/akrasia/compare/v.2.7.1...v.2.7.0
[2.7.0]: https://github.com/luis-octavius/akrasia/compare/v.2.6.0...v.2.7.0
[2.6.0]: https://github.com/luis-octavius/akrasia/compare/v.2.5.0...v.2.6.0
[2.5.0]: https://github.com/luis-octavius/akrasia/compare/v.2.4.0...v.2.5.0
[2.4.0]: https://github.com/luis-octavius/akrasia/compare/v.2.3.0...v.2.4.0
[2.3.0]: https://github.com/luis-octavius/akrasia/compare/v.2.2.0...v.2.3.0
[2.2.0]: https://github.com/luis-octavius/akrasia/compare/v.2.1.0...v.2.2.0
[2.1.0]: https://github.com/luis-octavius/akrasia/compare/v.2.0.0...v.2.1.0
[2.0.0]: https://github.com/luis-octavius/akrasia/compare/v.1.1.1...v.2.0.0
[1.1.1]: https://github.com/luis-octavius/akrasia/releases/tag/v.1.1.1
