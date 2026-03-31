package main

import (
	"database/sql"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/luis-octavius/akrasia/internal/database"
	"github.com/luis-octavius/akrasia/pkg/color"
)

func validateDescription(description string) sql.NullString {
	descriptionField := sql.NullString{}

	if description == "" {
		descriptionField.String = ""
		descriptionField.Valid = false
	} else {
		descriptionField.String = description
		descriptionField.Valid = true
	}

	return descriptionField
}

func parseDate(expireDate []int) (time.Time, error) {
	lenDate := len(expireDate)
	if lenDate > 2 {
		log.Fatal("Not enough arguments in date")
	}

	actualTime := time.Now()
	year, month, _ := actualTime.Date()
	date := time.Time{}

	switch lenDate {
	case 0:
		date = time.Now().AddDate(0, 0, 1)
		_, err := isDateBefore(date)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}
		return date, nil
	case 1:
		date := time.Date(year, month, expireDate[0], 0, 0, 0, 0, time.UTC)
		_, err := isDateBefore(date)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}
		return date, nil
	case 2:
		month = getMonthByNum(expireDate[1])
		date := time.Date(year, month, expireDate[0], 0, 0, 0, 0, time.UTC)
		_, err := isDateBefore(date)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}
		return date, nil
	}

	return date, nil
}

func checkIfTodoExpires(expiresAt time.Time) bool {
	actualDay := time.Now()

	diff := expiresAt.Sub(actualDay).String()
	hour, _, _ := strings.Cut(diff, "h")
	hourToInt, _ := strconv.Atoi(hour)
	if hourToInt <= (24 * 5) {
		return true
	}
	return false
}

// printTodo receives a Todo and create a readable output
func printTodo(todo database.Todo) {
	todoTime := todo.ExpiresAt.Format(time.RFC822)

	var status string

	if todo.Concluded == true {
		status = "Done"
	} else {
		status = "Not done"
	}

	s := fmt.Sprintf("%v | %v\n%v | %v\n\n", todo.Name, todo.Description.String, todoTime, status)

	if ok := checkIfTodoExpires(todo.ExpiresAt); !ok {
		color.MsgSuccess(s)
		return
	}

	color.MsgWarning(s)
}

func getMonthByNum(num int) time.Month {
	if num <= 0 || num > 12 {
		log.Fatal("month number is not between 1 and 12")
	}

	switch num {
	case 1:
		return time.January
	case 2:
		return time.February
	case 3:
		return time.March
	case 4:
		return time.April
	case 5:
		return time.May
	case 6:
		return time.June
	case 7:
		return time.July
	case 8:
		return time.August
	case 9:
		return time.September
	case 10:
		return time.October
	case 11:
		return time.November
	case 12:
		return time.December
	}

	return time.Now().Month()
}

func isDateBefore(date time.Time) (bool, error) {
	isBefore := date.Before(time.Now())
	if isBefore == true {
		return isBefore, fmt.Errorf("date %v is before right now - put a valid date", date)
	}

	return isBefore, nil
}

// filterTodosByPriority keeps only tasks matching the provided priority.
func filterTodosByPriority(todos []database.Todo, priorityFilter string) []database.Todo {
	if priorityFilter == "" {
		return todos
	}

	filtered := make([]database.Todo, 0, len(todos))
	for _, todo := range todos {
		if todo.Priority == priorityFilter {
			filtered = append(filtered, todo)
		}
	}

	return filtered
}

// buildTodayView classifies pending tasks into today-oriented sections.
func buildTodayView(todos []database.Todo, now time.Time) todayView {
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	dayEnd := dayStart.Add(24 * time.Hour)

	view := todayView{}

	for _, todo := range todos {
		if todo.Concluded {
			continue
		}

		if todo.IsDaily {
			view.Daily = append(view.Daily, todo)
			continue
		}

		if todo.ExpiresAt.Before(dayStart) {
			view.Overdue = append(view.Overdue, todo)
			continue
		}

		if !todo.ExpiresAt.Before(dayStart) && todo.ExpiresAt.Before(dayEnd) {
			view.DueToday = append(view.DueToday, todo)
			continue
		}

		if todo.ExpiresAt.Before(now.Add(5 * 24 * time.Hour)) {
			view.ExpiringSoon = append(view.ExpiringSoon, todo)
		}
	}

	sortTodayTodos(view.Overdue)
	sortTodayTodos(view.DueToday)
	sortTodayTodos(view.Daily)
	sortTodayTodos(view.ExpiringSoon)

	return view
}

// applyTodayOptions applies limit and section filters to a today view.
func applyTodayOptions(view todayView, opts todayOptions) todayView {
	if opts.Limit > 0 {
		view.Overdue = limitTodos(view.Overdue, opts.Limit)
		view.DueToday = limitTodos(view.DueToday, opts.Limit)
		view.Daily = limitTodos(view.Daily, opts.Limit)
		view.ExpiringSoon = limitTodos(view.ExpiringSoon, opts.Limit)
	}

	switch opts.Only {
	case "overdue":
		view.DueToday = nil
		view.Daily = nil
		view.ExpiringSoon = nil
	case "today":
		view.Overdue = nil
		view.Daily = nil
		view.ExpiringSoon = nil
	case "daily":
		view.Overdue = nil
		view.DueToday = nil
		view.ExpiringSoon = nil
	case "soon":
		view.Overdue = nil
		view.DueToday = nil
		view.Daily = nil
	}

	return view
}

// limitTodos truncates a slice to the requested size when needed.
func limitTodos(todos []database.Todo, limit int) []database.Todo {
	if limit <= 0 || len(todos) <= limit {
		return todos
	}

	return todos[:limit]
}

// sortTodayTodos orders tasks by priority first, then expiration date.
func sortTodayTodos(todos []database.Todo) {
	sort.Slice(todos, func(i, j int) bool {
		leftPriority := priorityRank(todos[i].Priority)
		rightPriority := priorityRank(todos[j].Priority)

		if leftPriority != rightPriority {
			return leftPriority < rightPriority
		}

		return todos[i].ExpiresAt.Before(todos[j].ExpiresAt)
	})
}

// priorityRank maps textual priority to a sortable numeric value.
func priorityRank(priority string) int {
	switch priority {
	case "high":
		return 1
	case "medium":
		return 2
	case "low":
		return 3
	default:
		return 4
	}
}

// printTodaySection renders one titled section from the today dashboard.
func printTodaySection(title string, todos []database.Todo) {
	if len(todos) == 0 {
		return
	}

	fmt.Printf("\n%s (%d)\n", title, len(todos))
	for _, todo := range todos {
		printTodo(todo)
	}
}
