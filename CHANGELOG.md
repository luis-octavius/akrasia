# Changelog

All notable changes to Akrasia are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and versions follow
[Semantic Versioning](https://semver.org/).

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

[2.0.0]: https://github.com/luis-octavius/akrasia/compare/v.1.1.1...dev
[1.1.1]: https://github.com/luis-octavius/akrasia/releases/tag/v.1.1.1
