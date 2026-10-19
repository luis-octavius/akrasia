# Changelog

All notable changes to Akrasia are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and versions follow
[Semantic Versioning](https://semver.org/).

## [Unreleased]

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

## [2.0.0] - unreleased

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

[Unreleased]: https://github.com/luis-octavius/akrasia/compare/v.1.1.1...dev
[2.0.0]: https://github.com/luis-octavius/akrasia/compare/v.1.1.1...dev
[1.1.1]: https://github.com/luis-octavius/akrasia/releases/tag/v.1.1.1
