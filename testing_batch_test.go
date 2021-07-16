package cptest_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/kuredoro/cptest"
	//"github.com/sanity-io/litter"
)

func ProcFuncMultiply(ctx context.Context, in io.Reader) (cptest.ProcessResult, error) {
	var a, b int
	fmt.Fscan(in, &a, &b)

	fmt.Printf("multiply %d * %d\n", a, b)

	return cptest.ProcessResult{
		ExitCode: 0,
		Stdout:   fmt.Sprintln(a * b),
		Stderr:   "",
	}, nil
}

func ProcFuncIntegerSequence(ctx context.Context, in io.Reader) (cptest.ProcessResult, error) {
	var n int
	fmt.Fscan(in, &n)

	buf := &bytes.Buffer{}
	for i := 1; i <= n; i++ {
		fmt.Fprint(buf, i, " ")
	}

	fmt.Fprintln(buf)

	return cptest.ProcessResult{
		ExitCode: 0,
		Stdout:   buf.String(),
		Stderr:   "",
	}, nil
}

func ProcFuncBogusFloatingPoint(ctx context.Context, in io.Reader) (cptest.ProcessResult, error) {
	var n int
	fmt.Fscan(in, &n)

	out := ""
	if n == 1 {
		out = "1.234567\n"
	} else if n == 2 {
		out = "2.345678\n"
	}

	return cptest.ProcessResult{
		ExitCode: 0,
		Stdout:   out,
		Stderr:   "",
	}, nil
}

func ProcFuncAnswer(ctx context.Context, in io.Reader) (cptest.ProcessResult, error) {
	return cptest.ProcessResult{
		ExitCode: 0,
		Stdout:   "42",
		Stderr:   "",
	}, nil
}

func TestNewTestingBatch(t *testing.T) {
	t.Run("no state altering configs", func(t *testing.T) {
		inputs := cptest.Inputs{
			Tests:  nil,
			Config: map[string]string{},
		}

		batch := cptest.NewTestingBatch(inputs, nil, nil, nil)

		if batch.Lx.Precision != cptest.DefaultPrecision {
			t.Errorf("got lexer precision %d, but want default value %d",
				batch.Lx.Precision, cptest.DefaultPrecision)
		}
	})

	t.Run("prec option", func(t *testing.T) {
		inputs := cptest.Inputs{
			Tests: nil,
			Config: map[string]string{
				"prec": "22",
			},
		}

		batch := cptest.NewTestingBatch(inputs, nil, nil, nil)

		if batch.Lx.Precision != 22 {
			t.Errorf("got lexer precision %d, but want 22", batch.Lx.Precision)
		}
	})
}

