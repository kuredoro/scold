package scold_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/kuredoro/scold"
	"github.com/sanity-io/litter"
)

func ProcFuncMultiply(ctx context.Context, in io.Reader) (scold.ExecutionResult, error) {
	var a, b int
	fmt.Fscan(in, &a, &b)

	return scold.ExecutionResult{
		ExitCode: 0,
		Stdout:   fmt.Sprintln(a * b),
		Stderr:   "",
	}, nil
}

func ProcFuncIntegerSequence(ctx context.Context, in io.Reader) (scold.ExecutionResult, error) {
	var n int
	fmt.Fscan(in, &n)

	buf := &bytes.Buffer{}
	for i := 1; i <= n; i++ {
		fmt.Fprint(buf, i, " ")
	}

	fmt.Fprintln(buf)

	return scold.ExecutionResult{
		ExitCode: 0,
		Stdout:   buf.String(),
		Stderr:   "",
	}, nil
}

func ProcFuncBogusFloatingPoint(ctx context.Context, in io.Reader) (scold.ExecutionResult, error) {
	var n int
	fmt.Fscan(in, &n)

	out := ""
	if n == 1 {
		out = "1.234567\n"
	} else if n == 2 {
		out = "2.345678\n"
	}

	return scold.ExecutionResult{
		ExitCode: 0,
		Stdout:   out,
		Stderr:   "",
	}, nil
}

func ProcFuncAnswer(ctx context.Context, in io.Reader) (scold.ExecutionResult, error) {
	return scold.ExecutionResult{
		ExitCode: 0,
		Stdout:   "42",
		Stderr:   "",
	}, nil
}

func TestNewTestingBatch(t *testing.T) {
	t.Run("no state altering configs", func(t *testing.T) {
		inputs := scold.Inputs{
			Tests:  nil,
			Config: scold.InputsConfig{},
		}

		batch := scold.NewTestingBatch(inputs, nil, nil, nil)

		if batch.Lx.Precision != 0 {
			t.Errorf("got lexer precision %d, but want zero", batch.Lx.Precision)
		}
	})

	t.Run("prec option", func(t *testing.T) {
		inputs := scold.Inputs{
			Tests: nil,
			Config: scold.InputsConfig{
				Prec: 22,
			},
		}

		batch := scold.NewTestingBatch(inputs, nil, nil, nil)

		if batch.Lx.Precision != 22 {
			t.Errorf("got lexer precision %d, but want 22", batch.Lx.Precision)
		}
	})
}

