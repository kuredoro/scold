package cptest_test

import (
	"sync"
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/kuredoro/cptest"
)

func TestConfigurableStopwatcher(t *testing.T) {

	t.Run("timeout after specified time given some initial time",
		func(t *testing.T) {
			TL := 5 * time.Second
            clock := clockwork.NewFakeClock()

            since := clock.Now().Add(5 * time.Second)

            swatch := &cptest.ConfigurableStopwatcher{
                TL: TL,
                Clock: clock,
            }

            done := make(chan time.Time)

            var wg sync.WaitGroup
            wg.Add(1)
            go func() {
                t := <-swatch.TimeLimit(since)
                done <- t
            }()

            clock.BlockUntil(1)

            testStartTime := clock.Now()

            var TLgot time.Time
            var elapsedWant time.Duration
			for i := 0; i < 20; i++ {
				select {
                case t := <-done:
                    TLgot = t
				case <-time.After(time.Millisecond):
					elapsedGot := swatch.Elapsed(since)

					if elapsedGot != elapsedWant {
					    t.Errorf("at %v, got elapsed %v, want %v", clock.Now(), elapsedGot, elapsedWant)
					}

                    clock.Advance(1 * time.Second)
                    if clock.Now().After(since) {
                        elapsedWant += time.Second
                    }
				}

                if TLgot != (time.Time{}) {
                    break
                }

			}

            TLwant := testStartTime.Add(10 * time.Second)

            if TLgot == (time.Time{}) {
                t.Fatalf("never received TL, want at %v\n", TLwant)
            }

            if TLgot != TLwant {
                t.Fatalf("received TL at %v, want at %v\n", TLgot, TLwant)
            }

		})

        /*
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
        */
}
