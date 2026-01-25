package main

import (
	"testing"
	"time"
)

func TestParseDateWithoutArgs(t *testing.T) {
	oneDay, err := time.ParseDuration("24h")
	if err != nil {
		t.Fatal("Error parsing duration: ", err)
	}

	expected := time.Now().Add(oneDay).Day()
	result, err := parseDate([]int{})
	dayResult := result.Day()
	if err != nil {
		t.Fatal("Error in parseTime: ", err)
	}

	if expected != dayResult {
		t.Errorf("expected %v, got %v", expected, dayResult)
	}
}
