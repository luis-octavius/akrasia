## Commands

```
Usage:
  akrasia [command]

Available Commands:

  add                create a new task with optional description, priority, and expiration date
  backfill-history   backfill missing daily task history entries (--task to target one, --days for range)
  check-expired      show tasks that have passed their expiration date
  check-expiring     check todos that are expiring in 5 days
  completion         Generate the autocompletion script for the specified shell
  config             manage application settings (themes, preferences)
  create-cron        create the cron job to update daily tasks
  delete-concluded   delete all concluded todos
  delete-by-name     delete a todo by name
  focus              show top 1-3 priority items to focus on right now
  get-all            returns all the todos saved in storage
  get-by-name        search for a task by name (case-insensitive, fuzzy match)
  get-daily          show all daily tasks
  help               Help about any command
  history            get the streak history of provided todo
  init               initialize akrasia app
  streak             get the current streak of provided todo
  today              show a daily focus dashboard (overdue, due today, daily pending, expiring soon)
  update-daily       update daily todos (used by cron, not intended for manual use)
  update-status      mark a task as completed and record it in history with optional notes

Flags:
  -h, --help   help for akrasia

Use "akrasia [command] --help" for more information about a command.
```

