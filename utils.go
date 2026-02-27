package main

import (
	"database/sql"
	"fmt"
	"log"
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
func printTodo(todo database.Todo, colorName string) {
	todoTime := todo.ExpiresAt.Format(time.RFC822)

	var status string

	if todo.Concluded == true {
		status = "Done"
	} else {
		status = "Not done"
	}

	s := fmt.Sprintf("%v | %v\n%v | %v\n\n", todo.Name, todo.Description.String, todoTime, status)

	colorized, _ := color.ColorizeOutput(colorName, s)

	fmt.Println(colorized)
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
