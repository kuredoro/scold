package cptest_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/kureduro/cptest"
)

func TestScanTest(t *testing.T) {

    t.Run("trim spaces",
    func(t *testing.T) {
        want := cptest.Test{
            Input: "5\n1 2 3 4 5",
            Output: "5 4 3 2 1",
        }

        text := `
 
  5
1 2 3 4 5
   
---

  5 4 3 2 1

  
        `

        test, errs := cptest.ScanTest(text)

        cptest.AssertTest(t, test, want)
        cptest.AssertNoErrors(t, errs)
    })

    t.Run("IO delimeters in wrong places are ignored",
    func(t *testing.T) {
        inputText := `3
abc%
%%
trash%and%trash`

        inputText = strings.ReplaceAll(inputText, "%", cptest.IODelim)

        want := cptest.Test{
            Input: inputText,
            Output: "correct",
        }

        text := fmt.Sprintf("%s\n%s\ncorrect", inputText, cptest.IODelim)

        test, errs := cptest.ScanTest(text)

        cptest.AssertTest(t, test, want)
        cptest.AssertNoErrors(t, errs)
    })

    t.Run("second+ IO delimeters are ignored",
    func(t *testing.T) {
        text := `
a
---
b
---
c
        `

        want := cptest.Test{
            Input: "a",
            Output: "b\n---\nc",
        }

        test, errs := cptest.ScanTest(text)

        cptest.AssertTest(t, test, want)
        cptest.AssertNoErrors(t, errs)
    })

    t.Run("no IO delimeters result in error",
    func(t *testing.T) {
        text := `
abcd
dcba
        `

        errsWant := []error{
            cptest.NoSections,
        }

        test, errs := cptest.ScanTest(text)

        cptest.AssertTest(t, test, cptest.Test{})
        cptest.AssertErrors(t, errs, errsWant)
    })

    t.Run("empty string returns empty test",
    func(t *testing.T) {
        test, errs := cptest.ScanTest("")

        cptest.AssertTest(t, test, cptest.Test{})
        cptest.AssertNoErrors(t, errs)
    })
}

func TestScanInputs(t *testing.T) {

    t.Run("single",
    func(t *testing.T) {

        testsWant := []cptest.Test{
            {
                Input: "foo",
                Output: "bar",
            },
        }

        text := "foo\n---\nbar\n"
        text = strings.ReplaceAll(text, "---", cptest.IODelim)

        inputs, errs := cptest.ScanInputs(strings.NewReader(text))
        
        cptest.AssertTests(t, inputs.Tests, testsWant)
        cptest.AssertNoErrors(t, errs)
    })

    t.Run("multiple",
    func(t *testing.T) {

        testsWant := []cptest.Test{
            {
                Input: "4\n1 2 3 4",
                Output: "4 3 2 1",
            },
            {
                Input: "6\n1 2 3 4 5 6",
                Output: "6 5 4 3 2 1",
            },
            {
                Input: "1\n1",
                Output: "1",
            },
        }

        text := `
4
1 2 3 4
---
4 3 2 1
===
6
1 2 3 4 5 6
---
6 5 4 3 2 1
===
1
1
---
1
        `
        text = strings.ReplaceAll(text, "---", cptest.IODelim)
        text = strings.ReplaceAll(text, "===", cptest.TestDelim)

        inputs, errs := cptest.ScanInputs(strings.NewReader(text))
        
        cptest.AssertTests(t, inputs.Tests, testsWant)
        cptest.AssertNoErrors(t, errs)
    })

    t.Run("skip empty tests",
    func(t *testing.T) {

        testsWant := []cptest.Test{
            {
                Input: "abc",
                Output: "cba",
            },
            {
                Input: "xyz",
                Output: "zyx",
            },
        }

        text := `
===
===
abc
---
cba
===
===
===
xyz
---
zyx
===
        `
        text = strings.ReplaceAll(text, "---", cptest.IODelim)
        text = strings.ReplaceAll(text, "===", cptest.TestDelim)

        inputs, errs := cptest.ScanInputs(strings.NewReader(text))

        cptest.AssertTests(t, inputs.Tests, testsWant)
        cptest.AssertNoErrors(t, errs)
    })

    t.Run("TestDelimeters in wrong places",
    func(t *testing.T) {
        testsWant := []cptest.Test{
            {
                Input: "===>",
                Output: "<===\n====",
            },
            {
                Input: "=== \ntra^iling space",
                Output: "",
            },
        }

        text := `
===>
---
<===
====
===
=== 
tra^iling space
---
        `
        text = strings.ReplaceAll(text, "---", cptest.IODelim)
        text = strings.ReplaceAll(text, "===", cptest.TestDelim)

        inputs, errs := cptest.ScanInputs(strings.NewReader(text))

        cptest.AssertTests(t, inputs.Tests, testsWant)
        cptest.AssertNoErrors(t, errs)
    })

    t.Run("emtpy returns empty inputs",
    func(t *testing.T) {
        inputs, errs := cptest.ScanInputs(strings.NewReader(""))

        cptest.AssertTests(t, inputs.Tests, nil)
        cptest.AssertNoErrors(t, errs)
    })
}

func TestScanConfig(t *testing.T) {

    t.Run("trim spaces",
    func(t *testing.T) {
        text := `
hello = world
foo=bar
  two words   =  is   true   
        `

        got, errs := cptest.ScanConfig(text)

        want := map[string]string{
            "hello": "world",
            "foo": "bar",
            "two words": "is   true",
        }

        cptest.AssertConfig(t, got, want)
        cptest.AssertNoErrors(t, errs)
    })

    t.Run("Lines without assignments are gibberish",
    func(t *testing.T) {
        text := `
hi = owww
key assign value
zap = paz
uoenahonetuhneo
        `

        got, errs := cptest.ScanConfig(text)

        want := map[string]string{
            "hi": "owww",
            "zap": "paz",
        }

        errLines := []int{3, 5}

        cptest.AssertConfig(t, got, want)
        cptest.AssertErrorLines(t, errs, errLines)
    })
}
