package cptest_test

import (
	"strings"
	"testing"

    "github.com/sanity-io/litter"

    "github.com/kureduro/cptest"
)

func TestScanInputs(t *testing.T) {
    text := `
    5
    1 2 3 4 5
    ---
    5 4 3 2 1
    `

    testsWant := []cptest.Test{
        {
            Input: "5\n1 2 3 4 5",
            Output: "5 4 3 2 1",
        },
    }

    inputs := cptest.ScanInputs(strings.NewReader(text))
    
    if len(inputs.Tests) != len(testsWant) {
        t.Errorf("\ngot %v\nwant %v", litter.Sdump(inputs.Tests), litter.Sdump(testsWant))
    }
}
