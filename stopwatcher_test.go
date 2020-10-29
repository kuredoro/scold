package cptest_test

import (
	"testing"
	"time"

	"github.com/kuredoro/cptest"
)

func TestSpyStopwatcher(t *testing.T) {

    t.Run("5 calls",
    func(t *testing.T) {
        const totalCalls = 5

        swatch := &cptest.SpyStopwatcher{
            TLAtCall: totalCalls,
        }

        for i := 0; i < totalCalls - 1; i++ {
            select {
            case <-time.After(1 * time.Millisecond):
                elapsedGot := swatch.Elapsed()
                elapsedWant := time.Duration(i + 1) * time.Second
                if elapsedGot != elapsedWant {
                    t.Errorf("got %v elapsed, want %v", elapsedGot, elapsedWant)
                }
            case <-swatch.TimeLimit():
                t.Errorf("got TL at call #%d, want at call #%d", i + 1, totalCalls)
            }
        }

        var firstTL, secondTL time.Duration
        select {
        case <-time.After(1 * time.Millisecond):
            t.Fatalf("go no TL at call #%d, want one", totalCalls)
        case firstTL = <-swatch.TimeLimit():
        }

        select {
        case <-time.After(1 * time.Millisecond):
            t.Fatalf("go no TL at call #%d, want one", totalCalls + 1)
        case secondTL = <-swatch.TimeLimit():
        }

        if firstTL != secondTL {
            t.Errorf("got first and seconds TLs that don't match (%v != %v), want matching ones", firstTL, secondTL)
        }

        if firstTL != time.Duration(totalCalls) * time.Second {
            t.Errorf("got TL equal to %v, want it equal to %v", firstTL, time.Duration(totalCalls))
        }
    })

    t.Run("TL at 0 should never TL",
    func(t *testing.T) {

        swatch := &cptest.SpyStopwatcher{}

        // There won't be more than 10 test cases in the tests, so I think it's
        // enough
        for i := 0; i < 10; i++ {
            select {
            case <-time.After(1 * time.Millisecond):
            case <-swatch.TimeLimit():
                t.Errorf("got TL at call #%d, want none", i + 1)
            }
        }
    })
}

func TestConfigurableStopwatcher(t *testing.T) {

    t.Run("timeout after specified time",
    func(t *testing.T) {

        TL := 7 * time.Millisecond
        timeStep := 2 * time.Millisecond
        steps := int(TL / timeStep)

        swatch := cptest.NewConfigurableStopwatcher(TL)

        // It seems that this is enough of error for the check of the TL times
        // at the end to work
        const eps = 400 * time.Microsecond

        for i := 0; i < steps; i++ {
            elapsedWant := time.Duration(i + 1) * timeStep

            select {
            case <-swatch.TimeLimit():
                t.Errorf("got timelimit at %v, want at %v", elapsedWant, TL)
            case <-time.After(timeStep):
                elapsedGot := swatch.Elapsed()

                // It seems that time.After is not precise enough. It's error
                // seems to be ~0.2ms maximum (on my computer at least).
                // Hence after consequent time.After in the case statement above
                // error starts to accumulate and 'skew' the elapsedWant and
                // elapsedGot by making the first one fall behind.
                // This is my weak attempt to mitigate this.
                skew := time.Duration(i + 1) * eps

                if elapsedGot - elapsedWant > skew {
                    t.Errorf("got elapsed %v, want %v", elapsedGot, elapsedWant)
                }
            }
        }

        time.Sleep(TL - time.Duration(steps) * timeStep)

        select {
        case realTL := <-swatch.TimeLimit():
            if realTL - TL > eps {
                t.Errorf("received TL with value %v, it deviates from expected %v too much (%v)", 
                    realTL, TL, realTL - TL)
            }
        case <-time.After(timeStep):
            t.Errorf("got no timelimit at %v, want one", TL)
        }
    })

    t.Run("don't timeout if TL less or equal 0",
    func(t *testing.T) {
        swatch := cptest.NewConfigurableStopwatcher(0)

        limit := 20 * time.Millisecond
        step := 1 * time.Millisecond
        for tm := time.Duration(0); tm <= limit; tm += step {
            select {
            case <-swatch.TimeLimit():
                t.Fatalf("got time limit at %v, want none", tm)
            case <-time.After(step):
            }
        }
    })
}
