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

// ConfigurableStopwatcher implements Stopwatcher but instead of real time substitutes
// index sequence numbers. If time limit equals zero, then the time limit will
// never fire.
type ConfigurableStopwatcher struct {
	TL    time.Duration
	Clock clockwork.Clock
}

func (s *ConfigurableStopwatcher) Now() time.Time {
	fmt.Printf("Return now: %v\n", s.Clock.Now())
	return s.Clock.Now()
}

// Elapsed will return the number of seconds that equals to the number of
// calls made to the TimeLimit method.
func (s *ConfigurableStopwatcher) Elapsed(since time.Time) time.Duration {
	if s.Clock.Now().After(since) {
		return s.Clock.Since(since)
	}

	return 0
}

// TimeLimit returns a channel that sends the TLAtCall number of seconds
// back at the TLAtCall-th call to the TimeLimit method.
// -----------------------
// Returns a channel that will never fire if configured with TL = 0 or
// if since is zero-initialized.
func (s *ConfigurableStopwatcher) TimeLimit(since time.Time) <-chan time.Time {
	fmt.Printf("TimeLimit(%v)\n", since)
	if s.TL == 0 || since == (time.Time{}) {
		ch := make(chan time.Time, 1)
		return ch
	}

	fmt.Printf("TimeLimit(%v) after %v (with now=%v)\n", since, since.Add(s.TL).Sub(s.Clock.Now()), s.Clock.Now())

	return s.Clock.After(since.Add(s.TL).Sub(s.Clock.Now()))
}
