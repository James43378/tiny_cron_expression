# tiny_cron_expression

A small Go library that parses standard five-field cron expressions and returns the upcoming run times from a given starting timestamp.

## Usage

```go
package main

import (
	"fmt"
	"time"

	tiny_cron_expression "github.com/example/tiny_cron_expression"
)

func main() {
	s, err := tiny_cron_expression.Parse("*/15 9-17 * * MON-FRI")
	if err != nil {
		panic(err)
	}
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, t := range s.Next(start, 5) {
		fmt.Println(t.Format(time.RFC3339))
	}
}
```

## Why this exists

Most cron parsers pull in a dependency or expose a large API. This one is deliberately tiny: it parses the five standard fields, supports the common range/list/step syntax plus month and day names, and gives you the next N times. The trade-off is that it does not support the non-standard six-field (seconds) or seven-field (year) variants, and it does not implement the `@daily`-style macros. If you need those, use a larger library.

## The awkward edge

When both day-of-month and day-of-week are restricted (neither is `*`), classic cron treats them as an OR: a time matches if either field matches. This library follows that behaviour. If only one is restricted, only that field is checked. This is a common source of surprise, so it is worth stating plainly.

## API

- `Parse(expr string) (*Schedule, error)` — parse a five-field cron expression.
- `(*Schedule) Next(start time.Time, count int) []time.Time` — return the next `count` run times strictly after `start`.
- `(*Schedule) Matches(t time.Time) bool` — report whether the schedule matches a given time.
