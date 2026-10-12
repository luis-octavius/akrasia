## Usage

```bash
# add command with positional arguments (new style - simpler!)
akrasia add "Stendhal"
2026/01/04 11:38:31 Task Stendhal created successfully!

# add command with positional name and description
akrasia add "Stendhal" "Finish the book The Red and The Black" --priority high
2026/01/04 11:38:31 Task Stendhal created successfully!

# add command with flags (backward compatible)
akrasia add --name Stendhal --desc "Finish the book The Red and The Black" 13
2026/01/04 11:38:31 Todo Stendhal created successfully!

# add a daily task
akrasia add "Morning run" --daily --priority high
2026/03/08 08:15:20 Task Morning run created successfully!

# manage color themes
akrasia config theme list           # show available themes
Available themes:
  - default
  - high-contrast

akrasia config theme show           # show current theme
Current theme: default

akrasia config theme high-contrast  # switch to high-contrast theme for better accessibility
Theme set to: high-contrast

# mark task as done
akrasia update-status --name stendhal # case-insensitive, mark Todo as done
Stendhal | Finish the book The Red and The Black |
13 Feb 26 00:00 UTC | Done

# get all tasks with priority filter
akrasia get-all --priority high
Todos:

Stendhal | Finish the book The Red and The Black |
13 Feb 26 00:00 UTC | Done

# delete concluded tasks (requires confirmation)
akrasia delete-concluded --yes
Concluded Todos deleted successfully!

# backfill history for the last 30 days (default, all daily tasks)
akrasia backfill-history
  Morning run: .......... 30 days marked as completed
  Read 30min: .......... 30 days marked as completed
Backfilled 60 history entries for 2 daily task(s)

# backfill history for a specific task and 60 days back
akrasia backfill-history --task "Morning run" --days 60
  Morning run: .......... 60 days marked as completed
Backfilled 60 history entries for task 'Morning run'

# see what needs attention now
akrasia today --only overdue --limit 5 --priority high
TODAY FOCUS

OVERDUE (1)
Pay electricity bill | ...

DUE TODAY (2)
Prepare weekly report | ...

DAILY PENDING (1)
Morning run | ...

# pick top items to execute now
akrasia focus --limit 3 --priority high
FOCUS (3)
Pay electricity bill | ...
Prepare weekly report | ...
Morning run | ...

# get streak for a daily task
akrasia streak --name "Morning run"
Your current streak with Morning run is: 5

# get streak history  
akrasia history --name "Morning run"
1. Started Date: 2026-02-15 | End Date: 2026-02-20 | Total Days: 6
2. Started Date: 2026-01-10 | End Date: 2026-01-15 | Total Days: 6

# inspect exactly what the streak math sees (history window, per-row flags)
akrasia debug "Morning run"
DEBUG: Morning run
  id            : b4ff5267-a7e4-41cf-b589-5bcec01bc8e5
  is_daily      : true
  created_at    : 2026-10-01
  history_since : 2026-10-08

History (5 rows):
  2026-10-06  completed=true  notes=manual       in_streak=false
  2026-10-07  completed=true  notes=manual       in_streak=false
  2026-10-08  completed=true  notes=manual       in_streak=true
  2026-10-09  completed=true  notes=manual       in_streak=true
  2026-10-10  completed=true  notes=manual       in_streak=true

Current streak: 3 (counting history dates >= history_since = 2026-10-08)

# store a dump entry, or list them
akrasia dump "remember this"
Dump 'remember this' added

akrasia dump
remember this | 11 Oct 26 20:45 UTC
```