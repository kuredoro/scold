package scold_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/kuredoro/scold"
	"github.com/maxatome/go-testdeep/td"
)

func TestScanTest(t *testing.T) {

	t.Run("don't trim space",
		func(t *testing.T) {
			want := scold.Test{
				Input:  "\n \n  5\n  1 2 3 4 5\n   \n",
				Output: "\n  5 4 3 2 1\n\n  \n",
			}

			text := `
 
  5
  1 2 3 4 5
   
---

  5 4 3 2 1

  
`

			test, errs := scold.ScanTest(text)

			scold.AssertTest(t, test, want)
			scold.AssertNoErrors(t, errs)
		})

	t.Run("IO delimeters in wrong places are ignored",
		func(t *testing.T) {
			inputText := `3
abc%
trash%and%trash
`

			inputText = strings.ReplaceAll(inputText, "%", scold.IODelim)

			want := scold.Test{
				Input:  inputText,
				Output: "correct\n",
			}

			text := fmt.Sprintf("%s%s\ncorrect", inputText, scold.IODelim)

			test, errs := scold.ScanTest(text)

			scold.AssertTest(t, test, want)
			scold.AssertNoErrors(t, errs)
		})

	t.Run("second+ IO delimeters are ignored",
		func(t *testing.T) {
			text := `a
---
b
---
c`

			want := scold.Test{
				Input:  "a\n",
				Output: "b\n---\nc\n",
			}

			test, errs := scold.ScanTest(text)

			scold.AssertTest(t, test, want)
			scold.AssertNoErrors(t, errs)
		})

	t.Run("only the prefix of a line should match IO delimeter",
		func(t *testing.T) {
			inputText := "3\r\n" +
				"---this text is discarded\r\n" +
				"random 12345\r\n"

			inputText = strings.ReplaceAll(inputText, "---", scold.IODelim)

			want := scold.Test{
				Input:  "3\n",
				Output: "random 12345\n",
			}

			test, errs := scold.ScanTest(inputText)

			scold.AssertTest(t, test, want)
			scold.AssertNoErrors(t, errs)
		})

	t.Run("no IO delimeters result in error",
		func(t *testing.T) {
			text := `
abcd
dcba
        `

			errsWant := []error{
				scold.IOSeparatorMissing,
			}

			test, errs := scold.ScanTest(text)

			scold.AssertTest(t, test, scold.Test{})
			scold.AssertErrors(t, errs, errsWant)
		})

	t.Run("empty string returns empty test",
		func(t *testing.T) {
			test, errs := scold.ScanTest("")

			scold.AssertTest(t, test, scold.Test{})
			scold.AssertNoErrors(t, errs)
		})

	t.Run("a lonely separator also counts",
		func(t *testing.T) {
			test, errs := scold.ScanTest("---")

			scold.AssertTest(t, test, scold.Test{})
			scold.AssertNoErrors(t, errs)
		})

	t.Run("empty input",
		func(t *testing.T) {
			test, errs := scold.ScanTest("---\ntwo\n")

			want := scold.Test{
				Input:  "",
				Output: "two\n",
			}

			scold.AssertTest(t, test, want)
			scold.AssertNoErrors(t, errs)
		})

	t.Run("empty output",
		func(t *testing.T) {
			test, errs := scold.ScanTest("one\n---\n")

			want := scold.Test{
				Input:  "one\n",
				Output: "",
			}

			scold.AssertTest(t, test, want)
			scold.AssertNoErrors(t, errs)
		})
}

