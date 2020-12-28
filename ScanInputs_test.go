package cptest_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/kuredoro/cptest"
)

func TestScanTest(t *testing.T) {

	t.Run("don't trim space",
		func(t *testing.T) {
			want := cptest.Test{
				Input:  "\n \n  5\n  1 2 3 4 5\n   \n",
				Output: "\n  5 4 3 2 1\n\n  \n",
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
trash%and%trash
`

			inputText = strings.ReplaceAll(inputText, "%", cptest.IODelim)

			want := cptest.Test{
				Input:  inputText,
				Output: "correct\n",
			}

			text := fmt.Sprintf("%s%s\ncorrect", inputText, cptest.IODelim)

			test, errs := cptest.ScanTest(text)

			cptest.AssertTest(t, test, want)
			cptest.AssertNoErrors(t, errs)
		})

	t.Run("second+ IO delimeters are ignored",
		func(t *testing.T) {
			text := `a
---
b
---
c`

			want := cptest.Test{
				Input:  "a\n",
				Output: "b\n---\nc\n",
			}

			test, errs := cptest.ScanTest(text)

			cptest.AssertTest(t, test, want)
			cptest.AssertNoErrors(t, errs)
		})

	t.Run("only the prefix of a line should match IO delimeter",
		func(t *testing.T) {
			inputText := "3\r\n" +
				"---this text is discarded\r\n" +
				"random 12345\r\n"

			inputText = strings.ReplaceAll(inputText, "---", cptest.IODelim)

			want := cptest.Test{
				Input:  "3\n",
				Output: "random 12345\n",
			}

			test, errs := cptest.ScanTest(inputText)

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
				cptest.IOSeparatorMissing,
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

	t.Run("a lonely separator also counts",
		func(t *testing.T) {
			test, errs := cptest.ScanTest("---")

			cptest.AssertTest(t, test, cptest.Test{})
			cptest.AssertNoErrors(t, errs)
		})

	t.Run("empty input",
		func(t *testing.T) {
			test, errs := cptest.ScanTest("---\ntwo\n")

			want := cptest.Test{
				Input:  "",
				Output: "two\n",
			}

			cptest.AssertTest(t, test, want)
			cptest.AssertNoErrors(t, errs)
		})

	t.Run("empty output",
		func(t *testing.T) {
			test, errs := cptest.ScanTest("one\n---\n")

			want := cptest.Test{
				Input:  "one\n",
				Output: "",
			}

			cptest.AssertTest(t, test, want)
			cptest.AssertNoErrors(t, errs)
		})
}

func TestScanInputs(t *testing.T) {

	t.Run("emtpy returns empty inputs",
		func(t *testing.T) {
			inputs, errs := cptest.ScanInputs("")

			cptest.AssertTests(t, inputs.Tests, nil)
			cptest.AssertNoErrors(t, errs)
			cptest.AssertNoConfig(t, inputs.Config)
		})

	t.Run("single",
		func(t *testing.T) {

			testsWant := []cptest.Test{
				{
					Input:  "foo\n",
					Output: "bar\n",
				},
			}

			text := "foo\n---\nbar\n"
			text = strings.ReplaceAll(text, "---", cptest.IODelim)

			inputs, errs := cptest.ScanInputs(text)

			cptest.AssertTests(t, inputs.Tests, testsWant)
			cptest.AssertNoErrors(t, errs)
			cptest.AssertNoConfig(t, inputs.Config)
		})

	t.Run("multiple",
		func(t *testing.T) {

			testsWant := []cptest.Test{
				{
					Input:  "4\n1 2 3 4\n",
					Output: "4 3 2 1\n",
				},
				{
					Input:  "6\n1 2 3 4 5 6\n",
					Output: "6 5 4 3 2 1\n",
				},
				{
					Input:  "1\n1\n",
					Output: "1\n",
				},
			}

			text := `4
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

			inputs, errs := cptest.ScanInputs(text)

			cptest.AssertTests(t, inputs.Tests, testsWant)
			cptest.AssertNoErrors(t, errs)
			cptest.AssertNoConfig(t, inputs.Config)
		})

	t.Run("multiple with CRLF",
		func(t *testing.T) {

			testsWant := []cptest.Test{
				{
					Input:  "4\n1 2 3 4\n",
					Output: "4 3 2 1\n",
				},
				{
					Input:  "6\n1 2 3 4 5 6\n",
					Output: "6 5 4 3 2 1\n",
				},
				{
					Input:  "1\n1\n",
					Output: "1\n",
				},
			}

			text := "4\r\n" +
				"1 2 3 4\r\n" +
				"---\r\n" +
				"4 3 2 1\r\n" +
				"===\r\n" +
				"6\r\n" +
				"1 2 3 4 5 6\r\n" +
				"---\r\n" +
				"6 5 4 3 2 1\r\n" +
				"===\r\n" +
				"1\r\n" +
				"1\r\n" +
				"---\r\n" +
				"1\r\n"

			text = strings.ReplaceAll(text, "---", cptest.IODelim)
			text = strings.ReplaceAll(text, "===", cptest.TestDelim)

			inputs, errs := cptest.ScanInputs(text)

			cptest.AssertTests(t, inputs.Tests, testsWant)
			cptest.AssertNoErrors(t, errs)
			cptest.AssertNoConfig(t, inputs.Config)
		})

	t.Run("skip empty tests",
		func(t *testing.T) {

			testsWant := []cptest.Test{
				{
					Input:  "abc\n",
					Output: "cba\n",
				},
				{
					Input:  "xyz\n",
					Output: "zyx\n",
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

			inputs, errs := cptest.ScanInputs(text)

			cptest.AssertTests(t, inputs.Tests, testsWant)
			cptest.AssertNoErrors(t, errs)
			cptest.AssertNoConfig(t, inputs.Config)
		})

	t.Run("only line's prefix should match TestDelimeter",
		func(t *testing.T) {
			testsWant := []cptest.Test{
				{
					Input:  "",
					Output: "<===\n",
				},
				{
					Input:  "--===\n",
					Output: "",
				},
			}

			text := `
===>
---
<===
====
===
--===
---
===---
`
			text = strings.ReplaceAll(text, "---", cptest.IODelim)
			text = strings.ReplaceAll(text, "===", cptest.TestDelim)

			inputs, errs := cptest.ScanInputs(text)

			cptest.AssertTests(t, inputs.Tests, testsWant)
			cptest.AssertNoErrors(t, errs)
			cptest.AssertNoConfig(t, inputs.Config)
		})

	t.Run("configs may be listed before first test and once",
		func(t *testing.T) {
			testsWant := []cptest.Test{
				{
					Input:  "2 2\n",
					Output: "4\n",
				},
			}

			configWant := map[string]string{
				"tl":    "2.0",
				"foo":   "bar",
			}

			text := `
tl = 2.0
foo= bar
===
2 2
---
4
`
			text = strings.ReplaceAll(text, "---", cptest.IODelim)
			text = strings.ReplaceAll(text, "===", cptest.TestDelim)

			inputs, errs := cptest.ScanInputs(text)

			cptest.AssertTests(t, inputs.Tests, testsWant)
			cptest.AssertNoErrors(t, errs)
			cptest.AssertConfig(t, inputs.Config, configWant)
		})

	t.Run("configs are treated as such only before test 1",
		func(t *testing.T) {
			testsWant := []cptest.Test{
				{
					Input:  "2 2\n",
					Output: "4\n",
				},
			}

			text := `
===
2 2
---
4
===
tl = 2.0
foo= bar
        `
			text = strings.ReplaceAll(text, "---", cptest.IODelim)
			text = strings.ReplaceAll(text, "===", cptest.TestDelim)

			errsWant := []error{
				cptest.IOSeparatorMissing,
			}

			inputs, errs := cptest.ScanInputs(text)

			cptest.AssertTests(t, inputs.Tests, testsWant)
			cptest.AssertErrors(t, errs, errsWant)
			cptest.AssertNoConfig(t, inputs.Config)
		})

	t.Run("errors in config",
		func(t *testing.T) {
			text := `= foo
== aaa
a=b
===
extra=love
===`
			configWant := map[string]string{
				"a":    "b",
			}

            errLines := []int{1, 2}
			errsWant := []error{
                cptest.KeyMissing,
                cptest.KeyMissing,
                cptest.IOSeparatorMissing,
			}

			inputs, errs := cptest.ScanInputs(text)

			cptest.AssertTests(t, inputs.Tests, nil)
            cptest.AssertErrorLines(t, errs, errLines)
			cptest.AssertErrors(t, errs, errsWant)
			cptest.AssertConfig(t, inputs.Config, configWant)
		})

	t.Run("wierd (empty inputs)",
		func(t *testing.T) {
			text := `
===
---
===
---
===
---
===`

			inputs, errs := cptest.ScanInputs(text)

			cptest.AssertTests(t, inputs.Tests, nil)
			cptest.AssertNoErrors(t, errs)
			cptest.AssertNoConfig(t, inputs.Config)
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
				"hello":     "world",
				"foo":       "bar",
				"two words": "is   true",
			}

			cptest.AssertConfig(t, got, want)
			cptest.AssertNoErrors(t, errs)
		})

	t.Run("lines without assignments are keys without values",
		func(t *testing.T) {
			text := `
hi = owww
key assign value
zap = paz
ignore_newline
this is ok =
        `

			got, errs := cptest.ScanConfig(text)

			want := map[string]string{
				"hi":               "owww",
				"key assign value": "",
				"zap":              "paz",
				"ignore_newline":   "",
				"this is ok":       "",
			}

			cptest.AssertConfig(t, got, want)
			cptest.AssertNoErrors(t, errs)
		})

	t.Run("assignments with empty lhs are erroneous",
		func(t *testing.T) {
			text := `
foo=bar
foo=
=bar
=
 = 
        `

			got, errs := cptest.ScanConfig(text)

			want := map[string]string{
				"foo": "",
			}

			errLines := []int{4, 5, 6}
			errsWant := []error{
				cptest.KeyMissing,
				cptest.KeyMissing,
				cptest.KeyMissing,
			}

			cptest.AssertConfig(t, got, want)
			cptest.AssertErrorLines(t, errs, errLines)
			cptest.AssertErrors(t, errs, errsWant)
		})
}
