package main

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/luis-octavius/akrasia/internal/database"
	"github.com/luis-octavius/akrasia/pkg/color"
	"github.com/luis-octavius/akrasia/pkg/emoji"
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

func validateTime(expiresAt time.Time) sql.NullTime {
	t := sql.NullTime{}

	if expiresAt.IsZero() {
		t.Valid = false
		return t
	}

	t.Time = expiresAt
	t.Valid = true

	return t
}

// parseTime receives the expireDate that is an argument
// from the command add. it returns a time that is valid
// for saving the todo at database properly
func parseTime(expireDate []string) (time.Time, error) {
	lenExpire := len(expireDate)

	// create a time with an added day
	t := time.Now()
	year, month, day := t.Date()

	switch lenExpire {
	// if expireDate is empty, than the todo is a daily todo
	case 0:
		day += 1
		forwardDay := time.Date(year, month, day+1, 0, 0, 0, 0, time.UTC)
		return forwardDay, nil
	case 1:
		parsedDay, _ := strconv.Atoi(expireDate[0])
		date := time.Date(year, month, parsedDay, 0, 0, 0, 0, time.UTC)
		return date, nil
	case 2:
		parsedDay, _ := strconv.Atoi(expireDate[0])
		parsedMonth, _ := strconv.Atoi(expireDate[1])
		date := time.Date(year, time.Month(parsedMonth), parsedDay, 0, 0, 0, 0, time.UTC)
		return date, nil
	case 3:
		// split day, month and hour
		// parse day string to int
		parsedDay, _ := strconv.Atoi(expireDate[0])
		// parse month string to time.Month
		parsedMonth, _ := strconv.Atoi(expireDate[1])
		date, err := parseHourLayout(year, parsedDay, time.Month(parsedMonth), expireDate[2])
		if err != nil {
			return time.Time{}, fmt.Errorf("Error in parsetime - %v", err)
		}
		return date, nil
	}

	// default
	return time.Time{}, nil
}

// parseHourLayout must receive a string representation of time
// in a layout '00:00:00' and then returns a time.Time
// that represents the string input
func parseHourLayout(year, day int, month time.Month, timeLayout string) (time.Time, error) {
	splitTime := strings.Split(timeLayout, ":")

	if len(splitTime) != 3 {
		return time.Time{}, fmt.Errorf("The time layout is not the expected, try with '00:00:00'")
	}

	hour, _ := strconv.Atoi(splitTime[0])
	min, _ := strconv.Atoi(splitTime[1])
	sec, _ := strconv.Atoi(splitTime[2])
	date := time.Date(year, month, day, hour, min, sec, 0, time.UTC)

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
func printTodo(todo database.Todo, colorName string) {
	todoTime := todo.ExpiresAt.Time.Format(time.RFC822)

	// // divide the date and time to construct a readable expiring date
	// _, month, day := todoTime.Date()
	// onlyTime := todoTime.Format(time.TimeOnly)

	var status string

	if todo.Concluded == true {
		status = "Done"
	} else {
		status = "Not done"
	}

	s := fmt.Sprintf("%v %v | %v\n%v | %v\n\n", emoji.Todo, todo.Name, todo.Description.String, todoTime, status)

	colorized, _ := color.ColorizeOutput(colorName, s)

	fmt.Println(colorized)
}

func addColorAndEmoji(pickedEmoji, colorName, text string) string {
	colorized, _ := color.ColorizeOutput(colorName, text)
	strWithEmoji := emoji.AddEmoji(pickedEmoji, colorized)
	return strWithEmoji
}
