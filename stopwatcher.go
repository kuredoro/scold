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


type ConfigurableStopwatcher struct {
    tlChan <-chan time.Duration
}

func NewConfigurableStopwatcher(TL time.Duration) *ConfigurableStopwatcher {
    tlChan := make(chan time.Duration)

    go func() {
        if TL <= time.Duration(0) {
            return
        }

        start := time.Now()
        after := <-time.After(TL)
        realTL := after.Sub(start)
        tlChan<-realTL
    }()

    return &ConfigurableStopwatcher{
        tlChan: tlChan,
    }

}

func (s *ConfigurableStopwatcher) TimeLimit() <-chan time.Duration {
    return s.tlChan
}

func (s *ConfigurableStopwatcher) Elapsed() time.Duration {
    return 0
}
