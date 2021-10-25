package scold_test

import (
	"sync"
	"testing"
	"time"

	"context"
	"github.com/jonboulle/clockwork"
	"github.com/kuredoro/scold"
)

func TestConfigurableStopwatcher(t *testing.T) {
	t.Run("timeout after specified time given some initial time",
		func(t *testing.T) {
			TL := 5 * time.Second
			clock := clockwork.NewFakeClock()

			since := clock.Now().Add(5 * time.Second)

			swatch := &scold.ConfigurableStopwatcher{
				TL:    TL,
				Clock: clock,
			}

			done := make(chan time.Time)

			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				t := <-swatch.TimeLimit(since)
				wg.Done()
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

			wg.Wait()

			TLwant := testStartTime.Add(10 * time.Second)

			if TLgot == (time.Time{}) {
				t.Fatalf("never received TL, want at %v\n", TLwant)
			}

			if TLgot != TLwant {
				t.Fatalf("received TL at %v, want at %v\n", TLgot, TLwant)
			}

		})

	t.Run("don't time out when TL is 0",
		func(t *testing.T) {
			TL := 0 * time.Second
			clock := clockwork.NewFakeClock()

			since := clock.Now().Add(1 * time.Second)

			swatch := &scold.ConfigurableStopwatcher{
				TL:    TL,
				Clock: clock,
			}

			done := make(chan time.Time)
			ctx, cancel := context.WithCancel(context.Background())

			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				select {
				case t := <-swatch.TimeLimit(since):
					done <- t
				case <-ctx.Done():
				}
				wg.Done()
			}()

			// Make sure that the goroutine executed
			time.Sleep(time.Millisecond)

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

			cancel()
			wg.Wait()

			if TLgot != (time.Time{}) {
				t.Fatalf("received a TL at %v, want to never receive it\n", TLgot)
			}
		})

	t.Run("don't time out when `since` is 0",
		func(t *testing.T) {
			TL := 5 * time.Second
			clock := clockwork.NewFakeClock()

			var since time.Time

			swatch := &scold.ConfigurableStopwatcher{
				TL:    TL,
				Clock: clock,
			}

			done := make(chan time.Time)
			ctx, cancel := context.WithCancel(context.Background())

			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				select {
				case t := <-swatch.TimeLimit(since):
					done <- t
				case <-ctx.Done():
				}
				wg.Done()
			}()

			// Make sure that the goroutine executed
			time.Sleep(time.Millisecond)

			var TLgot time.Time
			for i := 0; i < 20; i++ {
				select {
				case t := <-done:
					TLgot = t
				case <-time.After(time.Millisecond):
					clock.Advance(1 * time.Second)
				}

				if TLgot != (time.Time{}) {
					break
				}

			}

			cancel()
			wg.Wait()

			if TLgot != (time.Time{}) {
				t.Fatalf("received a TL at %v, want to never receive it\n", TLgot)
			}
		})
}
