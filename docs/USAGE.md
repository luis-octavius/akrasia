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

# backfill history for the last 30 days (default)
akrasia backfill-history
Backfilled 90 history entries for 3 daily task(s)

# backfill history for a specific task and 60 days back
akrasia backfill-history --task "Morning run" --days 60
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
```