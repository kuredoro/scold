package cptest

import (
	"bufio"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
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

// Test represents a single test case: an input and the expected output.
type Test struct {
	Input  string
	Output string
}

// Duration is a wrapper around time.Duration that allows
// StringAttributesUnmarshal to parse it from a string using a common
// interface.
type Duration struct{ time.Duration }

// FromString will delegate parsing to built-in time.ParseDuration and, hence,
// accept the same format as time.ParseDuration.
func (d *Duration) FromString(str string) error {
	dur, err := time.ParseDuration(str)
	*d = Duration{dur}
	return err
}

// InputsConfig defines a schema for available configuration options that
// can be listed inside a config.
type InputsConfig struct {
	Prec uint8
	Tl   Duration
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

// ScanConfig tries to parse a stream of key-value pairs. Key-value pair is
// defined as `<string> "=" <string>`. Both strings are space trimmed. The key
// must be non-empty. Otherwise, a LinedError is produced. The final map
// will contain only correctly specified key-value pairs or an empty map.
// Duplicate keys are allowed, the later occurrence is preferred.
func ScanConfig(text string) (config map[string]string, key2line map[string]int, errs []error) {
	s := bufio.NewScanner(strings.NewReader(text))

	config = make(map[string]string)
	key2line = make(map[string]int)
	for lineNum := 1; s.Scan(); lineNum++ {
		key, val, err := ScanKeyValuePair(s.Text())

		if err != nil {
			errs = append(errs, &LineError{
				Line: lineNum,
				Err:  err,
			})
			continue
		}

		if key == "" && val == "" {
			continue
		}

		config[key] = val
		key2line[key] = lineNum
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
// one. If the first meaningful test could not be parsed without errors, it is
// interpreted as a configuration and parsed again. The empty tests are skipped
// (those that don't contain input, output and the separator). If test case
// could not be parsed, parsing continues to the next test case, but the errors
// are accumulated and returned together.
func ScanInputs(text string) (inputs Inputs, errs []error) {
	parts := SplitByInlinedPrefixN(text, TestDelim, 0)

	testNum := 0
	lineNum := 1
	for partNum, part := range parts {
		// Instead of incrementing lineNum at the code flow graph's each leaf,
		// we'll do it once at the beginning.
		if partNum != 0 {
			// A part is a string between two lines that start with a specified
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
				// We don't continue because ScanConfig gathers only correct
				// key-value pairs.
			}

			unmarshalErrs := StringMapUnmarshal(config, &inputs.Config)
			if unmarshalErrs != nil {
				merr := unmarshalErrs.(*multierror.Error)
				for i, err := range merr.Errors {
					merr.Errors[i] = &LineError{key2line[err.(*FieldError).FieldName], err}
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
            testLineCount := strings.Count(part, "\n")
			for i, err := range testErrs {
				testErrs[i] = &LineRangeError{lineNum, lineNum + testLineCount, &TestError{testNum, err}}
			}

			errs = append(errs, testErrs...)
			continue
		}

		inputs.Tests = append(inputs.Tests, test)
	}

	return
}
