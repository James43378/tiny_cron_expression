package tiny_cron_expression

import (
	"testing"
	"time"
)

func mustParse(t *testing.T, expr string) *Schedule {
	t.Helper()
	s, err := Parse(expr)
	if err != nil {
		t.Fatalf("Parse(%q) error: %v", expr, err)
	}
	return s
}

func TestParseValid(t *testing.T) {
	exprs := []string{
		"* * * * *",
		"0 0 * * *",
		"*/5 * * * *",
		"0 12 * * MON",
		"0 0 1 JAN *",
		"15 10 1,15 * *",
		"0 0 * * 0",
		"0 0 * * 7",
		"0-30/10 * * * *",
	}
	for _, expr := range exprs {
		if _, err := Parse(expr); err != nil {
			t.Errorf("Parse(%q) unexpected error: %v", expr, err)
		}
	}
}

func TestParseInvalid(t *testing.T) {
	exprs := []string{
		"",
		"* * * *",
		"* * * * * *",
		"60 * * * *",
		"* 24 * * *",
		"* * 0 * *",
		"* * * 13 *",
		"* * * * 8",
		"*/0 * * * *",
		"5-1 * * * *",
		"abc * * * *",
	}
	for _, expr := range exprs {
		if _, err := Parse(expr); err == nil {
			t.Errorf("Parse(%q) expected error, got nil", expr)
		}
	}
}

func TestNextSimple(t *testing.T) {
	s := mustParse(t, "* * * * *")
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	got := s.Next(start, 3)
	want := []time.Time{
		time.Date(2024, 1, 1, 0, 1, 0, 0, time.UTC),
		time.Date(2024, 1, 1, 0, 2, 0, 0, time.UTC),
		time.Date(2024, 1, 1, 0, 3, 0, 0, time.UTC),
	}
	if len(got) != len(want) {
		t.Fatalf("Next returned %d times, want %d", len(got), len(want))
	}
	for i := range want {
		if !got[i].Equal(want[i]) {
			t.Errorf("Next[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestNextSteps(t *testing.T) {
	s := mustParse(t, "*/15 * * * *")
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	got := s.Next(start, 4)
	want := []time.Time{
		time.Date(2024, 1, 1, 0, 15, 0, 0, time.UTC),
		time.Date(2024, 1, 1, 0, 30, 0, 0, time.UTC),
		time.Date(2024, 1, 1, 0, 45, 0, 0, time.UTC),
		time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC),
	}
	for i := range want {
		if !got[i].Equal(want[i]) {
			t.Errorf("Next[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestNextDayOfWeek(t *testing.T) {
	s := mustParse(t, "0 0 * * MON")
	// 2024-01-01 is a Monday.
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	got := s.Next(start, 2)
	want := []time.Time{
		time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
	}
	for i := range want {
		if !got[i].Equal(want[i]) {
			t.Errorf("Next[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestNextDayOfMonth(t *testing.T) {
	s := mustParse(t, "0 0 1 * *")
	start := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	got := s.Next(start, 2)
	want := []time.Time{
		time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
	}
	for i := range want {
		if !got[i].Equal(want[i]) {
			t.Errorf("Next[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestNextMonthNames(t *testing.T) {
	s := mustParse(t, "0 0 1 JAN *")
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	got := s.Next(start, 1)
	want := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	if !got[0].Equal(want) {
		t.Errorf("Next = %v, want %v", got[0], want)
	}
}

func TestNextCountZero(t *testing.T) {
	s := mustParse(t, "* * * * *")
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	got := s.Next(start, 0)
	if got != nil {
		t.Errorf("Next(0) = %v, want nil", got)
	}
}

func TestNextSundayAsSeven(t *testing.T) {
	s := mustParse(t, "0 0 * * 7")
	// 2024-01-01 is a Monday; next Sunday is 2024-01-07.
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	got := s.Next(start, 1)
	want := time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC)
	if !got[0].Equal(want) {
		t.Errorf("Next = %v, want %v", got[0], want)
	}
}

func TestNextList(t *testing.T) {
	s := mustParse(t, "0 0 1,15 * *")
	start := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
	got := s.Next(start, 2)
	want := []time.Time{
		time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
	}
	for i := range want {
		if !got[i].Equal(want[i]) {
			t.Errorf("Next[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestNextRangeStep(t *testing.T) {
	s := mustParse(t, "0-30/10 * * * *")
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	got := s.Next(start, 4)
	want := []time.Time{
		time.Date(2024, 1, 1, 0, 10, 0, 0, time.UTC),
		time.Date(2024, 1, 1, 0, 20, 0, 0, time.UTC),
		time.Date(2024, 1, 1, 0, 30, 0, 0, time.UTC),
		time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC),
	}
	for i := range want {
		if !got[i].Equal(want[i]) {
			t.Errorf("Next[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestNextStartExclusive(t *testing.T) {
	s := mustParse(t, "* * * * *")
	start := time.Date(2024, 1, 1, 0, 0, 30, 0, time.UTC)
	got := s.Next(start, 1)
	want := time.Date(2024, 1, 1, 0, 1, 0, 0, time.UTC)
	if !got[0].Equal(want) {
		t.Errorf("Next = %v, want %v", got[0], want)
	}
}
