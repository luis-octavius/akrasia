## Interactive interface

Run `akrasia` with no arguments to open the TUI (`akrasia tui` does the same).
When stdin or stdout is not a terminal, such as in a pipe or a script, it
prints the help instead. Use `akrasia help` to see the help at any time.

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

```
Usage:
  akrasia [command]

Available Commands:
 a   add              create a new task with optional description, priority, and expiration date
 bf  backfill-history backfill missing daily task history entries
 ce  check-expired    show tasks that have passed their expiration date
 cx  check-expiring   check todos that are expiring in 5 days
     completion       Generate the autocompletion script for the specified shell
 cfg config           manage application settings (themes, preferences)
 dbg debug            inspect a task's stored history and streak math (diagnostics)
 dn  delete-by-name   delete a todo by name
 dc  delete-concluded delete all concluded todos
 dp  dump             store a named dump entry, or list stored entries
 fc  focus            show top 1-3 priority items to focus on right now
 ga  get-all          returns all the todos saved in storage
 gn  get-by-name      search for a task by name (case-insensitive, fuzzy match)
 gd  get-daily        show all daily tasks
     help             Help about any command
 sh  history          get the streak history of provided todo
     init             initialize akrasia app
 cs  streak           get the current streak of provided todo
 td  today            show a daily focus dashboard (overdue, due today, daily pending, expiring soon)
 us  update-status    mark a task as completed and record it in history with optional notes

Flags:
  -h, --help   help for akrasia

Use "akrasia [command] --help" for more information about a command.
```
