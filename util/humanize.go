package util

import (
	"fmt"
	"time"
)

func HumanizeApproximateDuration(d time.Duration) string {
	const (
		day   = time.Hour * 24
		month = day * 30 // approximately :)
		year  = 365 * day
	)
	if d >= year {
		return fmt.Sprintf("%dy", d/year)
	}
	if d >= month {
		return fmt.Sprintf("%dm", d/month)
	}
	if d >= day {
		return fmt.Sprintf("%dd", d/day)
	}
	if d >= time.Hour {
		return fmt.Sprintf("%dh", d/time.Hour)
	}
	if d >= time.Minute {
		return fmt.Sprintf("%dm", d/time.Minute)
	}
	return fmt.Sprintf("%ds", d/time.Second)
}