// IDEA: Add support for presentation errors...
func TestTestingBatch(t *testing.T) {
	t.Run("all OK", func(t *testing.T) {
		inputs := scold.Inputs{
			Tests: []scold.Test{
				{
					Input:  "2 2\n",
					Output: "4\n",
				},
				{
					Input:  "-2 -2\n",
					Output: "4\n",
				},
			},
		}

		proc := &scold.SpyProcesser{
			Proc: scold.ProcesserFunc(ProcFuncMultiply),
		}

		swatch := &scold.ConfigurableStopwatcher{Clock: clockwork.NewFakeClock()}
		pool := scold.NewSpyThreadPool(2)

		batch := scold.NewTestingBatch(inputs, proc, swatch, pool)
		batch.Run()

		want := map[int]scold.Verdict{
			1: scold.OK,
			2: scold.OK,
		}

		scold.AssertResultIDInvariant(t, batch)
		scold.AssertVerdicts(t, batch.Results, want)
		scold.AssertCallCount(t, "proc.Run()", proc.CallCount(), 2)
		scold.AssertThreadCount(t, pool, 2)
	})

	t.Run("outputs are compared lexeme-wise", func(t *testing.T) {
		inputs := scold.Inputs{
			Tests: []scold.Test{
				{
					Input:  "2\n",
					Output: "1  2\n",
				},
				{
					Input:  "3\n",
					Output: "1  2  3\n",
				},
			},
		}

		proc := &scold.SpyProcesser{
			Proc: scold.ProcesserFunc(ProcFuncIntegerSequence),
		}

		swatch := &scold.ConfigurableStopwatcher{Clock: clockwork.NewFakeClock()}
		pool := scold.NewSpyThreadPool(2)

		batch := scold.NewTestingBatch(inputs, proc, swatch, pool)
		batch.Run()

		want := map[int]scold.Verdict{
			1: scold.OK,
			2: scold.OK,
		}

		scold.AssertResultIDInvariant(t, batch)
		scold.AssertVerdicts(t, batch.Results, want)
		scold.AssertCallCount(t, "proc.Run()", proc.CallCount(), 2)
		scold.AssertThreadCount(t, pool, 2)

		if len(batch.Results[1].RichAnswer) != 3 || len(batch.Results[2].RichAnswer) != 4 {
			t.Errorf("got wrong rich answers, %s", litter.Sdump(batch.Results))
		}

		if len(batch.Results[1].RichOut) != 3 || len(batch.Results[2].RichOut) != 4 {
			t.Errorf("got wrong rich outputs, %s", litter.Sdump(batch.Results))
		}
	})

	t.Run("floating point values are compared correctly", func(t *testing.T) {
		inputs := scold.Inputs{
			Tests: []scold.Test{
				{
					Input:  "1\n",
					Output: "1.25\n",
				},
				{
					Input:  "2\n",
					Output: "2.5\n",
				},
			},
			Config: scold.InputsConfig{
				Prec: 1,
			},
		}

		proc := &scold.SpyProcesser{
			Proc: scold.ProcesserFunc(ProcFuncBogusFloatingPoint),
		}

		swatch := &scold.ConfigurableStopwatcher{Clock: clockwork.NewFakeClock()}
		pool := scold.NewSpyThreadPool(2)

		batch := scold.NewTestingBatch(inputs, proc, swatch, pool)
		batch.Run()

		want := map[int]scold.Verdict{
			1: scold.OK,
			2: scold.WA,
		}

		scold.AssertResultIDInvariant(t, batch)
		scold.AssertVerdicts(t, batch.Results, want)
		scold.AssertCallCount(t, "proc.Run()", proc.CallCount(), 2)
		scold.AssertThreadCount(t, pool, 2)
	})

	t.Run("all WA",
		func(t *testing.T) {
			inputs := scold.Inputs{
				Tests: []scold.Test{
					{
						Input:  "4\n1 2 3 4\n",
						Output: "2 3 4 5\n",
					},
					{
						Input:  "2\n-2 -1\n",
						Output: "-1 0\n",
					},
				},
			}

			proc := &scold.SpyProcesser{
				Proc: scold.ProcesserFunc(ProcFuncAnswer),
			}

			swatch := &scold.ConfigurableStopwatcher{Clock: clockwork.NewFakeClock()}
			pool := scold.NewSpyThreadPool(2)

			batch := scold.NewTestingBatch(inputs, proc, swatch, pool)
			batch.Run()

			want := map[int]scold.Verdict{
				1: scold.WA,
				2: scold.WA,
			}

			scold.AssertResultIDInvariant(t, batch)
			scold.AssertVerdicts(t, batch.Results, want)
			scold.AssertCallCount(t, "proc.Run()", proc.CallCount(), 2)
			scold.AssertThreadCount(t, pool, 2)
		})

	t.Run("runtime error and internal error",
		func(t *testing.T) {
			inputs := scold.Inputs{
				Tests: []scold.Test{
					{
						Input:  "1\n",
						Output: "1\n",
					},
					{
						Input:  "2\n",
						Output: "2\n",
					},
					{
						Input:  "3\n",
						Output: "3\n",
					},
					{
						Input:  "4\n",
						Output: "4\n",
					},
					{
						Input:  "5\n",
						Output: "5\n",
					},
				},
			}

			proc := &scold.SpyProcesser{
				Proc: scold.ProcesserFunc(
					func(ctx context.Context, r io.Reader) (scold.ExecutionResult, error) {
						var num int
						fmt.Fscan(r, &num)

						if num == 3 {
							return scold.ExecutionResult{
								ExitCode: 1,
								Stdout:   "",
								Stderr:   "segfault. (core dumped)",
							}, nil
						}

						if num == 5 {
							panic("brrrr")
						}

						return scold.ExecutionResult{
							ExitCode: 0,
							Stdout:   "1\n",
							Stderr:   "",
						}, nil
					}),
			}

			swatch := &scold.ConfigurableStopwatcher{Clock: clockwork.NewFakeClock()}
			pool := scold.NewSpyThreadPool(3)

			batch := scold.NewTestingBatch(inputs, proc, swatch, pool)
			batch.Run()

			want := map[int]scold.Verdict{
				1: scold.OK,
				2: scold.WA,
				3: scold.RE,
				4: scold.WA,
				5: scold.IE,
			}

			scold.AssertResultIDInvariant(t, batch)
			scold.AssertVerdicts(t, batch.Results, want)
			scold.AssertCallCount(t, "proc.Run()", proc.CallCount(), 5)
			scold.AssertThreadCount(t, pool, 3)

			if len(batch.Results[3].RichAnswer) == 0 || len(batch.Results[5].RichAnswer) == 0 {
				t.Errorf("got wrong rich answers, %s", litter.Sdump(batch.Results))
			}
		})

	t.Run("single TL (proc doesn't run because it didn't have time to dispatch)",
		func(t *testing.T) {
			inputs := scold.Inputs{
				Tests: []scold.Test{
					{"\n", "bar\n"},
				},
			}

			clock := clockwork.NewFakeClock()

			killCount := 0

			proc := &scold.SpyProcesser{
				Proc: scold.ProcesserFunc(
					func(ctx context.Context, r io.Reader) (scold.ExecutionResult, error) {
						<-ctx.Done()
						killCount++

						return scold.ExecutionResult{
							ExitCode: 0,
							Stdout:   "",
							Stderr:   "",
						}, nil
					}),
			}

			swatch := &scold.ConfigurableStopwatcher{
				Clock: clock,
				TL:    3 * time.Second,
			}
			pool := scold.NewSpyThreadPool(1)

			batch := scold.NewTestingBatch(inputs, proc, swatch, pool)

			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				batch.Run()
				wg.Done()
			}()

			clock.BlockUntil(1)
			clock.Advance(3 * time.Second)

			wg.Wait()

			testsWant := map[int]scold.Verdict{
				1: scold.TL,
			}

			timesWant := map[int]time.Duration{
				1: 3 * time.Second,
			}

			scold.AssertResultIDInvariant(t, batch)
			scold.AssertVerdicts(t, batch.Results, testsWant)
			scold.AssertThreadCount(t, pool, 1)

			// Should be too fast for anyone to be killed.
			scold.AssertCallCount(t, "proc.Run()", proc.CallCount(), 0)
			scold.AssertCallCount(t, "process cancel", killCount, 0)
			scold.AssertTimes(t, batch.Results, timesWant)
		})

	t.Run("single TL (proc runs)",
		func(t *testing.T) {
			inputs := scold.Inputs{
				Tests: []scold.Test{
					{"\n", "bar\n"},
				},
			}

			clock := clockwork.NewFakeClock()

			killCount := 0

			proc := &scold.SpyProcesser{
				Proc: scold.ProcesserFunc(
					func(ctx context.Context, r io.Reader) (scold.ExecutionResult, error) {
						select {
						case <-clock.After(5 * time.Second):
						case <-ctx.Done():
							killCount++
							return scold.ExecutionResult{
								ExitCode: 0,
								Stdout:   "",
								Stderr:   "",
							}, scold.TLError
						}

						return scold.ExecutionResult{
							ExitCode: 0,
							Stdout:   "",
							Stderr:   "",
						}, nil

					}),
			}

			swatch := &scold.ConfigurableStopwatcher{
				Clock: clock,
				TL:    3 * time.Second,
			}
			pool := scold.NewSpyThreadPool(1)

			batch := scold.NewTestingBatch(inputs, proc, swatch, pool)

			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				batch.Run()
				wg.Done()
			}()

			clock.BlockUntil(2)
			clock.Advance(3 * time.Second)

			wg.Wait()

			testsWant := map[int]scold.Verdict{
				1: scold.TL,
			}

			timesWant := map[int]time.Duration{
				1: 3 * time.Second,
			}

			scold.AssertResultIDInvariant(t, batch)
			scold.AssertVerdicts(t, batch.Results, testsWant)
			scold.AssertThreadCount(t, pool, 1)

			scold.AssertCallCount(t, "proc.Run()", proc.CallCount(), 1)
			scold.AssertCallCount(t, "process cancel", killCount, 1)
			scold.AssertTimes(t, batch.Results, timesWant)
		})

	t.Run("two TL, thread count 1",
		func(t *testing.T) {
			inputs := scold.Inputs{
				Tests: []scold.Test{
					{"\n", "bar\n"},
					{"\n", "bar\n"},
				},
			}

			clock := clockwork.NewFakeClock()

			killCount := 0

			proc := &scold.SpyProcesser{
				Proc: scold.ProcesserFunc(
					func(ctx context.Context, r io.Reader) (scold.ExecutionResult, error) {
						select {
						case <-clock.After(5 * time.Second):
						case <-ctx.Done():
							killCount++
							return scold.ExecutionResult{
								ExitCode: 0,
								Stdout:   "",
								Stderr:   "",
							}, scold.TLError
						}

						return scold.ExecutionResult{
							ExitCode: 0,
							Stdout:   "",
							Stderr:   "",
						}, nil

					}),
			}

			swatch := &scold.ConfigurableStopwatcher{
				Clock: clock,
				TL:    3 * time.Second,
			}
			pool := scold.NewSpyThreadPool(1)

			batch := scold.NewTestingBatch(inputs, proc, swatch, pool)

			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				batch.Run()
				wg.Done()
			}()

			clock.BlockUntil(2)
			clock.Advance(3 * time.Second)
			clock.BlockUntil(3)
			clock.Advance(3 * time.Second)

			wg.Wait()

			testsWant := map[int]scold.Verdict{
				1: scold.TL,
				2: scold.TL,
			}

			timesWant := map[int]time.Duration{
				1: 3 * time.Second,
				2: 3 * time.Second,
			}

			scold.AssertResultIDInvariant(t, batch)
			scold.AssertVerdicts(t, batch.Results, testsWant)
			scold.AssertThreadCount(t, pool, 1)

			scold.AssertCallCount(t, "proc.Run()", proc.CallCount(), 2)
			scold.AssertCallCount(t, "process cancel", killCount, 2)
			scold.AssertTimes(t, batch.Results, timesWant)
		})

	t.Run("two TL, thread count 2",
		func(t *testing.T) {
			inputs := scold.Inputs{
				Tests: []scold.Test{
					{"\n", "bar\n"},
					{"\n", "bar\n"},
				},
			}

			clock := clockwork.NewFakeClock()

			var mu sync.Mutex
			killCount := 0

			proc := &scold.SpyProcesser{
				Proc: scold.ProcesserFunc(
					func(ctx context.Context, r io.Reader) (scold.ExecutionResult, error) {
						select {
						case <-clock.After(5 * time.Second):
						case <-ctx.Done():
							mu.Lock()
							killCount++
							mu.Unlock()
							return scold.ExecutionResult{
								ExitCode: 0,
								Stdout:   "",
								Stderr:   "",
							}, scold.TLError
						}

						return scold.ExecutionResult{
							ExitCode: 0,
							Stdout:   "",
							Stderr:   "",
						}, nil

					}),
			}

			swatch := &scold.ConfigurableStopwatcher{
				Clock: clock,
				TL:    3 * time.Second,
			}
			pool := scold.NewSpyThreadPool(2)

			batch := scold.NewTestingBatch(inputs, proc, swatch, pool)

			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				batch.Run()
				wg.Done()
			}()

			clock.BlockUntil(3)
			clock.Advance(3 * time.Second)

			wg.Wait()

			testsWant := map[int]scold.Verdict{
				1: scold.TL,
				2: scold.TL,
			}

			timesWant := map[int]time.Duration{
				1: 3 * time.Second,
				2: 3 * time.Second,
			}

			scold.AssertResultIDInvariant(t, batch)
			scold.AssertVerdicts(t, batch.Results, testsWant)
			scold.AssertThreadCount(t, pool, 2)

			scold.AssertCallCount(t, "proc.Run()", proc.CallCount(), 2)
			scold.AssertCallCount(t, "process cancel", killCount, 2)
			scold.AssertTimes(t, batch.Results, timesWant)
		})

	t.Run("two TL two OK, thread count 2",
		func(t *testing.T) {
			inputs := scold.Inputs{
				Tests: []scold.Test{
					{"2\n", "2\n"},
					{"5\n", "5\n"},
					{"2\n", "2\n"},
					{"5\n", "5\n"},
				},
			}

			clock := clockwork.NewFakeClock()

			var mu sync.Mutex
			killCount := 0

			proc := &scold.SpyProcesser{
				Proc: scold.ProcesserFunc(
					func(ctx context.Context, r io.Reader) (scold.ExecutionResult, error) {
						line, _ := ioutil.ReadAll(r)

						var num int
						num, err := strconv.Atoi(string(line[:len(line)-1]))

						if err != nil {
							panic(err)
						}

						select {
						case <-clock.After(time.Duration(num) * time.Second):
						case <-ctx.Done():
							mu.Lock()
							killCount++
							mu.Unlock()
							return scold.ExecutionResult{
								ExitCode: 0,
								Stdout:   "",
								Stderr:   "",
							}, nil
						}

						return scold.ExecutionResult{
							ExitCode: 0,
							Stdout:   string(line),
							Stderr:   "",
						}, nil

					}),
			}

			swatch := &scold.ConfigurableStopwatcher{
				Clock: clock,
				TL:    3 * time.Second,
			}
			pool := scold.NewSpyThreadPool(2)

			doneCh := make(chan struct{}, 1)
			done := func(*scold.Test, *scold.TestResult) {
				doneCh <- (struct{}{})
			}

			batch := scold.NewTestingBatch(inputs, proc, swatch, pool)
			batch.Listener = scold.TestFinishedCallback(done)

			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				batch.Run()
				wg.Done()
			}()

			// Time:   12345
			// test 1: --
			// test 2: ---
			// test 3:   --
			// test 4:    ---
			clock.BlockUntil(3)
			clock.Advance(2 * time.Second)

			advances := []time.Duration{time.Second, time.Second, 2 * time.Second}
			blocks := []int{4, 5, 5}
			for i := range advances {
				<-doneCh
				clock.BlockUntil(blocks[i])

				clock.Advance(advances[i])
				if i == 2 {
					time.Sleep(200 * time.Millisecond)
					break
				}
			}

			wg.Wait()

			testsWant := map[int]scold.Verdict{
				1: scold.OK,
				2: scold.TL,
				3: scold.OK,
				4: scold.TL,
			}

			timesWant := map[int]time.Duration{
				1: 2 * time.Second,
				2: 3 * time.Second,
				3: 2 * time.Second,
				4: 3 * time.Second,
			}

			scold.AssertResultIDInvariant(t, batch)
			scold.AssertVerdicts(t, batch.Results, testsWant)
			scold.AssertThreadCount(t, pool, 2)

			scold.AssertCallCount(t, "proc.Run()", proc.CallCount(), 4)
			scold.AssertCallCount(t, "process cancel", killCount, 2)
			scold.AssertTimes(t, batch.Results, timesWant)
		})
}
