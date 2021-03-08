package cptest_test

import (
	"bytes"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/kuredoro/cptest"
	"github.com/sanity-io/litter"
)

func ProcFuncMultiply(in io.Reader) (cptest.ProcessResult, error) {
	var a, b int
	fmt.Fscan(in, &a, &b)

	return cptest.ProcessResult{
		ExitCode: 0,
		Stdout:   fmt.Sprintln(a * b),
		Stderr:   "",
	}, nil
}

func ProcFuncIntegerSequence(in io.Reader) (cptest.ProcessResult, error) {
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

func ProcFuncBogusFloatingPoint(in io.Reader) (cptest.ProcessResult, error) {
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

func ProcFuncAnswer(in io.Reader) (cptest.ProcessResult, error) {
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

		batch := cptest.NewTestingBatch(inputs, nil, nil)

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

		batch := cptest.NewTestingBatch(inputs, nil, nil)

		if batch.Lx.Precision != 22 {
			t.Errorf("got lexer precision %d, but want 22", batch.Lx.Precision)
		}
	})
}

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

		swatch := &cptest.SpyStopwatcher{}

		batch := cptest.NewTestingBatch(inputs, proc, swatch)

		batch.Run()

		want := map[int]cptest.Verdict{
			1: cptest.OK,
			2: cptest.OK,
		}

		cptest.AssertVerdicts(t, batch.Verdicts, want)
		cptest.AssertCallCount(t, proc.CallCount(), 2)
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

		swatch := &cptest.SpyStopwatcher{}

		batch := cptest.NewTestingBatch(inputs, proc, swatch)

		batch.Run()

		want := map[int]cptest.Verdict{
			1: cptest.OK,
			2: cptest.OK,
		}

		cptest.AssertVerdicts(t, batch.Verdicts, want)
		cptest.AssertCallCount(t, proc.CallCount(), 2)

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

		swatch := &cptest.SpyStopwatcher{}

		batch := cptest.NewTestingBatch(inputs, proc, swatch)

		batch.Run()

		want := map[int]cptest.Verdict{
			1: cptest.OK,
			2: cptest.WA,
		}

		cptest.AssertVerdicts(t, batch.Verdicts, want)
		cptest.AssertCallCount(t, proc.CallCount(), 2)
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

			swatch := &cptest.SpyStopwatcher{}

			batch := cptest.NewTestingBatch(inputs, proc, swatch)

			batch.Run()

			want := map[int]cptest.Verdict{
				1: cptest.WA,
				2: cptest.WA,
			}

			cptest.AssertVerdicts(t, batch.Verdicts, want)
			cptest.AssertCallCount(t, proc.CallCount(), 2)
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
					func(r io.Reader) (cptest.ProcessResult, error) {
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

			swatch := &cptest.SpyStopwatcher{}

			batch := cptest.NewTestingBatch(inputs, proc, swatch)

			batch.Run()

			want := map[int]cptest.Verdict{
				1: cptest.OK,
				2: cptest.WA,
				3: cptest.RE,
				4: cptest.WA,
				5: cptest.IE,
			}

			cptest.AssertVerdicts(t, batch.Verdicts, want)
			cptest.AssertCallCount(t, proc.CallCount(), 5)

			if len(batch.RichAnswers[3]) == 0 || len(batch.RichAnswers[5]) == 0 {
				t.Errorf("got wrong rich answers, %s", litter.Sdump(batch.RichAnswers))
			}
		})

	t.Run("test cases may be abandoned at TL",
		func(t *testing.T) {
			inputs := cptest.Inputs{
				Tests: []cptest.Test{
					{"1\n", "1\n"},
					{"2\n", "2\n"},
					{"3\n", "3\n"},
					{"4\n", "4\n"},
				},
			}

			proc := &cptest.SpyProcesser{
				Proc: cptest.ProcesserFunc(
					func(r io.Reader) (cptest.ProcessResult, error) {
						var num int
						fmt.Fscan(r, &num)

						dur := time.Duration(num)
						time.Sleep(5 * dur * time.Millisecond)

						return cptest.ProcessResult{
							ExitCode: 0,
							Stdout:   fmt.Sprintln(num),
							Stderr:   "",
						}, nil
					}),
			}

			swatch := &cptest.SpyStopwatcher{
				TLAtCall: 3,
			}

			batch := cptest.NewTestingBatch(inputs, proc, swatch)

			batch.Run()

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
			cptest.AssertCallCount(t, proc.CallCount(), 4)
			cptest.AssertTimes(t, batch.Times, timesWant)
		})
}
