Here you can see how you can contribute to the application in general and in other specific cases. 

## Table of Contents
- [Branches and Pull Requests](#branches-and-pull-requests)
- [Adding a command](#adding-a-command)
## Branches and Pull Requests

Every change should be made in its own branch, created from `main`. Please don't commit directly into the `main` branch.

1. Name your branch as `type/short-description`, using the same types as [Conventional Commits](https://www.conventionalcommits.org/):
   - `feat/` for new features or commands
   - `fix/` for bug fixes
   - `docs/` for documentation changes
   - `refactor/` for code changes that don't alter behavior
   - `test/` for adding or adjusting tests
   - `chore/` for maintenance tasks

   Use lowercase and hyphens in the description, e.g. `fix/streak-reset-bug` or `feat/export-command`.

2. Commit your changes, preferably following the Conventional Commits style as well.
3. Open a Pull Request against `main` and reference the related issue in the description.
4. I'll review it and may ask for adjustments before merging. 

## Adding a command 
The flow to add a command is simple. 

1. If it is a new functionality that the existing queries do not cover, you'll have to create one in `/internal/db/queries/`. If it is related to tasks, you'll have to put into the `todos.sql`, otherwise, put it in `todos_history.sql`. The shape is this: 

```sqlite
-- name: GetTodoByName :one 
SELECT * FROM todos 
where name LIKE '%';

```
2. Generate the new query(ies) in the code with `sqlc generate`.

3. Create a method for the `TaskManager` in `internal/tasks/task.go`. I'll leave a detailed example below:

```go 
// the signature to create a new method for the TaskManager 
func (tkm *TaskManager) AddTodo(name, description, priority string, isDaily bool, expiresAt time.Time) error {
    // here I am using a method to validate the Description input
	descriptionField := validateDescription(description)
	now := time.Now()

    // that is how you call the created method in step 1
	_, err := tkm.Queries.AddTodo(context.Background(), database.AddTodoParams{
		ID:           uuid.New(),
		Name:         name,
		Description:  descriptionField,
		CreatedAt:    now,
		UpdatedAt:    now,
		Concluded:    false,
		ExpiresAt:    expiresAt,
		Priority:     priority,
		IsDaily:      isDaily,
		HistorySince: sql.NullString{String: now.Format(time.DateOnly), Valid: true},
	})
	if err != nil {
        // use i18n to show an error based on the language of the system
        // if there isn't a message that represents what you want, you'll 
        // have to create one in all locales
		return fmt.Errorf(i18n.T("errorCreateTask"), err)
	}

    // color the output based on success
	color.MsgSuccess(fmt.Sprintf(i18n.T("createdTask"), name))

    // use, if you want, the method to generate a random philosophic quote
	generateRandomQuote()
	return nil
}
```

4. Now, you can create the command in `internal/commands/commands.go`: 

```go
// add creates a new task with optional metadata such as priority and daily mode.
var add = &cobra.Command{
	Use:     i18n.T("addUse"), // how to use it 
	Short:   i18n.T("addShort"), // a concise description
	Aliases: []string{"a"}, // aliases for the command
	Example: i18n.T("addExample"), // an example of the command
	Args: cobra.MaximumNArgs(2), // how many args the command accept (optional)

    // here it is where the magic happens, you will treat the args and validate them
	RunE: func(cmd *cobra.Command, args []string) error { 
		tkm, err := taskManagerFromContext(cmd.Context())
		if err != nil {
			return err
		}

		// Handle positional arguments and flags
		// Priority: positional args > flags (for backward compatibility)
		taskName := name
		taskDesc := description

		if len(args) > 0 {
			taskName = args[0]
		}
		if len(args) > 1 {
			taskDesc = args[1]
		}

		// Validate that name was provided
		if taskName == "" {
			return errors.New(i18n.T("errorTaskName"))
		}

		expiresAt, err := parseDate(date)
		if err != nil {
			return err
		}

		isDaily, err := cmd.Flags().GetBool("daily")

		if (err != nil) {
			return err
		}
        
        // usage of the created method in step 3
		err = tkm.AddTodo(taskName, taskDesc, priority, isDaily, expiresAt)
		if err != nil {
			return err
		}

		return nil
	},
}

```

5. Add the command in the map `commands` inside `init()`. 

6. Add flags in the created command. The cobra documentation for how to work with flags is [this](https://cobra.dev/docs/how-to-guides/working-with-flags/). A simple example: 

```go
	add.Flags().IntSliceVar(&date, "date", []int{}, i18n.T("addFlagDate"))
	add.Flags().StringVar(&name, "name", "", i18n.T("addFlagName"))
	add.Flags().StringVar(&priority, "priority", "", i18n.T("addFlagPriority"))
	add.Flags().Bool("daily", false, i18n.T("addFlagDaily"))
	add.Flags().StringVar(&description, "desc", "", i18n.T("addFlagDescription"))
```

## Adding an issue 
If you want to contribute with ideas and recommendations, just open an issue with your idea. I'll answer and will discuss it to see how, if it is a doable thing, we can achieve such thing. 


