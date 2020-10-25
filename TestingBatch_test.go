package cptest_test

import (
	"fmt"
	"io"
	"testing"

	"github.com/kureduro/cptest"
)

func ProcFuncMultiply(in io.Reader, out io.Writer) {
    var a, b int
    fmt.Fscan(in, &a, &b)

    fmt.Fprintln(out, a * b)
}

func ProcFuncAnswer(in io.Reader, out io.Writer) {
    fmt.Fprintln(out, 42)
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

        proc := cptest.NewProcessFunc(ProcFuncMultiply)

        want := map[int]cptest.Verdict{
            1: cptest.OK,
        }

        batch := cptest.NewTestingBatch(inputs, proc)

        batch.Run()

        cptest.AssertVerdicts(t, batch.Stat, want)
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

        proc := cptest.NewProcessFunc(ProcFuncAnswer)

        batch := cptest.NewTestingBatch(inputs, proc)

        batch.Run()

        want := map[int]cptest.Verdict{
            1: cptest.WA,
            2: cptest.WA,
        }

        cptest.AssertVerdicts(t, batch.Stat, want)
    })
}
