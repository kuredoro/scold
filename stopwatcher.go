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

// SpyStopwatcher implements Stopwatcher but instead of real time substitutes
// index sequence numbers. If time limit equals zero, then the time limit will
// never fire.
type SpyStopwatcher struct {
	TL  time.Duration
	Clock clockwork.Clock
}

func (s *SpyStopwatcher) Now() time.Time {
    fmt.Printf("Return now: %v\n", s.Clock.Now())
    return s.Clock.Now()
}

// Elapsed will return the number of seconds that equals to the number of
// calls made to the TimeLimit method.
func (s *SpyStopwatcher) Elapsed(since time.Time) time.Duration {
	return s.Clock.Since(since)
}

// TimeLimit returns a channel that sends the TLAtCall number of seconds
// back at the TLAtCall-th call to the TimeLimit method.
// -----------------------
// Returns a channel that will never fire if configured with TL = 0 or
// if since is zero-initialized.
func (s *SpyStopwatcher) TimeLimit(since time.Time) <-chan time.Time {
    fmt.Printf("TimeLimit(%v)\n", since)
    fmt.Printf("TimeLimit(%v) after %v\n", since, since.Add(s.TL).Sub(s.Clock.Now()))
    if s.TL == 0 || since == (time.Time{}) {
        ch := make(chan time.Time, 1)
        return ch
    }


	return s.Clock.After(since.Add(s.TL).Sub(s.Clock.Now()))
}

// ConfigurableStopwatcher is an implementation of the Stopwatcher that
// uses real time.
type ConfigurableStopwatcher struct {
	tlChan    <-chan time.Duration
	startTime time.Time
}

// NewConfigurableStopwatcher will return an initialized
// ConfigurableStopwatcher with the desired time limit. If time limit
// specified is zero or negative, the time limit will never fire.
func NewConfigurableStopwatcher(TL time.Duration) *ConfigurableStopwatcher {
	tlChan := make(chan time.Duration)

	go func() {
		if TL <= time.Duration(0) {
			return
		}

		start := time.Now()
		after := <-time.After(TL)
		realTL := after.Sub(start)
		tlChan <- realTL
	}()

	return &ConfigurableStopwatcher{
		tlChan:    tlChan,
		startTime: time.Now(),
	}

}

// Elapsed returns the true number of seconds since the initialization of
// the ConfigurableStopwatcher.
func (s *ConfigurableStopwatcher) Elapsed() time.Duration {
	return time.Since(s.startTime)
}

// TimeLimit returns a channel that will send back the number of seconds
// passed since beginning until the time limit was fired. The returned
// value may not equal to the time limit ConfigurableStopwatcher was
// initialized with.
func (s *ConfigurableStopwatcher) TimeLimit() <-chan time.Duration {
	return s.tlChan
}
