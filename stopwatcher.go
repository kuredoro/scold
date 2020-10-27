package cptest

import "time"

type Stopwatcher interface {
    Elapsed() time.Duration
    TimeLimit() <-chan time.Duration
}


type SpyStopwatcher struct {
    TLAtCall int
    callCount int
}

func (s *SpyStopwatcher) Elapsed() time.Duration {

    return time.Duration(s.callCount)
}

func (s *SpyStopwatcher) TimeLimit() <-chan time.Duration {
    s.callCount++

    ch := make(chan time.Duration, 1)
    
    if s.TLAtCall != 0 && s.TLAtCall <= s.callCount {
        ch<-time.Duration(s.TLAtCall)
    }

    return ch
} 
