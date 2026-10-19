## Interactive interface

Run `akrasia` with no arguments to open the TUI (`akrasia tui` does the same).
When stdin or stdout is not a terminal, such as in a pipe or a script, it
prints the help instead. Use `akrasia help` to see the help at any time.

Below the tabs, a bar shows today's progress (see `akrasia progress`).

| Key | Action |
| --- | --- |
| `↑` `↓` / `home` `end` | move / jump to top or bottom |
| `tab` `shift+tab` / `←` `→` | switch tabs (Today, All, Daily, Dump) |
| `space` | complete the selected task (asks for optional notes) |
| `a` | new task (on the Dump tab: new dump entry) |
| `x` | delete the selected task (asks for confirmation) |
| `enter` | toggle the details panel (streak and streak history for daily tasks) |
| `/` | filter by name; `esc` clears it |
| `p` | cycle the priority filter |
| `t` | switch to the next theme and save it |
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
 ce  check-expired    show tasks that have passed their expiration date
 cx  check-expiring   check todos that are expiring in 5 days
     completion       Generate the autocompletion script for the specified shell
 cfg config           manage settings (theme, language, TUI keymap)
 dbg debug            inspect a task's stored history and streak math (diagnostics)
 dn  delete-by-name   delete a todo by name
 dp  dump             store a named dump entry, or list stored entries
 fc  focus            show top 1-3 priority items to focus on right now
 ga  get-all          returns all the todos saved in storage
 gn  get-by-name      search for a task by name (case-insensitive, fuzzy match)
 gd  get-daily        show all daily tasks
     help             Help about any command
 sh  history          get the streak history of provided todo
     init             initialize akrasia app
 pg  progress         show today's progress (tasks done out of today's tasks)
 cs  streak           get the current streak of provided todo
 td  today            show a daily focus dashboard (overdue, due today, daily pending, expiring soon)
     tui              open the interactive terminal interface
 us  update-status    mark a task as completed and record it in history with optional notes
```

### Config

```
 k keymap      manage the TUI key layout        keymap [list|show|default|vim]
 l language    manage languages                 language [list|show|en|pt]
 t theme       manage color themes              theme [list|show|<theme-name>]
```

### Common flags

| Command | Flags |
| --- | --- |
| `add` | `--desc`, `--priority high\|medium\|low`, `--daily`, `--date dd` or `--date dd,mm` (default: tomorrow), `--name` (instead of the positional name) |
| `today` | `--only overdue\|today\|daily\|soon`, `--limit N`, `--priority`, `--json` |
| `focus` | `--limit 1-3` (default 3), `--priority` |
| `get-all` | `--priority` |
| `update-status` | `--notes "text"` (saved in the task's history for today) |
| `delete-by-name` | `--yes` (required) |
| `backfill-history` | `--days N` (default 30), `--task <name>` (default: every daily task) |

Names are matched case-insensitively and fuzzily: `akrasia get-by-name stend`
finds "Stendhal".

## Files and environment

| What | Where |
| --- | --- |
| Database | `$DB_PATH`, or `akrasia/akrasia.db` inside the OS config directory (`~/.config` on Linux) |
| Theme, language, keymap | `theme.json`, `lang.json`, `keymap.json` in `$XDG_CONFIG_HOME/akrasia` (default `~/.config/akrasia`) |