/*
// IDEA: Add support for presentation errors...
func TestTestingBatch(t *testing.T) {
	t.Run("all OK", func(t *testing.T) {
		inputs := cptest.Inputs{
			Tests: []cptest.Test{
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

		proc := &cptest.SpyProcesser{
			Proc: cptest.ProcesserFunc(ProcFuncMultiply),
		}

		swatch := &cptest.SpyStopwatcher{Clock: clockwork.NewFakeClock()}
		pool := cptest.NewSpyThreadPool(2)

		batch := cptest.NewTestingBatch(inputs, proc, swatch, pool)
		batch.Run()

		want := map[int]cptest.Verdict{
			1: cptest.OK,
			2: cptest.OK,
		}

		cptest.AssertVerdicts(t, batch.Verdicts, want)
		cptest.AssertCallCount(t, "proc.Run()", proc.CallCount(), 2)
		cptest.AssertThreadCount(t, pool, 2)
	})

	t.Run("outputs are compared lexeme-wise", func(t *testing.T) {
		inputs := cptest.Inputs{
			Tests: []cptest.Test{
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

		proc := &cptest.SpyProcesser{
			Proc: cptest.ProcesserFunc(ProcFuncIntegerSequence),
		}

		swatch := &cptest.SpyStopwatcher{Clock: clockwork.NewFakeClock()}
		pool := cptest.NewSpyThreadPool(2)

		batch := cptest.NewTestingBatch(inputs, proc, swatch, pool)
		batch.Run()

		want := map[int]cptest.Verdict{
			1: cptest.OK,
			2: cptest.OK,
		}

		cptest.AssertVerdicts(t, batch.Verdicts, want)
		cptest.AssertCallCount(t, "proc.Run()", proc.CallCount(), 2)
		cptest.AssertThreadCount(t, pool, 2)

		if len(batch.RichAnswers[1]) != 3 || len(batch.RichAnswers[2]) != 4 {
			t.Errorf("got wrong rich answers, %s", litter.Sdump(batch.RichAnswers))
		}

		if len(batch.RichOuts[1]) != 3 || len(batch.RichOuts[2]) != 4 {
			t.Errorf("got wrong rich outputs, %s", litter.Sdump(batch.RichOuts))
		}
	})

	t.Run("floating point values are compared correctly", func(t *testing.T) {
		inputs := cptest.Inputs{
			Tests: []cptest.Test{
				{
					Input:  "1\n",
					Output: "1.25\n",
				},
				{
					Input:  "2\n",
					Output: "2.5\n",
				},
			},
			Config: map[string]string{
				"prec": "1",
			},
		}

		proc := &cptest.SpyProcesser{
			Proc: cptest.ProcesserFunc(ProcFuncBogusFloatingPoint),
		}

		swatch := &cptest.SpyStopwatcher{Clock: clockwork.NewFakeClock()}
		pool := cptest.NewSpyThreadPool(2)

		batch := cptest.NewTestingBatch(inputs, proc, swatch, pool)
		batch.Run()

		want := map[int]cptest.Verdict{
			1: cptest.OK,
			2: cptest.WA,
		}

		cptest.AssertVerdicts(t, batch.Verdicts, want)
		cptest.AssertCallCount(t, "proc.Run()", proc.CallCount(), 2)
		cptest.AssertThreadCount(t, pool, 2)
	})

	t.Run("all WA",
		func(t *testing.T) {
			inputs := cptest.Inputs{
				Tests: []cptest.Test{
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

			proc := &cptest.SpyProcesser{
				Proc: cptest.ProcesserFunc(ProcFuncAnswer),
			}

			swatch := &cptest.SpyStopwatcher{Clock: clockwork.NewFakeClock()}
			pool := cptest.NewSpyThreadPool(2)

			batch := cptest.NewTestingBatch(inputs, proc, swatch, pool)
			batch.Run()

			want := map[int]cptest.Verdict{
				1: cptest.WA,
				2: cptest.WA,
			}

			cptest.AssertVerdicts(t, batch.Verdicts, want)
			cptest.AssertCallCount(t, "proc.Run()", proc.CallCount(), 2)
			cptest.AssertThreadCount(t, pool, 2)
		})

	t.Run("runtime error and internal error",
		func(t *testing.T) {
			inputs := cptest.Inputs{
				Tests: []cptest.Test{
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

			proc := &cptest.SpyProcesser{
				Proc: cptest.ProcesserFunc(
					func(ctx context.Context, r io.Reader) (cptest.ProcessResult, error) {
						var num int
						fmt.Fscan(r, &num)

						if num == 3 {
							return cptest.ProcessResult{
								ExitCode: 1,
								Stdout:   "",
								Stderr:   "segfault. (core dumped)",
							}, nil
						}

						if num == 5 {
							panic("brrrr")
						}

						return cptest.ProcessResult{
							ExitCode: 0,
							Stdout:   "1\n",
							Stderr:   "",
						}, nil
					}),
			}

			swatch := &cptest.SpyStopwatcher{Clock: clockwork.NewFakeClock()}
			pool := cptest.NewSpyThreadPool(3)

			batch := cptest.NewTestingBatch(inputs, proc, swatch, pool)
			batch.Run()

			want := map[int]cptest.Verdict{
				1: cptest.OK,
				2: cptest.WA,
				3: cptest.RE,
				4: cptest.WA,
				5: cptest.IE,
			}

			cptest.AssertVerdicts(t, batch.Verdicts, want)
			cptest.AssertCallCount(t, "proc.Run()", proc.CallCount(), 5)
			cptest.AssertThreadCount(t, pool, 3)

			if len(batch.RichAnswers[3]) == 0 || len(batch.RichAnswers[5]) == 0 {
				t.Errorf("got wrong rich answers, %s", litter.Sdump(batch.RichAnswers))
			}
		})

}
*/

