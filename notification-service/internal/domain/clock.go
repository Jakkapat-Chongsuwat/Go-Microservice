package domain

import "time"

// ClockInterface abstracts obtaining the current time.
type ClockInterface interface {
	Now() time.Time
}

// RealClock implements ClockInterface using the real time.
type RealClock struct{}

// Now returns the current time in UTC.
func (RealClock) Now() time.Time {
	return time.Now().UTC()
}

// Clock is the global clock used by the domain.
// In production it is set to RealClock, and tests can override it.
var Clock ClockInterface = RealClock{}
