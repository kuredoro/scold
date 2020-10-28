package cptest_test

import (
	"errors"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/kureduro/cptest"
)

func ProcFuncMultiply(in io.Reader, out io.Writer) error {
    var a, b int
    fmt.Fscan(in, &a, &b)

    fmt.Fprintln(out, a * b)

    return nil
}

func ProcFuncAnswer(in io.Reader, out io.Writer) error {
    fmt.Fprintln(out, 42)

    return nil
}

// IDEA: Add support for presentation errors...
func TestTestingBatch(t *testing.T) {
    t.Run("all OK", 
    func(t *testing.T) {
        inputs := cptest.Inputs{
            Tests: []cptest.Test{
                {
                    Input: "2 2",
                    Output: "4",
                },
                {
                    Input: "-2 -2",
                    Output: "4",
                },
            },
        }

        proc := &cptest.SpyProcesser{
            Proc: cptest.ProcesserFunc(ProcFuncMultiply),
        }

        swatch := &cptest.SpyStopwatcher{}

        batch := cptest.NewTestingBatch(inputs, proc, swatch)
        batch.ResultPrinter = cptest.BlankResultPrinter

        batch.Run()

        want := map[int]cptest.Verdict{
            1: cptest.OK,
            2: cptest.OK,
        }

        cptest.AssertVerdicts(t, batch.Stat, want)
        cptest.AssertCallCount(t, proc.CallCount, 2)
    })

    t.Run("all WA",
    func(t *testing.T) {
        inputs := cptest.Inputs{
            Tests: []cptest.Test{
                {
                    Input: "4\n1 2 3 4",
                    Output: "2 3 4 5",
                },
                {
                    Input: "2\n-2 -1",
                    Output: "-1 0",
                },
            },
        }

        proc := &cptest.SpyProcesser{
            Proc: cptest.ProcesserFunc(ProcFuncAnswer),
        }

        swatch := &cptest.SpyStopwatcher{}

        batch := cptest.NewTestingBatch(inputs, proc, swatch)
        batch.ResultPrinter = cptest.BlankResultPrinter

        batch.Run()

        want := map[int]cptest.Verdict{
            1: cptest.WA,
            2: cptest.WA,
        }

        cptest.AssertVerdicts(t, batch.Stat, want)
        cptest.AssertCallCount(t, proc.CallCount, 2)
    })

    t.Run("runtime error and internal error",
    func(t *testing.T) {
        inputs := cptest.Inputs{
            Tests: []cptest.Test{
                {
                    Input: "1",
                    Output: "1",
                },
                {
                    Input: "2",
                    Output: "2",
                },
                {
                    Input: "3",
                    Output: "3",
                },
                {
                    Input: "4",
                    Output: "4",
                },
                {
                    Input: "5",
                    Output: "5",
                },
            },
        }

        proc := &cptest.SpyProcesser{
            Proc: cptest.ProcesserFunc(
            func(r io.Reader, w io.Writer) error {
                var num int
                fmt.Fscan(r, &num)

                if num == 3 {
                    return errors.New("segfault. core dumped.")
                }

                if num == 5 {
                    panic("brrrr")
                }

                fmt.Fprintln(w, 1)

                return nil
            }),
        }

        swatch := &cptest.SpyStopwatcher{}

        batch := cptest.NewTestingBatch(inputs, proc, swatch)
        batch.ResultPrinter = cptest.BlankResultPrinter

        batch.Run()

        want := map[int]cptest.Verdict{
            1: cptest.OK,
            2: cptest.WA,
            3: cptest.RE,
            4: cptest.WA,
            5: cptest.IE,
        }

        cptest.AssertVerdicts(t, batch.Stat, want)
        cptest.AssertCallCount(t, proc.CallCount, 5)
    })

    t.Run("test cases may be abandoned at TL",
    func(t *testing.T) {
        inputs := cptest.Inputs{
            Tests: []cptest.Test{
                {"1", "1"},
                {"2", "2"},
                {"3", "3"},
                {"4", "4"},
            },
        }

        proc := &cptest.SpyProcesser{
            Proc: cptest.ProcesserFunc(
            func(r io.Reader, w io.Writer) error {
                var num int
                fmt.Fscan(r, &num)

                dur := time.Duration(num)
                time.Sleep(dur * time.Millisecond)

                fmt.Fprintln(w, num)
                return nil
            }),
        }

        swatch := &cptest.SpyStopwatcher{
            TLAtCall: 3,
        }

        batch := cptest.NewTestingBatch(inputs, proc, swatch)
        batch.ResultPrinter = cptest.BlankResultPrinter

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

        cptest.AssertVerdicts(t, batch.Stat, testsWant)
        cptest.AssertCallCount(t, proc.CallCount, 4)
        cptest.AssertTimes(t, batch.Times, timesWant)
    })
}
