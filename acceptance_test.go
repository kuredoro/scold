package cptest_test

import (
    "testing"

    "github.com/kuredoro/cptest"
)

func TestAcceptanceSimple(t *testing.T) {
    inputsText := `
1 1
---
1
===
2 2
---
4
===
-2 2
---
4
    `

    inputs, err := cptest.ScanInputs(inputsText)

    cptest.AssertNoErrors(t, err)

    proc := cptest.ProcesserFunc(ProcFuncMultiply)

    swatch := cptest.NewConfigurableStopwatcher(0)

    batch := cptest.NewTestingBatch(inputs, proc, swatch)

    batch.Run()

    want := map[int]cptest.Verdict{
        1: cptest.OK,
        2: cptest.OK,
        3: cptest.WA,
    }

    cptest.AssertVerdicts(t, batch.Verdicts, want)
}
