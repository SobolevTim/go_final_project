package task

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// NextDate function is used to calculate the next date based on the current date, date, and repeat value
//
// now: current date
// date: date value in format "YYYYMMDD"
// repeat: repeat value in format "d 1", "y", "w 1,2,3", "m -1,1,2 1,2,3" (d: day, y: year, w: week, m: month)
//
// return: next date in format "YYYYMMDD" if repeat is empty, return "-1"
func NextDate(now time.Time, date string, repeat string) (string, error) {
	if repeat == "" {
		return "", errors.New("repeat is empty")
	}
	dateParse, err := time.Parse("20060102", date)
	if err != nil {
		return "", fmt.Errorf("date is invalid: %s, err: %v", date, err)
	}
	repeatSlice := strings.Split(repeat, " ")
	if len(repeatSlice) == 0 {
		return "-1", nil
	}
	switch repeatSlice[0] {
	case "d":
		if len(repeatSlice) != 2 {
			return "", errors.New("repeat is invalid")
		}
		// validate repeat day
		day, err := strconv.Atoi(repeatSlice[1])
		if err != nil || day <= 0 || day > 400 { // according to the terms, the number of days cannot be more than 400
			return "", fmt.Errorf("invalid repeat day value: %v", repeatSlice[1])
		}

		// calculate next date
		if dateParse.After(now) {
			dateParse = dateParse.AddDate(0, 0, day)
		} else {
			for dateParse.Before(now) {
				dateParse = dateParse.AddDate(0, 0, day)
			}
		}
		return dateParse.Format("20060102"), nil
	case "y":
		// validate repeat year
		if len(repeatSlice) != 1 {
			return "", errors.New("repeat is invalid")
		}

		// calculate next date
		if dateParse.After(now) {
			dateParse = dateParse.AddDate(1, 0, 0)
		} else {
			for dateParse.Before(now) {
				dateParse = dateParse.AddDate(1, 0, 0)
			}
		}
		return dateParse.Format("20060102"), nil
	case "w":
		// validate repeat day of week
		if len(repeatSlice) != 2 {
			return "", errors.New("repeat is invalid")
		}
		var dayOfWeekSlice []int

		// conversion day of week to int value (0-6) 0 = Sunday, 1 = Monday, ..., 6 = Saturday
		for _, day := range strings.Split(repeatSlice[1], ",") {
			switch day {
			case "1", "2", "3", "4", "5", "6":
				dayOfWeekSlice = append(dayOfWeekSlice, int(day[0]-'0'))
			case "7":
				dayOfWeekSlice = append(dayOfWeekSlice, 0)
			default:
				return "", fmt.Errorf("invalid repeat day of week value: %v", repeatSlice[1])
			}
		}
		// sort day of week
		sort.Ints(dayOfWeekSlice)

		// Calculate next date directly
		var currentWeekday int
		var after bool
		if dateParse.After(now) {
			currentWeekday = int(dateParse.Weekday())
			after = true
		} else {
			currentWeekday = int(now.Weekday())
		}
		minDays := 7
		for _, targetDay := range dayOfWeekSlice {
			diff := (targetDay - currentWeekday + 7) % 7
			if diff == 0 {
				diff = 7 // Skip to next week if the day is today
			}
			if diff < minDays {
				minDays = diff
			}
		}
		var nextDate time.Time
		if after {
			nextDate = dateParse.AddDate(0, 0, minDays)
		} else {
			nextDate = now.AddDate(0, 0, minDays)
		}
		return nextDate.Format("20060102"), nil
	case "m":
		// validate repeat month
		if len(repeatSlice) < 2 {
			return "", errors.New("repeat is invalid")
		}

		// convert day to int value
		var daySlice []int
		for _, d := range strings.Split(repeatSlice[1], ",") {
			day, err := strconv.Atoi(d)
			if err != nil || (day < -2 || day > 31 || day == 0) { // days must be between 1 and 31 and -1, -2
				return "", fmt.Errorf("invalid repeat day value: %v", d)
			}
			daySlice = append(daySlice, day)
		}
		sort.Ints(daySlice)
		// convert month to int value
		var monthSlice []int
		var monthRepeat bool
		if len(repeatSlice) == 3 {
			monthRepeat = true
			for _, m := range strings.Split(repeatSlice[2], ",") {
				month, err := strconv.Atoi(m)
				if err != nil || month < 1 || month > 12 { // months must be between 1 and 12
					return "", fmt.Errorf("invalid repeat month value: %v", m)
				}
				monthSlice = append(monthSlice, month)
			}
		}
		sort.Ints(monthSlice)
		if now.After(dateParse) {
			dateParse = now
		}
		// calculate next date
		if !monthRepeat {
			dateParse = calculateNextDay(dateParse, daySlice, false)
		} else {
			dateParse = calculateNextMonth(dateParse, monthSlice)
			dateParse = calculateNextDay(dateParse, daySlice, true)

		}
		return dateParse.Format("20060102"), nil

	default:
		return "", errors.New("repeat is invalid")
	}
}

// calculateNextDay function is used to calculate the next day based on the current date and daySlice
func calculateNextDay(date time.Time, daySlice []int, isNextMonth bool) time.Time {
	firstDayOfMonth := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
	lastDayOfMonth := firstDayOfMonth.AddDate(0, 1, -1)
	daysInMonth := lastDayOfMonth.Day()
	var currentDayOfMotnh []int
	for _, day := range daySlice {
		if day > 0 {
			if day > daysInMonth {
				continue
			}
			currentDayOfMotnh = append(currentDayOfMotnh, day)
		} else {
			day = daysInMonth + day + 1
			currentDayOfMotnh = append(currentDayOfMotnh, day)
		}
	}
	sort.Ints(currentDayOfMotnh)
	if len(currentDayOfMotnh) == 0 {
		return calculateNextDay(time.Date(date.Year(), date.Month()+1, date.Day(), 0, 0, 0, 0, date.Location()), daySlice, true)
	}
	for _, day := range currentDayOfMotnh {
		if isNextMonth {
			if date.Day() <= day {
				return time.Date(date.Year(), date.Month(), day, 0, 0, 0, 0, date.Location())
			}
		}
		if date.Day() < day {
			return time.Date(date.Year(), date.Month(), day, 0, 0, 0, 0, date.Location())
		}
	}
	if date.Month() == 12 {
		return time.Date(date.Year()+1, 1, currentDayOfMotnh[0], 0, 0, 0, 0, date.Location())
	} else {
		return time.Date(date.Year(), date.Month()+1, currentDayOfMotnh[0], 0, 0, 0, 0, date.Location())
	}
}

// calculateNextMonth function is used to calculate the next month based on the current date and monthSlice
func calculateNextMonth(date time.Time, monthSlice []int) time.Time {
	for _, month := range monthSlice {
		if month > int(date.Month()) {
			return time.Date(date.Year(), time.Month(month), 1, 0, 0, 0, 0, date.Location())
		}
	}
	return time.Date(date.Year()+1, time.Month(monthSlice[0]), 1, 0, 0, 0, 0, date.Location())
}
