package cptest_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/kureduro/cptest"
)

func TestScanInputs(t *testing.T) {

    t.Run("trim spaces",
    func(t *testing.T) {
        testsWant := []cptest.Test{
            {
                Input: "5\n1 2 3 4 5",
                Output: "5 4 3 2 1",
            },
        }

        text := `
 
  5
1 2 3 4 5
   
---

  5 4 3 2 1

  
        `

        inputs, errs := cptest.ScanInputs(strings.NewReader(text))

        cptest.AssertTests(t, inputs, testsWant)
        cptest.AssertNoErrors(t, errs)
    })

    t.Run("IO delimeter is alone on its own line",
    func(t *testing.T) {
        inputText := `3
abc%
%%
trash%and%trash`

        inputText = strings.ReplaceAll(inputText, "%", cptest.IODelim)

        testsWant := []cptest.Test{
            {
                Input: inputText,
                Output: "correct",
            },
        }

        text := fmt.Sprintf("%s\n%s\ncorrect", inputText, cptest.IODelim)

        inputs, errs := cptest.ScanInputs(strings.NewReader(text))

        cptest.AssertTests(t, inputs, testsWant)
        cptest.AssertNoErrors(t, errs)
    })

    t.Run("many IO delimeters",
    func(t *testing.T) {
        text := `
a
---
b
---
c
        `

        testsWant := []cptest.Test{
            {
                Input: "a",
                Output: "b\n---\nc",
            },
        }

        inputs, errs := cptest.ScanInputs(strings.NewReader(text))

        cptest.AssertTests(t, inputs, testsWant)
        cptest.AssertNoErrors(t, errs)
    })

    t.Run("no IO delimeters",
    func(t *testing.T) {
        text := `
abcd
dcba
        `

        errsWant := []error{
            cptest.NoSections,
        }

        inputs, errs := cptest.ScanInputs(strings.NewReader(text))

        cptest.AssertTests(t, inputs, nil)
        cptest.AssertErrors(t, errs, errsWant)
    })
}
