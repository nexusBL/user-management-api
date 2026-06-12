package utils

import "time"

func CalculateAge(dob time.Time, today time.Time) int {
	normalizedDOB := normalizeDate(dob)
	normalizedToday := normalizeDate(today)

	age := normalizedToday.Year() - normalizedDOB.Year()
	birthdayThisYear := time.Date(
		normalizedToday.Year(),
		normalizedDOB.Month(),
		normalizedDOB.Day(),
		0,
		0,
		0,
		0,
		time.UTC,
	)

	if normalizedToday.Before(birthdayThisYear) {
		age--
	}

	if age < 0 {
		return 0
	}

	return age
}

func normalizeDate(value time.Time) time.Time {
	value = value.UTC()
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}
