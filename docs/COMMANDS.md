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
