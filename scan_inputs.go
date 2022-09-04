package scold

import (
	"bufio"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/stoewer/go-strcase"
)

// A set of errors that may be produced during scanning of the inputs file.
const (
	IOSeparatorMissing = StringError("IO separator missing")
	KeyMissing         = StringError("key cannot be empty")
)

// The set of delimeters used when partitioning inputs file.
const (
	IODelim   = "---"
	TestDelim = "==="
)

// DefaultInputsConfig is used to define default values for the InputsConfig
// inside Inputs. It is a starting ground. It may be altered further by
// customization points inside the inputs.txt file.
var DefaultInputsConfig InputsConfig

// Test represents a single test case: an input and the expected output.
type Test struct {
	Input  string
	Output string
}

// InputsConfig defines a schema for available configuration options that
// can be listed inside a config.
type InputsConfig struct {
	Prec uint8
	Tl   PositiveDuration
}

// Inputs contains all information located in the inputs file: tests and
// a valid configuration that were provided. Inputs is supposed to be copied
// around.
type Inputs struct {
	Tests  []Test
	Config InputsConfig
}

// ScanKeyValuePair parses the key-value pair of form 'key=value'.
// Strings without assignment are treated as keys with empty value.
// Strings with assignment but with empty key are erroneous.
// The space around key and value respectively is trimmed.
func ScanKeyValuePair(line string) (string, string, error) {
	parts := strings.SplitN(line, "=", 2)

	if len(parts) == 1 {
		cleanLine := strings.TrimSpace(line)

		if cleanLine == "" {
			return "", "", nil
		}

		return parts[0], "", nil
	}

	key := strings.TrimSpace(parts[0])
	val := strings.TrimSpace(parts[1])

	if key == "" {
		return "", "", KeyMissing
	}

	return key, val, nil
}

// NumberedLine is used to assign line number information to a string.
// It is used to map string map keys produced during config parsing to
// the relevant lines for error reporting.
type NumberedLine struct {
	Num  int
	Line string
}

// ScanConfig tries to parse a stream of key-value pairs. Key-value pair is
// defined as `<string> "=" <string>`. Both strings are space trimmed. The key
// must be non-empty. Otherwise, a LineRangeError is issued. Duplicate keys
// are allowed, the later occurrence is preferred.
//
// The function returns two maps: the first is a key-value map as defined in the
// supplied text, and the second maps keys to the relevant lines inside the
// config. The first one will contain only correctly defined keys. The second
// one is mainly used to correlate errors from StringMapUnmarshal to the
// inputs.txt lines and produce error messages.
func ScanConfig(text string) (config map[string]string, key2line map[string]NumberedLine, errs []error) {
	s := bufio.NewScanner(strings.NewReader(text))

	config = make(map[string]string)
	key2line = make(map[string]NumberedLine)
	for lineNum := 1; s.Scan(); lineNum++ {
		key, val, err := ScanKeyValuePair(s.Text())

		if err != nil {
			errs = append(errs, &LineRangeError{
				Begin: lineNum,
				Lines: []string{s.Text()},
				Err:   err,
			})
			continue
		}

		if key == "" && val == "" {
			continue
		}

		config[key] = val
		key2line[key] = NumberedLine{lineNum, s.Text()}
	}

	return
}

// SplitByInlinedPrefixN works in the same way as strings.SplitN. However,
// it does one additional thing. It matches the *prefixes of the lines*
// for equality with the delimeter. Upon match the entire line is discarded.
//
// If text doesn't contain the delimeter, only one part is returned.
// User can specify the number of parts they want at most via the third
// argument.
func SplitByInlinedPrefixN(text, delim string, n int) (parts []string) {

	var str strings.Builder

	s := bufio.NewScanner(strings.NewReader(text))
	for s.Scan() {

		if (n == 0 || len(parts)+1 < n) && strings.HasPrefix(s.Text(), delim) {
			part := str.String()
			parts = append(parts, part)

			str = strings.Builder{}
			continue
		}

		str.WriteString(s.Text())
		str.WriteRune('\n')
	}

	part := str.String()
	parts = append(parts, part)

	return
}

// ScanTest parses a single test case: input and output, separated with the
// Input/Output separator. If separator is absent, it returns an error.
func ScanTest(testStr string) (Test, []error) {
	if strings.TrimSpace(testStr) == "" {
		return Test{}, nil
	}

	parts := SplitByInlinedPrefixN(testStr, IODelim, 2)

	if len(parts) == 1 {
		return Test{}, []error{IOSeparatorMissing}
	}

	test := Test{
		Input:  parts[0],
		Output: parts[1],
	}

	return test, nil
}

// ScanInputs is the main routine for parsing inputs file. It splits the input
// by test case separator, and tries to parse each individual test case one by
// one. At the very beginning of the input file a configuration map can be
// specified (refer to ScanConfig for syntax), granted that it is not part of
// a meaninful test case. The empty tests are skipped (those that composed of
// non-printable characters and that don't contain IO separator). If a test case
// could not be parsed, parsing continues to the next test case, but the errors
// are accumulated and returned together.
func ScanInputs(text string) (inputs Inputs, errs []error) {
	inputs.Config = DefaultInputsConfig

	parts := SplitByInlinedPrefixN(text, TestDelim, 0)

	testNum := 0
	lineNum := 1
	for partNum, part := range parts {
		// Instead of incrementing lineNum at the code flow graph's each leaf,
		// we'll do it once at the beginning.
		if partNum != 0 {
			// A part is a string between two lines that starts with a specified
			// prefix. We need to add excluded line with prefix to keep line number
			// correct.
			lineNum += strings.Count(parts[partNum-1], "\n") + 1
		}

		test, testErrs := ScanTest(part)

		// Try to parse config
		if testErrs != nil && partNum == 0 {
			config, key2line, configErrs := ScanConfig(part)

			if configErrs != nil {
				errs = append(errs, configErrs...)
				// We don't stop because ScanConfig gathers only correct
				// key-value pairs.
			}

			unmarshalErrs := StringMapUnmarshal(config, &inputs.Config, strcase.UpperCamelCase)
			if unmarshalErrs != nil {
				merr := unmarshalErrs.(*multierror.Error)
				for i, err := range merr.Errors {
					numberedLine := key2line[err.(*FieldError).FieldName]
					merr.Errors[i] = &LineRangeError{
						Begin: numberedLine.Num,
						Lines: []string{numberedLine.Line},
						Err:   err,
					}
				}
				errs = append(errs, merr.Errors...)
			}

			continue
		}

		// Skip empty tests
		if testErrs == nil && test.Input == "" && test.Output == "" {
			continue
		}

		testNum++

		if testErrs != nil {
			// TODO: add scold.SplitAndTrim
			testLines := strings.Split(part, "\n")
			if len(testLines) != 0 && testLines[len(testLines)-1] == "" {
				testLines = testLines[:len(testLines)-1]
			}

			for i, err := range testErrs {
				testErrs[i] = &LineRangeError{
					Begin: lineNum,
					Lines: testLines,
					Err:   &TestError{testNum, err},
				}
			}

			errs = append(errs, testErrs...)
			continue
		}

		inputs.Tests = append(inputs.Tests, test)
	}

	return
}
