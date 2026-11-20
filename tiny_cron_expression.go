package tiny_cron_expression

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Schedule represents a parsed standard five-field cron expression.
type Schedule struct {
	minute     [60]bool
	hour       [24]bool
	dayOfMonth [32]bool // index 1..31
	month      [13]bool // index 1..12
	dayOfWeek  [7]bool  // index 0..6, Sunday=0
}

// Parse parses a standard five-field cron expression.
// Fields are: minute hour day-of-month month day-of-week.
// Supported syntax: numbers, ranges (a-b), lists (a,b,c), steps (*/n, a-b/n),
// and the wildcard '*'. Day-of-week accepts 0-6 (Sunday=0) and 7 as Sunday.
// Month names (JAN-DEC) and day names (SUN-SAT) are accepted case-insensitively.
func Parse(expr string) (*Schedule, error) {
	fields := strings.Fields(expr)
	if len(fields) != 5 {
		return nil, fmt.Errorf("cron: expected 5 fields, got %d", len(fields))
	}

	s := &Schedule{}

	minute, err := parseField(fields[0], 0, 59, nil)
	if err != nil {
		return nil, fmt.Errorf("cron: minute: %w", err)
	}
	copy(s.minute[:], minute)

	hour, err := parseField(fields[1], 0, 23, nil)
	if err != nil {
		return nil, fmt.Errorf("cron: hour: %w", err)
	}
	copy(s.hour[:], hour)

	dayOfMonth, err := parseField(fields[2], 1, 31, nil)
	if err != nil {
		return nil, fmt.Errorf("cron: day-of-month: %w", err)
	}
	for i := 1; i <= 31; i++ {
		s.dayOfMonth[i] = dayOfMonth[i]
	}

	month, err := parseField(fields[3], 1, 12, monthNames)
	if err != nil {
		return nil, fmt.Errorf("cron: month: %w", err)
	}
	for i := 1; i <= 12; i++ {
		s.month[i] = month[i]
	}

	dayOfWeek, err := parseField(fields[4], 0, 6, dayNames)
	if err != nil {
		return nil, fmt.Errorf("cron: day-of-week: %w", err)
	}
	for i := 0; i <= 6; i++ {
		s.dayOfWeek[i] = dayOfWeek[i]
	}
	// Normalize 7 to 0 (Sunday).
	if dayOfWeek[7] {
		s.dayOfWeek[0] = true
	}

	return s, nil
}

var monthNames = map[string]int{
	"JAN": 1, "FEB": 2, "MAR": 3, "APR": 4, "MAY": 5, "JUN": 6,
	"JUL": 7, "AUG": 8, "SEP": 9, "OCT": 10, "NOV": 11, "DEC": 12,
}

var dayNames = map[string]int{
	"SUN": 0, "MON": 1, "TUE": 2, "WED": 3, "THU": 4, "FRI": 5, "SAT": 6,
}

// parseField parses a single cron field. The returned slice is indexed by value
// from min to max. For day-of-week, index 7 is also populated for Sunday.
func parseField(field string, min, max int, names map[string]int) ([]bool, error) {
	result := make([]bool, max+2)
	parts := strings.Split(field, ",")
	for _, part := range parts {
		if err := parsePart(part, min, max, names, result); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func parsePart(part string, min, max int, names map[string]int, result []bool) error {
	if part == "" {
		return fmt.Errorf("empty field")
	}

	step := 1
	if idx := strings.Index(part, "/"); idx >= 0 {
		stepStr := part[idx+1:]
		part = part[:idx]
		s, err := strconv.Atoi(stepStr)
		if err != nil || s <= 0 {
			return fmt.Errorf("invalid step %q", stepStr)
		}
		step = s
	}

	var lo, hi int
	switch {
	case part == "*":
		lo, hi = min, max
	case strings.Contains(part, "-"):
		idx := strings.Index(part, "-")
		loStr, hiStr := part[:idx], part[idx+1:]
		var err error
		lo, err = parseValue(loStr, min, max, names)
		if err != nil {
			return err
		}
		hi, err = parseValue(hiStr, min, max, names)
		if err != nil {
			return err
		}
		if lo > hi {
			return fmt.Errorf("range start %d greater than end %d", lo, hi)
		}
	default:
		var err error
		lo, err = parseValue(part, min, max, names)
		if err != nil {
			return err
		}
		hi = lo
	}

	for v := lo; v <= hi; v += step {
		result[v] = true
	}
	return nil
}

func parseValue(s string, min, max int, names map[string]int) (int, error) {
	if names != nil {
		if v, ok := names[strings.ToUpper(s)]; ok {
			return v, nil
		}
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid value %q", s)
	}
	// Special case: day-of-week allows 7 as Sunday.
	if max == 6 && n == 7 {
		return 7, nil
	}
	if n < min || n > max {
		return 0, fmt.Errorf("value %d out of range [%d,%d]", n, min, max)
	}
	return n, nil
}

// Matches reports whether the schedule matches the given time.
func (s *Schedule) Matches(t time.Time) bool {
	if !s.minute[t.Minute()] {
		return false
	}
	if !s.hour[t.Hour()] {
		return false
	}
	if !s.month[int(t.Month())] {
		return false
	}
	// Cron semantics: if both day-of-month and day-of-week are restricted
	// (not '*'), the match is the union (OR). If one is '*', it must match.
	// We detect '*' by checking whether the field is fully populated.
	domRestricted := !allTrue(s.dayOfMonth[1:])
	dowRestricted := !allTrue(s.dayOfWeek[:])

	domMatch := s.dayOfMonth[t.Day()]
	dowMatch := s.dayOfWeek[int(t.Weekday())]

	switch {
	case domRestricted && dowRestricted:
		return domMatch || dowMatch
	case domRestricted:
		return domMatch
	case dowRestricted:
		return dowMatch
	default:
		return true
	}
}

func allTrue(b []bool) bool {
	for _, v := range b {
		if !v {
			return false
		}
	}
	return true
}

// Next returns the next count run times strictly after the given start time.
// The start time is not included in the result. The returned times are in the
// same location as the start time.
func (s *Schedule) Next(start time.Time, count int) []time.Time {
	if count <= 0 {
		return nil
	}
	result := make([]time.Time, 0, count)
	// Start from the next minute boundary strictly after start.
	t := start.Truncate(time.Minute).Add(time.Minute)
	for len(result) < count {
		if s.Matches(t) {
			result = append(result, t)
		}
		t = t.Add(time.Minute)
	}
	return result
}
