package main

import (
	"time"

	_ "time/tzdata"
)

func inTimeSpan(start, end, check time.Time) bool {
	if start.Before(end) {
		return !check.Before(start) && !check.After(end)
	}
	if start.Equal(end) {
		return check.Equal(start)
	}
	return !start.After(check) || !end.Before(check)
}

func timeIn(t time.Time, name string) (time.Time, error) {
	loc, err := time.LoadLocation(name)
	if err == nil {
		t = t.In(loc)
	}
	return t, err
}

// Steam downtimes usually happen on Tuesdays around 1PM to 3PM PST.
// Gotta check those numbers.
func checkForSteamUsualDowntime(currentTime time.Time) (bool, error) {
	if currentTime.Weekday().String() != "Tuesday" {
		return false, nil
	}

	valveTime, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		return true, err
	}

	// Steam downtime start in PST
	usualDowntimeStart := time.Date(2023, 04, 20, 13, 0, 0, 0, valveTime)
	// Steam downtime end in PST
	usualDowntimeEnd := time.Date(2023, 04, 20, 15, 0, 0, 0, valveTime)

	if !inTimeSpan(usualDowntimeStart, usualDowntimeEnd, currentTime) {
		return false, nil
	}

	return true, nil
}
