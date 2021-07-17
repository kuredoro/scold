package cptest

import (
	"fmt"
	"time"

	"github.com/jonboulle/clockwork"
)

// Stopwatcher abstracts away the concept of the stopwatch.
// At any time, one can look up the elapsed time. Additionally, one
// can be notified when the time is up.
type Stopwatcher interface {
	Now() time.Time
	Elapsed(since time.Time) time.Duration
	TimeLimit(since time.Time) <-chan time.Time
}

// ConfigurableStopwatcher implements Stopwatcher and allows its user to
// fully customize it with no make function required.
type ConfigurableStopwatcher struct {
	TL    time.Duration
	Clock clockwork.Clock
}

// Now will returns time point since internal's clock epoch.
func (s *ConfigurableStopwatcher) Now() time.Time {
	fmt.Printf("Return now: %v\n", s.Clock.Now())
	return s.Clock.Now()
}

// Elapsed will return the duration since the `since` time, or zero if `since`
// is in the future.
func (s *ConfigurableStopwatcher) Elapsed(since time.Time) time.Duration {
	if s.Clock.Now().After(since) {
		return s.Clock.Since(since)
	}

	return 0
}

// TimeLimit for a given `since` time point returns a channel that will return
// at `since` + TL time point the very same time point. I.e., it works just
// just like time.After, but you can specify the *time* after which you want to
// be notified.
func (s *ConfigurableStopwatcher) TimeLimit(since time.Time) <-chan time.Time {
	fmt.Printf("TimeLimit(%v)\n", since)
	if s.TL == 0 || since == (time.Time{}) {
		ch := make(chan time.Time, 1)
		return ch
	}

	fmt.Printf("TimeLimit(%v) after %v (with now=%v)\n", since, since.Add(s.TL).Sub(s.Clock.Now()), s.Clock.Now())

	return s.Clock.After(since.Add(s.TL).Sub(s.Clock.Now()))
}
