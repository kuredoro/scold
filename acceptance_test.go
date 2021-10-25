package scold_test

/*
import (
	"testing"

	"github.com/kuredoro/scold"
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

	inputs, err := scold.ScanInputs(inputsText)

	scold.AssertNoErrors(t, err)

	proc := scold.ProcesserFunc(ProcFuncMultiply)
	swatch := scold.NewConfigurableStopwatcher(0)
    pool := scold.NewThreadPool(3)

	batch := scold.NewTestingBatch(inputs, proc, swatch, pool)
	batch.Run()

	want := map[int]scold.Verdict{
		1: scold.OK,
		2: scold.OK,
		3: scold.WA,
	}

	scold.AssertVerdicts(t, batch.Verdicts, want)
}
*/