func TestScanInputs(t *testing.T) {

	t.Run("emtpy returns empty inputs",
		func(t *testing.T) {
			inputs, errs := scold.ScanInputs("")

			scold.AssertTests(t, inputs.Tests, nil)
			scold.AssertNoErrors(t, errs)
			scold.AssertDefaultConfig(t, inputs.Config)
		})

	t.Run("single",
		func(t *testing.T) {

			testsWant := []scold.Test{
				{
					Input:  "foo\n",
					Output: "bar\n",
				},
			}

			text := "foo\n---\nbar\n"
			text = strings.ReplaceAll(text, "---", scold.IODelim)

			inputs, errs := scold.ScanInputs(text)

			scold.AssertTests(t, inputs.Tests, testsWant)
			scold.AssertNoErrors(t, errs)
			scold.AssertDefaultConfig(t, inputs.Config)
		})

	t.Run("multiple",
		func(t *testing.T) {

			testsWant := []scold.Test{
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
			text = strings.ReplaceAll(text, "---", scold.IODelim)
			text = strings.ReplaceAll(text, "===", scold.TestDelim)

			inputs, errs := scold.ScanInputs(text)

			scold.AssertTests(t, inputs.Tests, testsWant)
			scold.AssertNoErrors(t, errs)
			scold.AssertDefaultConfig(t, inputs.Config)
		})

	t.Run("multiple with CRLF",
		func(t *testing.T) {

			testsWant := []scold.Test{
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

			text = strings.ReplaceAll(text, "---", scold.IODelim)
			text = strings.ReplaceAll(text, "===", scold.TestDelim)

			inputs, errs := scold.ScanInputs(text)

			scold.AssertTests(t, inputs.Tests, testsWant)
			scold.AssertNoErrors(t, errs)
			scold.AssertDefaultConfig(t, inputs.Config)
		})

	t.Run("skip empty tests",
		func(t *testing.T) {

			testsWant := []scold.Test{
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
			text = strings.ReplaceAll(text, "---", scold.IODelim)
			text = strings.ReplaceAll(text, "===", scold.TestDelim)

			inputs, errs := scold.ScanInputs(text)

			scold.AssertTests(t, inputs.Tests, testsWant)
			scold.AssertNoErrors(t, errs)
			scold.AssertDefaultConfig(t, inputs.Config)
		})

	t.Run("only line's prefix should match TestDelimeter",
		func(t *testing.T) {
			testsWant := []scold.Test{
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
			text = strings.ReplaceAll(text, "---", scold.IODelim)
			text = strings.ReplaceAll(text, "===", scold.TestDelim)

			inputs, errs := scold.ScanInputs(text)

			scold.AssertTests(t, inputs.Tests, testsWant)
			scold.AssertNoErrors(t, errs)
			scold.AssertDefaultConfig(t, inputs.Config)
		})

	t.Run("configs may be listed before first test and once",
		func(t *testing.T) {
			testsWant := []scold.Test{
				{
					Input:  "2 2\n",
					Output: "4\n",
				},
			}

			configWant := scold.InputsConfig{
				Tl:   scold.PositiveDuration{2 * time.Second},
				Prec: 16,
			}

			text := `
tl = 2s
prec= 16
===
2 2
---
4
`
			text = strings.ReplaceAll(text, "---", scold.IODelim)
			text = strings.ReplaceAll(text, "===", scold.TestDelim)

			inputs, errs := scold.ScanInputs(text)

			scold.AssertTests(t, inputs.Tests, testsWant)
			scold.AssertNoErrors(t, errs)
			td.Cmp(t, inputs.Config, configWant)
		})

	t.Run("not listed config keys shall be set to default",
		func(t *testing.T) {
			testsWant := []scold.Test{
				{
					Input:  "2 2\n",
					Output: "4\n",
				},
			}

			scold.DefaultInputsConfig = scold.InputsConfig{
				Tl:   scold.PositiveDuration{24 * time.Second},
				Prec: 6,
			}

			configWant := scold.InputsConfig{
				Tl:   scold.DefaultInputsConfig.Tl,
				Prec: 16,
			}

			text := `
prec =16
===
2 2
---
4
`
			text = strings.ReplaceAll(text, "---", scold.IODelim)
			text = strings.ReplaceAll(text, "===", scold.TestDelim)

			inputs, errs := scold.ScanInputs(text)

			scold.AssertTests(t, inputs.Tests, testsWant)
			scold.AssertNoErrors(t, errs)
			td.Cmp(t, inputs.Config, configWant)
		})

	t.Run("configs are treated as such only before test 1",
		func(t *testing.T) {
			testsWant := []scold.Test{
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
			text = strings.ReplaceAll(text, "---", scold.IODelim)
			text = strings.ReplaceAll(text, "===", scold.TestDelim)

			errsWant := []error{
				scold.IOSeparatorMissing,
			}

			inputs, errs := scold.ScanInputs(text)

			scold.AssertTests(t, inputs.Tests, testsWant)
			scold.AssertErrors(t, errs, errsWant)
			scold.AssertDefaultConfig(t, inputs.Config)
		})

	t.Run("errors in config",
		func(t *testing.T) {
			text := `= foo
foo= aaa
tl=10.a
===

===
extra=love
===
oh = and
by the way...
===`

			scold.DefaultInputsConfig = scold.InputsConfig{}

			errsWant := []error{
				&scold.LineRangeError{1, []string{"= foo"}, scold.KeyMissing},
				&scold.LineRangeError{2, []string{"foo= aaa"}, &scold.FieldError{"foo", scold.ErrUnknownField}},
				&scold.LineRangeError{3, []string{"tl=10.a"}, &scold.FieldError{"tl", &scold.NotValueOfTypeError{"PositiveDuration", "10.a", nil}}},
				&scold.LineRangeError{7, []string{"extra=love"}, &scold.TestError{1, scold.IOSeparatorMissing}},
				&scold.LineRangeError{9, []string{"oh = and", "by the way..."}, &scold.TestError{2, scold.IOSeparatorMissing}},
			}

			inputs, errs := scold.ScanInputs(text)

			scold.AssertTests(t, inputs.Tests, nil)
			td.Cmp(t, errs, td.Bag(td.Flatten(errsWant)))
			scold.AssertDefaultConfig(t, inputs.Config)
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

			inputs, errs := scold.ScanInputs(text)

			scold.AssertTests(t, inputs.Tests, nil)
			scold.AssertNoErrors(t, errs)
			scold.AssertDefaultConfig(t, inputs.Config)
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

			gotMap, gotLines, errs := scold.ScanConfig(text)

			wantMap := map[string]string{
				"hello":     "world",
				"foo":       "bar",
				"two words": "is   true",
			}

			wantLines := map[string]scold.NumberedLine{
				"hello":     {2, "hello = world"},
				"foo":       {3, "foo=bar"},
				"two words": {4, "  two words   =  is   true   "},
			}

			td.Cmp(t, gotMap, wantMap, "config contents")
			td.Cmp(t, gotLines, wantLines, "key to line mapping")
			scold.AssertNoErrors(t, errs)
		})

	t.Run("lines without assignments are keys without values",
		func(t *testing.T) {
			text := `hi = owww
key assign value
zap = paz
ignore_newline

this is ok =
        `

			gotMap, gotLines, errs := scold.ScanConfig(text)

			wantMap := map[string]string{
				"hi":               "owww",
				"key assign value": "",
				"zap":              "paz",
				"ignore_newline":   "",
				"this is ok":       "",
			}

			wantLines := map[string]scold.NumberedLine{
				"hi":               {1, "hi = owww"},
				"key assign value": {2, "key assign value"},
				"zap":              {3, "zap = paz"},
				"ignore_newline":   {4, "ignore_newline"},
				"this is ok":       {6, "this is ok ="},
			}

			td.Cmp(t, gotMap, wantMap, "config contents")
			td.Cmp(t, gotLines, wantLines, "key to line mapping")
			scold.AssertNoErrors(t, errs)
		})

	t.Run("assignments with empty lhs are erroneous",
		func(t *testing.T) {
			text := `
foo=bar
=bar
foo=
=
 = 
        `

			gotMap, gotLines, errs := scold.ScanConfig(text)

			wantMap := map[string]string{
				"foo": "",
			}

			wantLines := map[string]scold.NumberedLine{
				"foo": {4, "foo="},
			}

			errsWant := []error{
				&scold.LineRangeError{3, []string{"=bar"}, scold.KeyMissing},
				&scold.LineRangeError{5, []string{"="}, scold.KeyMissing},
				&scold.LineRangeError{6, []string{" = "}, scold.KeyMissing},
			}

			td.Cmp(t, gotMap, wantMap, "config contents")
			td.Cmp(t, gotLines, wantLines, "key to line mapping")
			td.Cmp(t, errs, td.Bag(td.Flatten(errsWant)))
		})
}
