package server

import (
	"errors"
	"time"
)

func formatDate(dt time.Time) string {
	return dt.Format("Jan 2, 2006")
}

func parseDate(dtStr string) (time.Time, error) {
	formats := []string{
		"Jan 2, 2006",
		"Jan2, 2006",
		"Jan 2",
		"Jan2",
		"2006-01-02",
	}

	for _, format := range formats {
		t, err := time.Parse(format, dtStr)
		if err == nil {
			if format == "Jan 2" {
				currentYear := time.Now().Year()
				return time.Date(currentYear, t.Month(), t.Day(), 0, 0, 0, 0, time.UTC), nil
			}
			return t, nil
		}
	}

	return time.Time{}, errors.New("cannot parse date in any supported format")
}
