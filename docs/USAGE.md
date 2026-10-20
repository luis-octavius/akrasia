## Usage

The examples below are real outputs. Colors are not shown, and the motivational
quote printed after some commands is shortened to one example.

### Interactive interface

```bash
akrasia        # or: akrasia tui
```

Use the arrows (or `j`/`k` with `akrasia config keymap vim`) to move, `space`
to complete, `a` to add, `e` to edit, `x` to delete, `enter` for details, `n`
for notes, `t` to switch themes and `?` for every key. Completing a task shows
a quote, just like the CLI does; `Q` shows one at any time. See
[Commands](COMMANDS.md#interactive-interface).

### Adding tasks

```bash
# name and description as positional arguments
akrasia add "Stendhal" "Finish the book The Red and The Black" --priority high --date 13,11
Task Stendhal created successfully!

"Self-control (enkrateia) is the foundation of virtue, and the first step is
to master your pleasures and appetites." - Aristotle, Nicomachean Ethics

# a daily task (habits have no due date; they reset every day)
akrasia add "Morning run" --daily --priority high
Task Morning run created successfully!

# flags instead of positional arguments
akrasia add --name "Read 30min" --daily
Task Read 30min created successfully!

# --date takes the day (this month), day,month (this year) or
# day,month,year; without it the task is due tomorrow. Today is allowed,
# past days are not.
akrasia add "Pay electricity bill" --priority medium --date 21,10
Task Pay electricity bill created successfully!

akrasia add "Trip to Lisbon" --date 5,1,2027
Task Trip to Lisbon created successfully!
```

### Completing tasks

```bash
# notes are saved in the task's history for today
akrasia update-status "Morning run" --notes "5km in 28 minutes"
Updated successfully
```

Daily tasks count as done only for the day you complete them. One-off tasks stay
done.

### Notes

Every completion can carry a note. `notes` turns them into a journal of the
streak, newest first:

```bash
akrasia notes rust
Notes · Estudar Rust (streak 4)
2026-10-19  aprendi sobre funções e modos de retorno
2026-10-18  backfilled
2026-10-17  backfilled
2026-10-16  backfilled

akrasia notes rust --limit 2   # only the last 2 days
```

Days recorded by `backfill-history` show as "backfilled"; days completed
without a note show "(no notes)".

There is one note per day. To rewrite today's note, complete the task again:
`akrasia update-status rust --notes "new text"` replaces it. In the TUI,
`space` on a task already done today opens today's note, ready to edit. In the TUI, the details panel (`enter`)
shows the last 3 notes and `n` opens all of them.

### Editing tasks

```bash
# only the flags you pass change; the history (and the streak) is kept
akrasia edit "Morning run" --name "Morning walk"
Task Morning walk updated

akrasia edit stend --priority low --desc "" --date 30,11
Task Stendhal updated

# turn a one-off task into a daily one, or back with --daily=false
akrasia edit Stendhal --daily
Task Stendhal updated

akrasia edit Stendhal --name "Morning walk"
Error: there is already a task called 'Morning walk'
```

In the TUI, press `e` on a task: the form opens prefilled.

### Subtasks

A task can be split into subtasks, one level deep. The parent shows how many
are done (`2/5`), and finishing the last one completes the parent.

```bash
akrasia add "Party" --date 25,10
akrasia add "Bake cake" --parent "Party"        # due with the parent
akrasia add "Invite friends" --parent "Party" --date 22,10

akrasia update-status "Bake cake"
akrasia update-status "Party/Invite friends"    # the full name, when needed

akrasia get-all
○ Party  1/2
  25 Oct 26 00:00 · Not done
    ✔ Bake cake · 25 Oct
    ○ Invite friends · 22 Oct
```

- A subtask follows its parent's daily mode and can't be due after it.
  Moving the parent's date earlier brings later subtasks along.
- Completing the parent itself while subtasks are pending completes them
  too, after a confirmation (`--yes` skips it, and scripts need it).
- Adding a subtask to a finished parent makes it pending again.
- Daily parent (a routine): each subtask is done once a day, and the parent
  closes the day, and its streak, when all of them are done.
- Subtasks have names unique within their parent. When the same name exists
  elsewhere, use `"Parent/Subtask"`: a bare name picks the top-level task,
  and a name shared by subtasks of several parents is refused with the
  full names to choose from.
- Deleting the parent deletes its subtasks.
- Today's progress, `today`, `focus` and the Discord reminders count a
  parent as one task, with its subtasks listed under it.

In the TUI, `s` adds a subtask to the selected task, `z` folds its subtasks
away, and `space` on a parent with pending subtasks asks before completing
them all.

### Listing tasks

```bash
akrasia get-all --priority high
Tasks:
✔ Morning run
  Priority: high

○ Stendhal
- Finish the book The Red and The Black
  13 Nov 26 00:00 · Not done
  Priority: high

akrasia get-daily
✔ Morning run
  Priority: high

○ Read 30min

# fuzzy, case-insensitive search
akrasia get-by-name stend
Name: Stendhal | Description: Finish the book The Red and The Black
Expires: Fri, 13 Nov 2026 00:00:00 UTC
```

### What needs attention now

```bash
# today's progress: every daily task, plus one-off tasks that are overdue,
# due today or completed today
akrasia progress
Today 1/2  ▌▌▌▌▌▌▌▌▌▌▌▌▌░░░░░░░░░░░░  50%

# the same line opens `today` and `focus`
akrasia today
Today 1/2  ▌▌▌▌▌▌▌▌▌▌▌▌▌░░░░░░░░░░░░  50%

TODAY FOCUS

DAILY PENDING (1)
○ Read 30min

EXPIRING SOON (1)
○ Pay electricity bill
  21 Oct 26 00:00 · Not done
  Priority: medium

# only one section, as JSON (for scripts)
akrasia today --only daily --json
{
  "Overdue": null,
  "DueToday": null,
  "Daily": [
    {
      "ID": "aac2460d-0e91-4042-a4be-3353d5b2b5ff",
      "Name": "Read 30min",
      "Description": { "String": "", "Valid": false },
      "Concluded": false,
      "ExpiresAt": "2026-10-20T14:38:24.665519395-03:00",
      "Priority": "",
      "IsDaily": true,
      "DoneToday": false
    }
  ],
  "ExpiringSoon": null,
  "Progress": { "Done": 1, "Total": 2 }
}

# the top items to work on, in order of urgency
akrasia focus --limit 2
Today 1/2  ▌▌▌▌▌▌▌▌▌▌▌▌▌░░░░░░░░░░░░  50%

FOCUS (2)
○ Read 30min

○ Pay electricity bill
  21 Oct 26 00:00 · Not done
  Priority: medium

# tasks due within 5 days, and tasks already past their date
akrasia check-expiring
akrasia check-expired
Your tasks are not fleeing. You have time, yet your focus must remain steadfast.
```

### Streaks

```bash
# record days you forgot to log (never today; use update-status for that)
akrasia backfill-history --task "Read 30min" --days 3
  Read 30min:  3 days marked as completed
Backfilled 3 history entries for task 'Read 30min'

akrasia streak "Read 30min"
Your current streak with Read 30min is: 3

akrasia history "Read 30min"
1. Started Date: 2026-10-16 | End Date: 2026-10-18 | Total Days: 3

# inspect exactly what the streak math sees
akrasia debug "Read 30min"
DEBUG: Read 30min
  id            : aac2460d-0e91-4042-a4be-3353d5b2b5ff
  is_daily      : true
  created_at    : 2026-10-19
  history_since : 2026-10-16

History (3 rows):
  2026-10-16  completed=true  notes=backfilled   in_streak=true
  2026-10-17  completed=true  notes=backfilled   in_streak=true
  2026-10-18  completed=true  notes=backfilled   in_streak=true

Current streak: 3 (counting history dates >= history_since = 2026-10-16)
```

### Dump

A quick place for loose ideas: some will become tasks, some never will.

```bash
akrasia dump "learn about sql indexes"
Dump 'learn about sql indexes' added

# the list is numbered; the numbers are what --delete and --to-task take
akrasia dump
1. app idea | 19 Oct 26 19:42
2. learn about sql indexes | 19 Oct 26 19:42
3. buy plants | 19 Oct 26 19:43

akrasia dump --delete 3
Dump 'buy plants' deleted

# turn an entry into a task (it leaves the dump); --priority, --daily and
# --date work like in `add`
akrasia dump --to-task 2 --priority high --date 30,11
Dump 'learn about sql indexes' is now a task
```

In the TUI, on the Dump tab: `a` adds, `x` deletes and `c` turns the selected
entry into a task through the new-task form.

### Deleting

```bash
# --yes is required; the name is matched fuzzily, and the message shows which
# task was deleted
akrasia delete-by-name "Pay electricity" --yes
Todo Pay electricity bill deleted successfully
```

### Backups

The database lives only on this machine. To avoid losing it, point Akrasia
at a folder a sync tool already copies elsewhere (Google Drive, Dropbox,
OneDrive, Syncthing). There is no account and no server: the sync tool does
the transport.

```bash
akrasia config backup dir ~/Dropbox/akrasia
```

From then on, the first command of each day (or opening the TUI) saves a
copy to that folder, and the newest 14 are kept. A failed backup only shows
a warning. The TUI status bar shows the date of the last one.

```bash
akrasia backup                         # a copy now
akrasia backup list                    # this machine's copies, newest first
akrasia config backup keep 30          # how many copies to keep
akrasia config backup show
akrasia config backup off              # stop; existing copies stay
```

Each machine writes under its own profile (the hostname by default), so a
personal and a work PC can share the folder without overwriting each
other's copies. `akrasia config backup profile work-pc` sets the name.

To restore, for example on a new PC:

```bash
akrasia config backup dir ~/Dropbox/akrasia
akrasia backup list --profile old-laptop
akrasia restore latest --profile old-laptop
# or a specific file
akrasia restore ~/Dropbox/akrasia/akrasia-old-laptop-2026-10-19_093000.db
```

`restore` checks the copy first, saves the current database to the backup
folder, then replaces it. It asks before replacing; scripts pass `--yes`.

### Discord reminders

Akrasia can post reminders to a Discord channel: tasks due in the next few
days, overdue ones and, in the evening, the daily tasks not done yet. It
uses a channel webhook, which is free and needs no bot.

1. In Discord, open the channel settings → Integrations → Webhooks → New
   Webhook → Copy Webhook URL.
2. To be mentioned (and get a push notification), turn on Developer Mode
   (Settings → Advanced), then right click your name → Copy User ID.

```bash
akrasia config notify discord https://discord.com/api/webhooks/…
akrasia config notify mention 123456789012345678   # or none
akrasia config notify test                         # posts a message now
akrasia config notify install                      # checks every 15 minutes
```

By default a reminder goes out at 10:00 and 19:00, with tasks due within 3
days and overdue ones; the 19:00 one also lists the daily tasks not done
today. Each time is sent once a day, and a day with nothing due sends
nothing. If the computer was off at a reminder time, it goes out at the
next check.

```bash
akrasia config notify days 5                 # 0 = only what's due today
akrasia config notify times 9:00,13:00,19:00
akrasia config notify daily-at 19:00         # which time lists daily tasks
akrasia config notify show
akrasia config notify off                    # also removes the crontab entry
```

`install` adds an entry to your crontab that runs `akrasia notify`, so it
needs cron (on Arch: `cronie`). Run `install` again after moving the
binary or changing `DB_PATH`. The webhook URL works like a password: anyone
who has it can post to the channel, so `notify.json` is readable only by
you.

### On your phone (`akrasia serve`)

`akrasia serve` runs a small web app on this computer: the Today and All
tabs, with subtasks and today's progress, a bar to add a task, a subtask
or a dump entry, and a tap to complete a task (with notes). The Dump tab
lists the dump entries: `→` turns one into a task (the name can be changed
first, and the entry leaves the dump) and `×` deletes it after a
confirmation. Messages, like an empty name, show above the add bar. The
database stays here: the phone only talks to it, so the computer has to be
on.

It is reached privately through [Tailscale](https://tailscale.com) (free
for personal use), which gives it an HTTPS address only your devices can
see. HTTPS is what lets the phone install it as an app.

```bash
akrasia serve install            # a systemd user service, started with the computer
tailscale serve --bg 8080        # https://<computer>.<tailnet>.ts.net for your devices
loginctl enable-linger $USER     # keep it running when you log out
akrasia serve token              # the access token
```

On the phone, with the Tailscale app connected, open the `ts.net` address,
enter the token once, and use the browser's "Add to Home screen".

- Every page needs the token; it is kept in the browser after the first
  login. `akrasia serve token --new` replaces it (every device then enters
  the new one). Scripts can send it as `Authorization: Bearer <token>`.
- It listens on `127.0.0.1:8080` only; `serve --addr` or `"addr"` in
  `serve.json` changes that.
- While it runs it also makes the daily backup and sends the Discord
  reminders, so `serve install` removes the crontab entry of
  `config notify install`.
- `restore` refuses to run while `serve` uses the database: stop it first
  with `systemctl --user stop akrasia`.
- `akrasia serve` alone runs it in the terminal (`ctrl+c` stops it), and
  `akrasia serve uninstall` removes the service.

### Settings

```bash
akrasia config theme list
Available themes:
  - default
  - high-contrast
  - purple
  - rose-pine
  - cyberpunk
  - catppuccin-mocha
  - tokyo-night
  - gruvbox
  - nord

akrasia config theme catppuccin-mocha   # or press t inside the TUI
akrasia config keymap vim               # adds j/k/g/G/h/l to the TUI
akrasia config language pt              # en or pt
```
