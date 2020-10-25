package cptest_test

import (
	"fmt"
	"io"
	"testing"

	"github.com/kureduro/cptest"
)

func TestTestingBatch(t *testing.T) {
    inputs := cptest.Inputs{
        Tests: []cptest.Test{
            {
                Input: "2 2",
                Output: "4",
            },
        },
    }

    proc := cptest.NewProcessFunc(
    func(in io.Reader, out io.Writer) {
        var a, b int
        fmt.Fscan(in, &a, &b)

        fmt.Fprintln(out, a * b)
    })

    want := map[int]cptest.Verdict{
        1: cptest.OK,
    }

    batch := cptest.NewTestingBatch(inputs, proc)

    batch.Run()

    if len(batch.Stat) != len(want) {
        t.Fatalf("got batch of size %d, want of size %d", len(batch.Stat), len(want))
    }

    for testId, got := range batch.Stat {
        if got != want[testId] {
            t.Errorf("for test %d got verdict %v, want %v", testId, got, want)
        }
    }
}
