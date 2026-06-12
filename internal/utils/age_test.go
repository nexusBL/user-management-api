package utils

import (
	"testing"
	"time"
)

func TestCalculateAge(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		dob      time.Time
		today    time.Time
		expected int
	}{
		{
			name:     "birthday today",
			dob:      time.Date(2000, time.June, 12, 0, 0, 0, 0, time.UTC),
			today:    time.Date(2026, time.June, 12, 9, 30, 0, 0, time.UTC),
			expected: 26,
		},
		{
			name:     "birthday tomorrow",
			dob:      time.Date(2000, time.June, 13, 0, 0, 0, 0, time.UTC),
			today:    time.Date(2026, time.June, 12, 9, 30, 0, 0, time.UTC),
			expected: 25,
		},
		{
			name:     "birthday yesterday",
			dob:      time.Date(2000, time.June, 11, 0, 0, 0, 0, time.UTC),
			today:    time.Date(2026, time.June, 12, 9, 30, 0, 0, time.UTC),
			expected: 26,
		},
		{
			name:     "leap year birthday before march first in non leap year",
			dob:      time.Date(2004, time.February, 29, 0, 0, 0, 0, time.UTC),
			today:    time.Date(2025, time.February, 28, 10, 0, 0, 0, time.UTC),
			expected: 20,
		},
		{
			name:     "leap year birthday on march first in non leap year",
			dob:      time.Date(2004, time.February, 29, 0, 0, 0, 0, time.UTC),
			today:    time.Date(2025, time.March, 1, 10, 0, 0, 0, time.UTC),
			expected: 21,
		},
		{
			name:     "leap year birthday on leap day",
			dob:      time.Date(2004, time.February, 29, 0, 0, 0, 0, time.UTC),
			today:    time.Date(2024, time.February, 29, 10, 0, 0, 0, time.UTC),
			expected: 20,
		},
		{
			name:     "future dob returns zero defensively",
			dob:      time.Date(2030, time.January, 1, 0, 0, 0, 0, time.UTC),
			today:    time.Date(2026, time.June, 12, 9, 30, 0, 0, time.UTC),
			expected: 0,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			actual := CalculateAge(testCase.dob, testCase.today)
			if actual != testCase.expected {
				t.Fatalf("expected age %d, got %d", testCase.expected, actual)
			}
		})
	}
}
