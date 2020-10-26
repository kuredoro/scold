package cptest_test

import (
	"fmt"
	"io"
    "errors"
	"testing"

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
            },
        }

        proc := &cptest.SpyProcesser{
            Proc: cptest.ProcesserFunc(ProcFuncMultiply),
        }

        want := map[int]cptest.Verdict{
            1: cptest.OK,
        }

        batch := cptest.NewTestingBatch(inputs, proc)
        batch.ResultPrinter = cptest.BlankResultPrinter

        batch.Run()

        cptest.AssertVerdicts(t, batch.Stat, want)
        cptest.AssertCallCount(t, proc.CallCount, 1)
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

        batch := cptest.NewTestingBatch(inputs, proc)
        batch.ResultPrinter = cptest.BlankResultPrinter

        batch.Run()

        want := map[int]cptest.Verdict{
            1: cptest.WA,
            2: cptest.WA,
        }

        cptest.AssertVerdicts(t, batch.Stat, want)
        cptest.AssertCallCount(t, proc.CallCount, 2)
    })

    t.Run("multiple with a runtime error",
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

                fmt.Fprintln(w, 1)

                return nil
            }),
        }

        batch := cptest.NewTestingBatch(inputs, proc)
        batch.ResultPrinter = cptest.BlankResultPrinter

        batch.Run()

        want := map[int]cptest.Verdict{
            1: cptest.OK,
            2: cptest.WA,
            3: cptest.RE,
            4: cptest.WA,
        }

        cptest.AssertVerdicts(t, batch.Stat, want)
        cptest.AssertCallCount(t, proc.CallCount, 4)
    })
}
