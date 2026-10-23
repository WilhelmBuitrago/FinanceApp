package app

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// monthByAbbr maps the month abbreviations Nu statements use (Spanish and
// English) to month numbers. It is the Go counterpart of the Python
// _MONTH_MAP: both must agree, and the tests pin that they do.
var monthByAbbr = map[string]int{
	"ene": 1, "jan": 1,
	"feb": 2,
	"mar": 3,
	"abr": 4, "apr": 4,
	"may": 5,
	"jun": 6,
	"jul": 7,
	"ago": 8, "aug": 8,
	"sep": 9, "set": 9,
	"oct": 10,
	"nov": 11,
	"dic": 12, "dec": 12,
}

// rawDayMonthRe reads the day+month at the start of a raw movement date,
// tolerating trailing decorations such as the year ("14 abr 2026").
var rawDayMonthRe = regexp.MustCompile(`(\d{1,2})\s*([A-Za-z]{3})`)

// parseRawDayMonth returns the day and month a raw date carries. ok=false
// means the raw holds no recognizable day+month.
func parseRawDayMonth(raw string) (day, month int, ok bool) {
	m := rawDayMonthRe.FindStringSubmatch(strings.ToLower(raw))
	if m == nil {
		return 0, 0, false
	}
	d, err := strconv.Atoi(m[1])
	if err != nil {
		return 0, 0, false
	}
	mm, known := monthByAbbr[m[2]]
	if !known {
		return 0, 0, false
	}
	return d, mm, true
}

// lastDayOfMonth handles leap years: 29 February exists only when the year
// allows it.
func lastDayOfMonth(year, month int) int {
	return time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC).AddDate(0, 1, -1).Day()
}

// PeriodRangeISO returns the first and last ISO dates of the given month.
func PeriodRangeISO(year, month int) (start, end string) {
	first := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	last := first.AddDate(0, 1, -1)
	return first.Format("2006-01-02"), last.Format("2006-01-02")
}

// DeriveMovementDate checks a parked raw date against the human-confirmed
// month and year: the day comes from the raw, the month must be the
// confirmed one, and the day must exist in that month of that year. ok=false
// means the movement does not fit the period and stays parked for human
// review. The movement date is checked against the period, never the
// reverse: the confirmed period is the truth.
func DeriveMovementDate(raw string, year, month int) (iso string, ok bool) {
	day, rawMonth, valid := parseRawDayMonth(raw)
	if !valid || rawMonth != month {
		return "", false
	}
	if day < 1 || day > lastDayOfMonth(year, month) {
		return "", false
	}
	return fmt.Sprintf("%04d-%02d-%02d", year, month, day), true
}
