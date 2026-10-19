## Usage

The examples below are real outputs. Colors are not shown, and the motivational
quote printed after some commands is shortened to one example.

### Interactive interface

```bash
akrasia        # or: akrasia tui
```

Use the arrows (or `j`/`k` with `akrasia config keymap vim`) to move, `space`
to complete, `a` to add, `x` to delete, `enter` for details, `t` to switch
themes and `?` for every key. See [Commands](COMMANDS.md#interactive-interface).

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

# --date takes the day (this month) or day,month (this year); without it the
# task is due tomorrow
akrasia add "Pay electricity bill" --priority medium --date 21,10
Task Pay electricity bill created successfully!
```

### Completing tasks

```bash
# notes are saved in the task's history for today
akrasia update-status "Morning run" --notes "5km in 28 minutes"
Updated successfully
```

Daily tasks count as done only for the day you complete them. One-off tasks stay
done.

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
akrasia today
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
  "ExpiringSoon": null
}

# the top items to work on, in order of urgency
akrasia focus --limit 2
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

A quick place for loose ideas that are not tasks yet.

```bash
akrasia dump "remember this"
Dump 'remember this' added

akrasia dump
remember this | 19 Oct 26 14:38 UTC
```

### Deleting

```bash
# --yes is required; the name is matched fuzzily, and the message shows which
# task was deleted
akrasia delete-by-name "Pay electricity" --yes
Todo Pay electricity bill deleted successfully
```

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
