package commands

import (
	"fmt"
	"log"
	"time"
)

func isDateBefore(date time.Time) (bool, error) {
	isBefore := date.Before(time.Now())
	if isBefore == true {
		return isBefore, fmt.Errorf("date %v is before right now - put a valid date", date)
	}

	return isBefore, nil
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