func TestTestingBatch2(t *testing.T) {
	t.Run("single TL (proc doesn't run because it didn't have time to dispatch)",
		func(t *testing.T) {
			inputs := cptest.Inputs{
				Tests: []cptest.Test{
					{"\n", "bar\n"},
				},
			}

			clock := clockwork.NewFakeClock()

			killCount := 0

			proc := &cptest.SpyProcesser{
				Proc: cptest.ProcesserFunc(
					func(ctx context.Context, r io.Reader) (cptest.ProcessResult, error) {
						<-ctx.Done()
						killCount++

						return cptest.ProcessResult{
							ExitCode: 0,
							Stdout:   "",
							Stderr:   "",
						}, cptest.TLError
					}),
			}

			swatch := &cptest.SpyStopwatcher{
				Clock: clock,
				TL:    3 * time.Second,
			}
			pool := cptest.NewSpyThreadPool(1)

			batch := cptest.NewTestingBatch(inputs, proc, swatch, pool)

			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				batch.Run()
				wg.Done()
			}()

			clock.BlockUntil(1)
			clock.Advance(3 * time.Second)

			wg.Wait()

			testsWant := map[int]cptest.Verdict{
				1: cptest.TL,
			}

			timesWant := map[int]time.Duration{
				1: 3 * time.Second,
			}

			cptest.AssertVerdicts(t, batch.Verdicts, testsWant)
			cptest.AssertThreadCount(t, pool, 1)

			// Should be too fast for anyone to be killed.
			cptest.AssertCallCount(t, "proc.Run()", proc.CallCount(), 0)
			cptest.AssertCallCount(t, "process cancel", killCount, 0)
			cptest.AssertTimes(t, batch.Times, timesWant)
		})

	t.Run("single TL (proc runs)",
		func(t *testing.T) {
			inputs := cptest.Inputs{
				Tests: []cptest.Test{
					{"\n", "bar\n"},
				},
			}

			clock := clockwork.NewFakeClock()

			killCount := 0

			proc := &cptest.SpyProcesser{
				Proc: cptest.ProcesserFunc(
					func(ctx context.Context, r io.Reader) (cptest.ProcessResult, error) {
						select {
						case <-clock.After(5 * time.Second):
						case <-ctx.Done():
							killCount++
                            return cptest.ProcessResult{
                                ExitCode: 0,
                                Stdout:   "",
                                Stderr:   "",
                            }, cptest.TLError
						}

						return cptest.ProcessResult{
							ExitCode: 0,
							Stdout:   "",
							Stderr:   "",
						}, nil

					}),
			}

			swatch := &cptest.SpyStopwatcher{
				Clock: clock,
				TL:    3 * time.Second,
			}
			pool := cptest.NewSpyThreadPool(1)

			batch := cptest.NewTestingBatch(inputs, proc, swatch, pool)

			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				batch.Run()
				wg.Done()
			}()

			//clock.BlockUntil(1)
			clock.BlockUntil(2)
			clock.Advance(3 * time.Second)

			wg.Wait()

			testsWant := map[int]cptest.Verdict{
				1: cptest.TL,
			}

			timesWant := map[int]time.Duration{
				1: 3 * time.Second,
			}

			cptest.AssertVerdicts(t, batch.Verdicts, testsWant)
			cptest.AssertThreadCount(t, pool, 1)

			cptest.AssertCallCount(t, "proc.Run()", proc.CallCount(), 1)
			cptest.AssertCallCount(t, "process cancel", killCount, 1)
			cptest.AssertTimes(t, batch.Times, timesWant)
		})

	t.Run("test cases may be abandoned at TL",
		func(t *testing.T) {
			return
			inputs := cptest.Inputs{
				Tests: []cptest.Test{
					{"1\n", "1\n"},
					{"2\n", "2\n"},
					{"4\n", "4\n"},
					{"5\n", "5\n"},
				},
			}

			clock := clockwork.NewFakeClock()
			clock.Advance(1 * time.Second)
			fmt.Fprintf(os.Stderr, "!!!!!!!!!!!!!!!!!!!!!!!!!! Now: %v\n", clock.Now())

			advanceTicks := []time.Duration{time.Second, time.Second, 3 * time.Second, time.Second}
			doneHandler := func(b *cptest.TestingBatch, t cptest.Test, id int) {
				fmt.Printf("done id=%d, andvancing %v\n", id, advanceTicks[0])
				clock.Advance(advanceTicks[0])
				advanceTicks = advanceTicks[1:]
			}

			var mu sync.Mutex
			killCount := 0

			proc := &cptest.SpyProcesser{
				Proc: cptest.ProcesserFunc(
					func(ctx context.Context, r io.Reader) (cptest.ProcessResult, error) {
						var num int
						fmt.Fscan(r, &num)

						dur := time.Duration(num)
						select {
						case <-clock.After(dur * time.Second):
						case <-ctx.Done():
							mu.Lock()
							killCount++
							mu.Unlock()
						}

						return cptest.ProcessResult{
							ExitCode: 0,
							Stdout:   fmt.Sprintln(num),
							Stderr:   "",
						}, cptest.TLError
					}),
			}

			swatch := &cptest.SpyStopwatcher{
				Clock: clock,
				TL:    3 * time.Second,
			}
			pool := cptest.NewSpyThreadPool(4)

			batch := cptest.NewTestingBatch(inputs, proc, swatch, pool)
			batch.TestEndCallback = doneHandler

			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				batch.Run()
				wg.Done()
			}()

			//clock.BlockUntil(1)
			clock.BlockUntil(1)
			clock.Advance(20 * time.Second)
			clock.BlockUntil(1)
			clock.Advance(20 * time.Second)
			clock.BlockUntil(1)
			clock.Advance(20 * time.Second)
			clock.BlockUntil(1)
			clock.Advance(20 * time.Second)
			clock.BlockUntil(1)
			clock.Advance(20 * time.Second)
			clock.BlockUntil(1)
			clock.Advance(20 * time.Second)

			//wg.Wait()

			testsWant := map[int]cptest.Verdict{
				1: cptest.OK,
				2: cptest.OK,
				3: cptest.TL,
				4: cptest.TL,
			}

			timesWant := map[int]time.Duration{
				1: 1 * time.Second,
				2: 2 * time.Second,
				3: 3 * time.Second,
				4: 3 * time.Second,
			}

			cptest.AssertVerdicts(t, batch.Verdicts, testsWant)
			cptest.AssertThreadCount(t, pool, 4)

			cptest.AssertCallCount(t, "proc.Run()", proc.CallCount(), 4)
			cptest.AssertCallCount(t, "process cancel", killCount, 2)
			cptest.AssertTimes(t, batch.Times, timesWant)

			cptest.AssertEnrichedLexSequence(t, batch.RichAnswers[3], []cptest.RichText{{"3", []bool{false}}, {"\n", []bool{false}}})
			cptest.AssertEnrichedLexSequence(t, batch.RichAnswers[4], []cptest.RichText{{"4", []bool{false}}, {"\n", []bool{false}}})
		})
}
