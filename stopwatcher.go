package cptest

import "time"

// Stopwatcher abstracts away the concept of the stopwatch.
// At any time, one can look up the elapsed time. Additionally, one
// can be notified when the time is up.
type Stopwatcher interface {
	Elapsed() time.Duration
	TimeLimit() <-chan time.Duration
}

// SpyStopwatcher implements Stopwatcher but instead of real time substitutes
// index sequence numbers. If time limit equals zero, then the time limit will
// never fire.
type SpyStopwatcher struct {
	TLAtCall  int
	callCount int
}

// Elapsed will return the number of seconds that equals to the number of
// calls made to the TimeLimit method.
func (s *SpyStopwatcher) Elapsed() time.Duration {

	return time.Duration(s.callCount) * time.Second
}

// TimeLimit returns a channel that sends the TLAtCall number of seconds
// back at the TLAtCall-th call to the TimeLimit method.
func (s *SpyStopwatcher) TimeLimit() <-chan time.Duration {
	s.callCount++

	ch := make(chan time.Duration, 1)

	if s.TLAtCall != 0 && s.TLAtCall <= s.callCount {
		ch <- time.Duration(s.TLAtCall) * time.Second
	}

	return ch
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
