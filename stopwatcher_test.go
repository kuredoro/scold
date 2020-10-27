package cptest_test

import (
	"testing"
	"time"

	"github.com/kureduro/cptest"
)

func TestSpyStopwatcher(t *testing.T) {

    t.Run("5 calls",
    func(t *testing.T) {
        const totalCalls = 5

        watch := &cptest.SpyStopwatcher{
            TLAtCall: totalCalls,
        }

        for i := 0; i < totalCalls - 1; i++ {
            select {
            case <-time.After(1 * time.Millisecond):
            case <-watch.TimeLimit():
                t.Errorf("got TL at call #%d, want at call #%d", i + 1, totalCalls)
            }
        }

        var firstTL, secondTL time.Duration
        select {
        case <-time.After(1 * time.Millisecond):
            t.Fatalf("go no TL at call #%d, want one", totalCalls)
        case firstTL = <-watch.TimeLimit():
        }

        select {
        case <-time.After(1 * time.Millisecond):
            t.Fatalf("go no TL at call #%d, want one", totalCalls + 1)
        case secondTL = <-watch.TimeLimit():
        }

        if firstTL != secondTL {
            t.Errorf("got first and seconds TLs that don't match (%v != %v), want matching ones", firstTL, secondTL)
        }

        if firstTL != time.Duration(totalCalls) {
            t.Errorf("got TL equal to %v, want it equal to %v", firstTL, time.Duration(totalCalls))
        }
    })

    t.Run("TL at 0 should never TL",
    func(t *testing.T) {

        watch := &cptest.SpyStopwatcher{}

        // There won't be more than 10 test cases in the tests, so I think it's
        // enough
        for i := 0; i < 10; i++ {
            select {
            case <-time.After(1 * time.Millisecond):
            case <-watch.TimeLimit():
                t.Errorf("got TL at call #%d, want none", i + 1)
            }
        }
    })
}
