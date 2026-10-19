## Interactive interface

Run `akrasia` with no arguments to open the TUI (`akrasia tui` does the same).
When stdin or stdout is not a terminal, such as in a pipe or a script, it
prints the help instead. Use `akrasia help` to see the help at any time.

Below the tabs, a bar shows today's progress (see `akrasia progress`).

| Key | Action |
| --- | --- |
| `↑` `↓` / `home` `end` | move / jump to top or bottom |
| `tab` `shift+tab` / `←` `→` | switch tabs (Today, All, Daily, Dump) |
| `space` | complete the selected task (asks for optional notes, then shows a quote; `esc` closes it). On a task already done today, it opens today's note to rewrite it |
| `a` | new task (on the Dump tab: new dump entry) |
| `e` | edit the selected task (the form opens prefilled; only what you change is saved) |
| `x` | delete the selected task or dump entry (asks for confirmation) |
| `c` | on the Dump tab: turn the selected entry into a task (opens the new-task form with its text) |
| `enter` | toggle the details panel (last notes, plus streak and streak history for daily tasks) |
| `n` | open every note of the selected task in a scrollable window (`esc` closes) |
| `/` | filter by name; `esc` clears it |
| `p` | cycle the priority filter |
| `t` | switch to the next theme and save it |
| `Q` | show a random quote |
| `r` | reload the tab |
| `?` | show every key |
| `q` | quit |

Vim keys (`j` `k` `g` `G` `h` `l`) are added on top of the arrows with
`akrasia config keymap vim`; `akrasia config keymap default` turns them off.

Themes: `default`, `high-contrast`, `purple`, `rose-pine`, `cyberpunk`,
`catppuccin-mocha`, `tokyo-night`, `gruvbox` and `nord`. Pick one with
`akrasia config theme <name>` or press `t` inside the TUI. Hex colors are
adapted automatically on terminals without true color.

## Commands

Every command has a short alias (left column). `akrasia <command> --help`
shows its flags and examples.

```
Usage:
  akrasia [flags]
  akrasia [command]

Available Commands:
 a   add              create a new task with optional description, priority, and expiration date
 bf  backfill-history backfill missing daily task history entries
 bk  backup           copy the database to the backup folder now
 ce  check-expired    show tasks that have passed their expiration date
 cx  check-expiring   check todos that are expiring in 5 days
     completion       Generate the autocompletion script for the specified shell
 cfg config           manage settings (theme, language, TUI keymap, backups)
 dbg debug            inspect a task's stored history and streak math (diagnostics)
 dn  delete-by-name   delete a todo by name
 dp  dump             store a dump entry, list them, delete one or turn one into a task
 e   edit             change the name, description, priority, daily mode or date of a task
 fc  focus            show top 1-3 priority items to focus on right now
 ga  get-all          returns all the todos saved in storage
 gn  get-by-name      search for a task by name (case-insensitive, fuzzy match)
 gd  get-daily        show all daily tasks
     help             Help about any command
 sh  history          get the streak history of provided todo
     init             initialize akrasia app
 nt  notes            show the notes written each day a task was completed
 pg  progress         show today's progress (tasks done out of today's tasks)
     restore          replace the database with a backup (the current one is saved first)
 cs  streak           get the current streak of provided todo
 td  today            show a daily focus dashboard (overdue, due today, daily pending, expiring soon)
     tui              open the interactive terminal interface
 us  update-status    mark a task as completed and record it in history with optional notes
```

### Config

```
 b backup      set the backup folder, copies and profile   backup [show|off|dir <path>|keep <n>|profile <name>]
 k keymap      manage the TUI key layout        keymap [list|show|default|vim]
 l language    manage languages                 language [list|show|en|pt]
 t theme       manage color themes              theme [list|show|<theme-name>]
```

### Common flags

| Command | Flags |
| --- | --- |
| `add` | `--desc`, `--priority high\|medium\|low`, `--daily`, `--date dd`, `--date dd,mm` or `--date dd,mm,yyyy` (default: tomorrow), `--name` (instead of the positional name) |
| `edit` | `--name`, `--desc` (`""` clears it), `--priority` (`""` clears it), `--daily=true\|false`, `--date dd`, `dd,mm` or `dd,mm,yyyy`; only the flags you pass are changed |
| `today` | `--only overdue\|today\|daily\|soon`, `--limit N`, `--priority`, `--json` |
| `focus` | `--limit 1-3` (default 3), `--priority` |
| `get-all` | `--priority` |
| `update-status` | `--notes "text"` (saved in the task's history for today; completing again the same day replaces it) |
| `dump` | `--delete N`, `--to-task N` (with optional `--priority`, `--daily`, `--date`); `N` is the number `akrasia dump` prints |
| `notes` | `--limit N` days (default 30, `0` = all) |
| `delete-by-name` | `--yes` (required) |
| `backup list` | `--profile <name>` (another machine's copies) |
| `restore` | `--yes` (skip the question; required in scripts), `--profile <name>` (with `latest`) |
| `backfill-history` | `--days N` (default 30), `--task <name>` (default: every daily task) |

Names are matched case-insensitively and fuzzily: `akrasia get-by-name stend`
finds "Stendhal".

## Files and environment

| What | Where |
| --- | --- |
| Database | `$DB_PATH`, or `akrasia/akrasia.db` inside the OS config directory (`~/.config` on Linux) |
| Theme, language, keymap, backups | `theme.json`, `lang.json`, `keymap.json`, `backup.json` in `$XDG_CONFIG_HOME/akrasia` (default `~/.config/akrasia`) |
| Backups | `akrasia-<profile>-<yyyy-mm-dd_hhmmss>.db` in the folder set with `config backup dir` |
